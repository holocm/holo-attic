/*******************************************************************************
*
*   Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
*   This program is free software; you can redistribute it and/or modify it
*   under the terms of the GNU General Public License as published by the Free
*   Software Foundation; either version 2 of the License, or (at your option)
*   any later version.
*
*   This program is distributed in the hope that it will be useful, but WITHOUT
*   ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
*   FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
*   more details.
*
*   You should have received a copy of the GNU General Public License along
*   with this program; if not, write to the Free Software Foundation, Inc.,
*   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.
*
********************************************************************************/

package files

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"../common"
)

//ScanRepo returns a slice of all the ConfigFiles which have accompanying RepoFiles,
//and also a string slice of all orphaned backup files (backup files without a
//ConfigFile).
func ScanRepo() (configFiles ConfigFiles, orphanedBackupFiles []string) {
	//check that the repo and backup directories exist
	repoPath := common.RepoDirectory()
	backupPath := common.BackupDirectory()
	pathsThatMustExist := []string{repoPath, backupPath}

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
	seen := make(map[string]bool) //used to avoid duplicates in result, and also to find orphaned backup files

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

	//walk over the backup directory to find orphaned backup files
	var backupOrphans []string
	filepath.Walk(backupPath, func(backupFile string, backupFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(backupFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(backupFileInfo)) {
			return nil
		}

		//check if we have seen the config file for this backup file
		configFile := NewConfigFileFromBackupPath(backupFile)
		if !seen[configFile.TargetPath()] {
			backupOrphans = append(backupOrphans, backupFile)
		}
		return nil
	})

	sort.Sort(result)
	sort.Strings(backupOrphans)
	return result, backupOrphans
}
