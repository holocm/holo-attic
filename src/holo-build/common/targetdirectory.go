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
)

//MakeTargetDirectory allocates a temporary directory for building this package
//and places all the files and directories contained within the package in this
//directory, with the directory itself corresponding to the root directory of
//the packaged filesystem. The return value is the path of this directory.
func (pkg *Package) MakeTargetDirectory() (string, error) {
	//choose target directory in such a way that the user can easily find and
	//inspect it in the case that an error occurs
	path := fmt.Sprintf("./pkg/%s-%s", pkg.Name, pkg.Version)

	//if the directory exists from a previous run, remove it recursively
	err := os.RemoveAll(path)
	if err != nil {
		return path, err
	}

	//create the directory
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return path, err
	}

	return path, nil
}
