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

func ScanRepo() ConfigFiles {
	//check that the repo directory exists
	repoPath := RepoDirectory()
	repoInfo, err := os.Lstat(repoPath)
	if err != nil {
		PrintError("Cannot open %s: %s", repoPath, err.Error())
		return nil
	}
	if !repoInfo.IsDir() {
		PrintError("Cannot open %s: not a directory!", repoPath)
		return nil
	}

	var result ConfigFiles

	//walk over the repo to find repo files (and thus the corresponding target files)
	filepath.Walk(repoPath, func(repoFile string, repoFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(repoFileInfo.Mode().IsRegular() || IsFileInfoASymbolicLink(repoFileInfo)) {
			return nil
		}

		result = append(result, NewRepoFile(repoFile).ConfigFile())
		return nil
	})

	sort.Sort(result)
	return result
}
