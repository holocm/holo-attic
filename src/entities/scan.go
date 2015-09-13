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

package entities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"../common"
)

//This returns a slice of all the found EntityDefinition. If an error is encountered
//during the scan, it will be reported on stdout, and nil is returned.
func Scan() []Definition {
	//look in the entity directory for entity definitions
	entityPath := common.EntityDirectory()
	dir, err := os.Open(entityPath)
	if err != nil {
		common.PrintError("Cannot read %s: %s", entityPath, err.Error())
		return nil
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		common.PrintError("Cannot read %s: %s", entityPath, err.Error())
		return nil
	}

	var result []Definition
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".yaml") {
			defPath := filepath.Join(entityPath, fi.Name())
			def, err := readDefinitionFile(defPath)
			switch {
			case def != nil:
				result = append(result, *def)
			case err != nil:
				common.PrintError("Cannot read %s: %s", defPath, err.Error())
				return nil
			default:
				return nil
			}
		}
	}

	return result
}

func readDefinitionFile(entityFile string) (*Definition, error) {
	file, err := os.Open(entityFile)
	if err != nil {
		return nil, err
	}

	var contents struct {
		groups []Group
	}
	err = json.NewDecoder(file).Decode(&contents)

	return &Definition{
		File:   entityFile,
		Groups: contents.groups,
	}, nil
}
