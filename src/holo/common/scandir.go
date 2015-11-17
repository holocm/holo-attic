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

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

//ScanDirectory reads the directory at path and returns the full paths of all
//files in it that match the given predicate. No recursive walking of
//subdirectories is performed. The result slice is sorted by file name.
func ScanDirectory(path string, predicate func(fi os.FileInfo) bool) ([]string, error) {
	//open directory for reading
	dir, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read %s: %s", path, err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("Cannot read %s: %s", path, err.Error())
	}

	//find matching entries
	var result []string
	for _, fi := range fis {
		if predicate(fi) {
			result = append(result, filepath.Join(path, fi.Name()))
		}
	}
	sort.Strings(result)

	return result, nil
}
