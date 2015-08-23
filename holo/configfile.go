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
	"path/filepath"
	"strings"
)

//This type represents a single target file, and includes methods to calculate
//the corresponding backup location and repo file(s). The string stored in it
//is the path of the target file relative to the target directory.
//
//For example, if the target file is "/etc/pacman.conf", the string stored is
//"etc/pacman.conf".
type ConfigFile string

func NewConfigFileFromRepoPath(repoFile string) ConfigFile {
	//the optional ".holoscript" suffix appears only on repo files
	if strings.HasSuffix(repoFile, ".holoscript") {
		repoFile = strings.TrimSuffix(repoFile, ".holoscript")
	}
	//make path relative
	relPath, _ := filepath.Rel(RepoDirectory(), repoFile)
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

func (file ConfigFile) RepoPath() string {
	//make path absolute
	repoPath := filepath.Join(RepoDirectory(), string(file))

	//the repo file may have an optional ".holoscript" suffix
	if repoPath2 := repoPath + ".holoscript"; IsManageableFile(repoPath2) {
		return repoPath2
	} else {
		return repoPath
	}
}
