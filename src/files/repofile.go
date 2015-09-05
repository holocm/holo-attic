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
	"fmt"
	"path/filepath"
	"strings"

	"../common"
)

//This type represents a single file in the configuration repository. The
//string stored in it is the path to the repo file (also accessible as Path()).
type RepoFile string

func NewRepoFile(path string) RepoFile {
	return RepoFile(path)
}

func (file RepoFile) Path() string {
	return string(file)
}

func (file RepoFile) ConfigFile() ConfigFile {
	//the optional ".holoscript" suffix appears only on repo files
	repoFile := file.Path()
	if strings.HasSuffix(repoFile, ".holoscript") {
		repoFile = strings.TrimSuffix(repoFile, ".holoscript")
	}

	//make path relative
	relPath, _ := filepath.Rel(common.RepoDirectory(), repoFile)
	//remove the disambiguation path element to get to the relPath for the ConfigFile
	//e.g. repoFile = '/holo/repo/23-foo/etc/foo.conf'
	//  -> relPath  = '23-foo/etc/foo.conf'
	//  -> relPath  = 'etc/foo.conf'
	segments := strings.SplitN(relPath, fmt.Sprintf("%c", filepath.Separator), 2)
	relPath = segments[1]

	return ConfigFile(relPath)
}

func (file RepoFile) ApplicationStrategy() string {
	if strings.HasSuffix(file.Path(), ".holoscript") {
		return "passthru"
	} else {
		return "apply"
	}
}

//This type holds a slice of RepoFile instances, and implements some methods
//to satisfy the sort.Interface interface.
type RepoFiles []RepoFile

func (f RepoFiles) Len() int           { return len(f) }
func (f RepoFiles) Less(i, j int) bool { return f[i] < f[j] }
func (f RepoFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
