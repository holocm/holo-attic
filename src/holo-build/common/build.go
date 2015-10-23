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
	"io/ioutil"
	"os"
	"strings"
)

//Build builds the package using the given Generator.
func (pkg *Package) Build(generator Generator, printToStdout bool) error {
	//choose root directory in such a way that the user can easily find and
	//inspect it in the case that an error occurs
	rootPath := fmt.Sprintf("./holo-build-%s-%s", pkg.Name, pkg.Version)

	//if the root directory exists from a previous run, remove it recursively
	err := os.RemoveAll(rootPath)
	if err != nil {
		return err
	}

	//create the root directory
	err = os.MkdirAll(rootPath, 0755)
	if err != nil {
		return err
	}

	//build package
	pkgBytes, err := generator.Build(pkg, rootPath)
	if err != nil {
		return err
	}

	//if requested, cleanup the target directory
	err = os.RemoveAll(rootPath)
	if err != nil {
		return err
	}

	//write package, either to stdout or to the working directory
	if printToStdout {
		_, err := os.Stdout.Write(pkgBytes)
		if err != nil {
			return err
		}
	} else {
		pkgFile := generator.RecommendedFileName(pkg)
		if strings.ContainsAny(pkgFile, "/ \t\r\n") {
			return fmt.Errorf("Unexpected filename generated: \"%s\"", pkgFile)
		}
		err := ioutil.WriteFile(pkgFile, pkgBytes, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}
