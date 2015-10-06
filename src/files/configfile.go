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

	"../common"
)

//ConfigFile represents a single target file, and includes methods to calculate
//the corresponding target base and repo file(s). The string stored in it
//is the path of the target file relative to the target directory.
//
//For example, if the target file is "/etc/pacman.conf", the string stored is
//"etc/pacman.conf".
type ConfigFile string

//NewConfigFileFromTargetBasePath creates a ConfigFile instance for which the path
//to the target base is known.
func NewConfigFileFromTargetBasePath(targetBaseFile string) ConfigFile {
	//make path relative
	relPath, _ := filepath.Rel(common.TargetBaseDirectory(), targetBaseFile)
	return ConfigFile(relPath)
}

//NewConfigFileFromTargetPath creates a ConfigFile instance for which the path
//to the target file is known.
func NewConfigFileFromTargetPath(targetFile string) ConfigFile {
	//make path relative
	relPath, _ := filepath.Rel(common.TargetDirectory(), targetFile)
	return ConfigFile(relPath)
}

//TargetPath returns the location where this config file is installed.
func (file ConfigFile) TargetPath() string {
	//make path absolute
	return filepath.Join(common.TargetDirectory(), string(file))
}

//TargetBasePath returns the location where the target base for this config
//file is stored.
func (file ConfigFile) TargetBasePath() string {
	//make path absolute
	return filepath.Join(common.TargetBaseDirectory(), string(file))
}

//ProvisionedPath returns the location where a duplicate of the last provisioned
//content for this config file is stored.
func (file ConfigFile) ProvisionedPath() string {
	//make path absolute
	return filepath.Join(common.ProvisionedDirectory(), string(file))
}

//RepoFiles returns all repo files that belong to this ConfigFile.
func (file ConfigFile) RepoFiles() RepoFiles {
	var result RepoFiles

	//check every subdirectory of the RepoDirectory() if it contains a repo file for this ConfigFile
	dirNames := repoSubDirectories()
	for _, dirName := range dirNames {
		//build absolute path
		repoPath := filepath.Join(common.RepoDirectory(), dirName, string(file))

		//check if the repo file exists
		if common.IsManageableFile(repoPath) {
			result = append(result, NewRepoFile(repoPath))
		} else {
			//it may have an optional ".holoscript" suffix
			repoPath2 := repoPath + ".holoscript"
			if common.IsManageableFile(repoPath2) {
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
	dir, err := os.Open(common.RepoDirectory())
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

//ConfigFiles holds a slice of ConfigFile instances, and implements some methods
//to satisfy the sort.Interface interface.
type ConfigFiles []ConfigFile

func (f ConfigFiles) Len() int           { return len(f) }
func (f ConfigFiles) Less(i, j int) bool { return f[i] < f[j] }
func (f ConfigFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
