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
	"path/filepath"
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

	//materialize FS entries in the root directory
	err = pkg.materializeFSEntries(rootPath)
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

func (pkg *Package) materializeFSEntries(rootPath string) error {
	var additionalSetupScript string

	for _, entry := range pkg.FSEntries {
		//find the path within the rootPath for this entry
		path, _ := filepath.Rel("/", entry.Path)
		path = filepath.Join(rootPath, path)

		//mkdir -p $(dirname $entry_path)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		//write entry
		switch entry.Type {
		case FSEntryTypeRegular:
			err = ioutil.WriteFile(path, []byte(entry.Content), entry.Mode)
		case FSEntryTypeDirectory:
			err = os.Mkdir(path, entry.Mode)
		case FSEntryTypeSymlink:
			err = os.Symlink(entry.Content, path)
		}
		if err != nil {
			return err
		}

		//ownership only applies to files and directories
		if entry.Type == FSEntryTypeSymlink {
			continue
		}

		//apply ownership (numeric ownership can be written into the package directly; ownership by name will be applied in the setupScript)
		var uid, gid uint32 = 0, 0
		if entry.Owner != nil {
			if entry.Owner.Str == "" {
				uid = entry.Owner.Int
			} else {
				additionalSetupScript += fmt.Sprintf("chown %s %s", entry.Owner.Str, entry.Path)
			}
		}
		if entry.Group != nil {
			if entry.Group.Str == "" {
				gid = entry.Group.Int
			} else {
				additionalSetupScript += fmt.Sprintf("chgrp %s %s\n", entry.Group.Str, entry.Path)
			}
		}
		if uid != 0 || gid != 0 {
			err = os.Chown(path, int(uid), int(gid))
			if err != nil {
				return err
			}
		}
	}

	if additionalSetupScript != "" {
		//ensure "\n" at end of existing setupScript
		if pkg.SetupScript != "" {
			pkg.SetupScript = strings.TrimSpace(pkg.SetupScript) + "\n"
		}
		pkg.SetupScript += additionalSetupScript
	}

	return nil
}
