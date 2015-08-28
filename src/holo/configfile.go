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

package holo

import (
	"os"
	"path/filepath"
	"sort"
)

//This type represents a single target file, and includes methods to calculate
//the corresponding backup location and repo file(s). The string stored in it
//is the path of the target file relative to the target directory.
//
//For example, if the target file is "/etc/pacman.conf", the string stored is
//"etc/pacman.conf".
type ConfigFile string

func NewConfigFileFromBackupPath(backupFile string) ConfigFile {
	//make path relative
	relPath, _ := filepath.Rel(BackupDirectory(), backupFile)
	return ConfigFile(relPath)
}

func (file ConfigFile) TargetPath() string {
	//make path absolute
	return filepath.Join(TargetDirectory(), string(file))
}

func (file ConfigFile) BackupPath() string {
	//make path absolute
	return filepath.Join(BackupDirectory(), string(file))
}

func (file ConfigFile) RepoFiles() RepoFiles {
	var result RepoFiles

	//check every subdirectory of the RepoDirectory() if it contains a repo file for this ConfigFile
	dirNames := repoSubDirectories()
	for _, dirName := range dirNames {
		//build absolute path
		repoPath := filepath.Join(RepoDirectory(), dirName, string(file))

		//check if the repo file exists
		if IsManageableFile(repoPath) {
			result = append(result, NewRepoFile(repoPath))
		} else {
			//it may have an optional ".holoscript" suffix
			repoPath2 := repoPath + ".holoscript"
			if IsManageableFile(repoPath2) {
				result = append(result, NewRepoFile(repoPath2))
			}
		}
	}

	sort.Sort(result)
	return result
}

func repoSubDirectories() []string {
	//NOTE: Any IO errors in here are silently ignored. If any subdirectory is
	//not accessible, we just ignore it.
	dir, err := os.Open(RepoDirectory())
	if err != nil {
		return []string{}
	}

	//read all the directory entries in the repo directory
	fis, _ := dir.Readdir(-1)
	var result []string
	for _, fi := range fis {
		if fi.IsDir() {
			result = append(result, fi.Name())
		}
	}
	return result
}

//This type holds a slice of ConfigFile instances, and implements some methods
//to satisfy the sort.Interface interface.
type ConfigFiles []ConfigFile

func (f ConfigFiles) Len() int           { return len(f) }
func (f ConfigFiles) Less(i, j int) bool { return f[i] < f[j] }
func (f ConfigFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
