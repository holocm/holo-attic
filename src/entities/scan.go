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
	"sort"
	"strings"

	"../common"
)

//Scan returns a slice of all the defined entities. If an error is encountered
//during the scan, it will be reported on stdout, and nil is returned.
func Scan() Entities {
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

	//cannot declare this as "var result []Definition" because then we would
	//return nil if there are no entity definitions, but nil indicates an error
	result := Entities{}
	success := true

	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".json") {
			defPath := filepath.Join(entityPath, fi.Name())
			entities, err := readDefinitionFile(defPath)
			if err != nil {
				common.PrintError("Cannot read %s: %s", defPath, err.Error())
				success = false //don't return nil immediately; report all broken files
			} else {
				result = append(result, entities...)
			}
		}
	}

	if success {
		sort.Sort(result)
		return result
	}
	return nil
}

func readDefinitionFile(entityFile string) (Entities, error) {
	file, err := os.Open(entityFile)
	if err != nil {
		return nil, err
	}

	//json.Unmarshal can only write into *exported* (i.e. upper-case) struct
	//fields, but the fields on the Group struct are private to emphasize their
	//readonly-ness, so we have to jump through some hoops to read these
	var contents struct {
		Groups []struct {
			Name   string
			Gid    int
			System bool
		}
	}
	err = json.NewDecoder(file).Decode(&contents)
	if err != nil {
		return nil, err
	}

	var result Entities
	for _, group := range contents.Groups {
		result = append(result, Group{
			name:           group.Name,
			gid:            group.Gid,
			system:         group.System,
			definitionFile: entityFile,
		})
	}

	return result, nil
}
