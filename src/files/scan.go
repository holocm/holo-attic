/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package files

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"../common"
)

//ScanRepo returns a slice of all the ConfigFiles which have accompanying RepoFiles,
//and also a string slice of all orphaned target bases (target bases without a
//ConfigFile).
func ScanRepo() (configFiles ConfigFiles, orphanedTargetBases []string) {
	//check that the repo and target base directories exist
	repoPath := common.RepoDirectory()
	targetBasePath := common.TargetBaseDirectory()
	pathsThatMustExist := []string{repoPath, targetBasePath}

	for _, path := range pathsThatMustExist {
		fi, err := os.Lstat(path)
		if err != nil {
			common.PrintError("Cannot open %s: %s", path, err.Error())
			return nil, nil
		}
		if !fi.IsDir() {
			common.PrintError("Cannot open %s: not a directory!", path)
			return nil, nil
		}
	}

	//cannot declare this as "var result ConfigFiles" because then we would
	//return nil if there are no entity definitions, but nil indicates an error
	result := ConfigFiles{}
	seen := make(map[string]bool) //used to avoid duplicates in result, and also to find orphaned target bases

	//walk over the repo to find repo files (and thus the corresponding target files)
	filepath.Walk(repoPath, func(repoFile string, repoFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(repoFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(repoFileInfo)) {
			return nil
		}
		//only look at files within subdirectories (files in the repo directory
		//itself are skipped)
		relPath, _ := filepath.Rel(repoPath, repoFile)
		if !strings.ContainsRune(relPath, filepath.Separator) {
			return nil
		}

		//check if we had this config file already
		configFile := NewRepoFile(repoFile).ConfigFile()
		if !seen[configFile.TargetPath()] {
			result = append(result, configFile)
			seen[configFile.TargetPath()] = true
		}
		return nil
	})

	//walk over the target base directory to find orphaned target bases
	var targetBaseOrphans []string
	filepath.Walk(targetBasePath, func(targetBaseFile string, targetBaseFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(targetBaseFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(targetBaseFileInfo)) {
			return nil
		}

		//check if we have seen the config file for this target base
		configFile := NewConfigFileFromTargetBasePath(targetBaseFile)
		if !seen[configFile.TargetPath()] {
			targetBaseOrphans = append(targetBaseOrphans, targetBaseFile)
		}
		return nil
	})

	sort.Sort(result)
	sort.Strings(targetBaseOrphans)
	return result, targetBaseOrphans
}
