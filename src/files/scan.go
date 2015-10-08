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

//ScanRepo returns a slice of all the TargetFile entities.
func ScanRepo() common.Entities {
	//check that the repo and target base directories exist
	repoPath := common.RepoDirectory()
	targetBasePath := common.TargetBaseDirectory()
	pathsThatMustExist := []string{repoPath, targetBasePath}

	for _, path := range pathsThatMustExist {
		fi, err := os.Lstat(path)
		if err != nil {
			common.PrintError("Cannot open %s: %s", path, err.Error())
			return nil
		}
		if !fi.IsDir() {
			common.PrintError("Cannot open %s: not a directory!", path)
			return nil
		}
	}

	//walk over the repo to find repo files (and thus the corresponding target files)
	targets := make(map[string]*TargetFile)
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

		//create new TargetFile if necessary and store the repo entry in it
		repoEntry := NewRepoFile(repoFile)
		targetPath := repoEntry.TargetPath()
		if targets[targetPath] == nil {
			targets[targetPath] = NewTargetFileFromPathIn(common.TargetDirectory(), targetPath)
		}
		targets[targetPath].AddRepoEntry(repoEntry)
		return nil
	})

	//walk over the target base directory to find orphaned target bases
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
		//(if not, it's orphaned)
		//TODO: s/(targetBase)Path/\1Dir/g and s/(targetBase)File/Path/g
		target := NewTargetFileFromPathIn(targetBasePath, targetBaseFile)
		targetPath := target.PathIn(common.TargetDirectory())
		if targets[targetPath] == nil {
			target.orphaned = true
			targets[targetPath] = target
		}
		return nil
	})

	//flatten result into list
	result := make(common.Entities, 0, len(targets))
	for _, target := range targets {
		result = append(result, target)
	}

	sort.Sort(result)
	return result
}
