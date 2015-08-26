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

import "path/filepath"

//This type represents a single target file, and includes methods to calculate
//the corresponding backup location and repo file(s). The string stored in it
//is the path of the target file relative to the target directory.
//
//For example, if the target file is "/etc/pacman.conf", the string stored is
//"etc/pacman.conf".
type ConfigFile string

func (file ConfigFile) TargetPath() string {
	//make path absolute
	return filepath.Join(TargetDirectory(), string(file))
}

func (file ConfigFile) BackupPath() string {
	//make path absolute
	return filepath.Join(BackupDirectory(), string(file))
}

func (file ConfigFile) RepoFile() RepoFile {
	//make path absolute
	repoPath := filepath.Join(RepoDirectory(), string(file))

	//the repo file may have an optional ".holoscript" suffix
	if repoPath2 := repoPath + ".holoscript"; IsManageableFile(repoPath2) {
		return NewRepoFile(repoPath2)
	} else {
		return NewRepoFile(repoPath)
	}
}

//This type holds a slice of ConfigFile instances, and implements some methods
//to satisfy the sort.Interface interface.
type ConfigFiles []ConfigFile

func (f ConfigFiles) Len() int           { return len(f) }
func (f ConfigFiles) Less(i, j int) bool { return f[i] < f[j] }
func (f ConfigFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
