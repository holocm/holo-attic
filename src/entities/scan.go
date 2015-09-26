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
	"fmt"
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

	//collect all definition files, sort by name
	var paths []string
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".json") {
			paths = append(paths, filepath.Join(entityPath, fi.Name()))
		}
	}
	sort.Strings(paths)

	//parse entity definitions
	groups := make(map[string]*Group)
	users := make(map[string]*User)
	success := true
	for _, path := range paths {
		err := readDefinitionFile(path, &groups, &users)
		if len(err) > 0 {
			success = false //don't return nil immediately; report all broken files
			if len(err) == 1 {
				common.PrintError("Failed to read %s: %s", path, err[0].Error())
			} else {
				common.PrintError("Failed to read %s because of multiple errors:", path)
				for _, suberr := range err {
					common.PrintError("    %s", suberr.Error())
				}
			}
		}
	}

	if !success {
		return nil
	}

	//flatten reuslt into a list sorted by EntityID
	entities := make(Entities, 0, len(groups)+len(users))
	for _, group := range groups {
		entities = append(entities, group)
	}
	for _, user := range users {
		entities = append(entities, user)
	}
	sort.Sort(entities)
	return entities
}

//json.Unmarshal can only write into *exported* (i.e. upper-case) struct
//fields, but the fields on the Group/User structs are private to emphasize
//their readonly-ness, so we need separate struct definitions here
type groupDefinition struct {
	Name   string
	Gid    int
	System bool
}
type userDefinition struct {
	Name    string
	Comment string
	UID     int
	System  bool
	Home    string
	Group   string
	Groups  []string
	Shell   string
}

func readDefinitionFile(entityFile string, groups *map[string]*Group, users *map[string]*User) []error {
	file, err := os.Open(entityFile)
	if err != nil {
		return []error{err}
	}

	//unmarshal JSON
	//json.Unmarshal can only write into *exported* (i.e. upper-case) struct
	//fields, but the fields on the Group/User structs are private to emphasize
	//their readonly-ness, so we have to jump through some hoops to read these
	var contents struct {
		Groups []groupDefinition
		Users  []userDefinition
	}
	err = json.NewDecoder(file).Decode(&contents)
	if err != nil {
		return []error{err}
	}

	//when checking the entity definitions, report all errors at once
	var errors []error

	//convert the definitions read into entities, or extend existing entities if
	//the definition is stacked on an earlier one (BUT: we only allow changes
	//that are compatible with the original definition; for example, users may
	//be extended with additional groups, but its UID may not be changed)
	for idx, groupDef := range contents.Groups {
		if groupDef.Name == "" {
			errors = append(errors, fmt.Errorf("groups[%d] is missing required 'name' attribute", idx))
			continue
		}
		group, exists := (*groups)[groupDef.Name]
		if exists {
			//stacked definition for this group - extend existing Group entity
			errors = append(errors, mergeGroupDefinition(groupDef, group)...)
			group.definitionFiles = append(group.definitionFiles, entityFile)
		} else {
			//first definition for this group - create new Group entity
			(*groups)[groupDef.Name] = &Group{
				name:            groupDef.Name,
				gid:             groupDef.Gid,
				system:          groupDef.System,
				definitionFiles: []string{entityFile},
			}
		}
	}

	for idx, userDef := range contents.Users {
		if userDef.Name == "" {
			errors = append(errors, fmt.Errorf("users[%d] is missing required 'name' attribute", idx))
			continue
		}
		user, exists := (*users)[userDef.Name]
		if exists {
			//stacked definition for this user - extend existing User entity
			errors = append(errors, mergeUserDefinition(userDef, user)...)
			user.definitionFiles = append(user.definitionFiles, entityFile)
		} else {
			//first definition for this user - create new User entity
			(*users)[userDef.Name] = &User{
				name:            userDef.Name,
				comment:         userDef.Comment,
				uid:             userDef.UID,
				system:          userDef.System,
				homeDirectory:   userDef.Home,
				group:           userDef.Group,
				groups:          userDef.Groups,
				shell:           userDef.Shell,
				definitionFiles: []string{entityFile},
			}
		}
	}

	return errors
}

//Merges `def` into `group` if possible, returns errors if merge conflicts arise.
func mergeGroupDefinition(def groupDefinition, group *Group) []error {
	var errors []error

	//GID can be set by `def` if `group` does not have a different value set
	if def.Gid != 0 {
		switch {
		case group.gid == 0:
			group.gid = def.Gid
		case def.Gid != 0 && group.gid != def.Gid:
			errors = append(errors, fmt.Errorf(
				"conflicting GID for group '%s' (existing: %d, new: %d)",
				group.name, group.gid, def.Gid,
			))
		}
	}

	//the system flag can be set by `def` if `group` did not set it yet
	group.system = group.system || def.System

	return errors
}

//Merges `def` into `user` if possible, returns errors if merge conflicts arise.
func mergeUserDefinition(def userDefinition, user *User) []error {
	var errors []error

	//comment is assumed to be informational only, the last definition always
	//takes precedence
	if def.Comment != "" {
		user.comment = def.Comment
	}

	//UID can be set by `def` if `user` does not have a different value set
	if def.UID != 0 {
		switch {
		case user.uid == 0:
			user.uid = def.UID
		case def.UID != 0 && user.uid != def.UID:
			errors = append(errors, fmt.Errorf(
				"conflicting UID for user '%s' (existing: %d, new: %d)",
				user.name, user.uid, def.UID,
			))
		}
	}

	//the system flag can be set by `def` if `user` did not set it yet
	user.system = user.system || def.System

	//homeDirectory may be set only once
	if def.Home != "" {
		switch {
		case user.homeDirectory == "":
			user.homeDirectory = def.Home
		case def.Home != "" && user.homeDirectory != def.Home:
			errors = append(errors, fmt.Errorf(
				"conflicting home directory for user '%s' (existing: %s, new: %s)",
				user.name, user.homeDirectory, def.Home,
			))
		}
	}

	//group may be set only once
	if def.Group != "" {
		switch {
		case user.group == "":
			user.group = def.Group
		case def.Group != "" && user.group != def.Group:
			errors = append(errors, fmt.Errorf(
				"conflicting login group for user '%s' (existing: %s, new: %s)",
				user.name, user.group, def.Group,
			))
		}
	}

	//shell may be set only once
	if def.Shell != "" {
		switch {
		case user.shell == "":
			user.shell = def.Shell
		case def.Shell != "" && user.shell != def.Shell:
			errors = append(errors, fmt.Errorf(
				"conflicting login shell for user '%s' (existing: %s, new: %s)",
				user.name, user.shell, def.Shell,
			))
		}
	}

	//auxiliary groups can always be added
	for _, group := range def.Groups {
		//append group to user.groups, but avoid duplicates
		missing := true
		for _, other := range user.groups {
			if other == group {
				missing = false
				break
			}
		}
		if !missing {
			user.groups = append(user.groups, group)
		}
	}

	return errors
}
