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

package entities

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"../common"
)

//User represents a UNIX user account (as registered in /etc/passwd). It
//implements the Entity interface and is handled accordingly.
type User struct {
	name            string   //the user name (the first field in /etc/passwd)
	comment         string   //the full name (sometimes also called "comment"; the fifth field in /etc/passwd)
	uid             int      //the user ID (the third field in /etc/passwd), or 0 if no specific UID is enforced
	system          bool     //whether the group is a system group (this influences the GID selection if gid = 0)
	homeDirectory   string   //path to the user's home directory (or empty to use the default)
	group           string   //the name of the user's initial login group (or empty to use the default)
	groups          []string //the names of supplementary groups which the user is also a member of
	shell           string   //path to the user's login shell (or empty to use the default)
	definitionFiles []string //path to the file defining this entity

	broken bool //whether the entity definition is invalid (default: false)
}

//isValid is used inside the scanning algorithm to filter entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) isValid() bool { return !u.broken }

//setInvalid is used inside the scnaning algorithm to mark entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) setInvalid() { u.broken = true }

//EntityID implements the Entity interface for User.
func (u User) EntityID() string { return "user:" + u.name }

//Report implements the Entity interface for User.
func (u User) Report() *common.Report {
	r := common.Report{Target: u.EntityID()}
	for _, defFile := range u.definitionFiles {
		r.AddLine("found in", defFile)
	}
	if attributes := u.attributes(); attributes != "" {
		r.AddLine("with", attributes)
	}
	return &r
}

func (u User) attributes() string {
	attrs := []string{}
	if u.system {
		attrs = append(attrs, "type: system")
	}
	if u.uid > 0 {
		attrs = append(attrs, fmt.Sprintf("uid: %d", u.uid))
	}
	if u.homeDirectory != "" {
		attrs = append(attrs, "home: "+u.homeDirectory)
	}
	if u.group != "" {
		attrs = append(attrs, "login group: "+u.group)
	}
	if len(u.groups) > 0 {
		attrs = append(attrs, "groups: "+strings.Join(u.groups, ","))
	}
	if u.shell != "" {
		attrs = append(attrs, "login shell: "+u.shell)
	}
	if u.comment != "" {
		attrs = append(attrs, "comment: "+u.comment)
	}
	return strings.Join(attrs, ", ")
}

//Apply performs the complete application algorithm for the given Entity.
//If the group does not exist yet, it is created. If it does exist, but some
//attributes do not match, it will be updated, but only if withForce is given.
func (u User) Apply(withForce bool) {
	r := u.Report()
	r.Action = "Working on"
	u.doApply(r, withForce)
	r.Print()
}

type userDiff struct {
	field    string
	actual   string
	expected string
}

func (u User) doApply(report *common.Report, withForce bool) {
	//check if we have that group already
	userExists, actualUser, err := u.checkExists()
	if err != nil {
		report.AddError("Cannot read user database: %s", err.Error())
		return
	}

	//check if the actual properties diverge from our definition
	if userExists {
		differences := []userDiff{}
		if u.comment != "" && u.comment != actualUser.comment {
			differences = append(differences, userDiff{"comment", actualUser.comment, u.comment})
		}
		if u.uid > 0 && u.uid != actualUser.uid {
			differences = append(differences, userDiff{"UID", strconv.Itoa(actualUser.uid), strconv.Itoa(u.uid)})
		}
		if u.homeDirectory != "" && u.homeDirectory != actualUser.homeDirectory {
			differences = append(differences, userDiff{"home directory", actualUser.homeDirectory, u.homeDirectory})
		}
		if u.shell != "" && u.shell != actualUser.shell {
			differences = append(differences, userDiff{"login shell", actualUser.shell, u.shell})
		}
		if u.group != "" && u.group != actualUser.group {
			differences = append(differences, userDiff{"login group", actualUser.group, u.group})
		}
		//to detect changes in u.groups <-> actualUser.groups, we sort and join both slices
		expectedGroupsSlice := append([]string(nil), u.groups...) //take a copy of the slice
		sort.Strings(expectedGroupsSlice)
		expectedGroups := strings.Join(expectedGroupsSlice, ", ")
		actualGroupsSlice := append([]string(nil), actualUser.groups...)
		sort.Strings(actualGroupsSlice)
		actualGroups := strings.Join(actualGroupsSlice, ", ")
		if expectedGroups != actualGroups {
			differences = append(differences, userDiff{"groups", actualGroups, expectedGroups})
		}

		if len(differences) != 0 {
			if withForce {
				for _, diff := range differences {
					report.AddLine("fix", fmt.Sprintf("%s (was: %s)", diff.field, diff.actual))
				}
				err := u.callUsermod(report)
				if err != nil {
					report.AddError(err.Error())
					return
				}
			} else {
				for _, diff := range differences {
					report.AddWarning("User has %s: %s, expected %s (use --force to overwrite)", diff.field, diff.actual, diff.expected)
				}
			}
		}
	} else {
		//create the user if it does not exist
		err := u.callUseradd(report)
		if err != nil {
			report.AddError(err.Error())
			return
		}
	}
}

//checkExists checks if the user exists in /etc/passwd. If it does, its actual
//properties will be returned in the second return argument.
func (u User) checkExists() (exists bool, currentUser *User, e error) {
	passwdFile := filepath.Join(common.TargetDirectory(), "etc/passwd")
	groupFile := filepath.Join(common.TargetDirectory(), "etc/group")

	//fetch entry from /etc/passwd
	fields, err := common.Getent(passwdFile, func(fields []string) bool { return fields[0] == u.name })
	if err != nil {
		return false, nil, err
	}
	//is there such a user?
	if fields == nil {
		return false, nil, nil
	}
	//is the passwd entry intact?
	if len(fields) < 4 {
		return true, nil, errors.New("invalid entry in /etc/passwd (not enough fields)")
	}

	//read fields in passwd entry
	actualUID, err := strconv.Atoi(fields[2])
	if err != nil {
		return true, nil, err
	}

	//fetch entry for login group from /etc/group (to resolve actualGID into a
	//group name)
	actualGIDString := fields[3]
	groupFields, err := common.Getent(groupFile, func(fields []string) bool {
		if len(fields) <= 2 {
			return false
		}
		return fields[2] == actualGIDString
	})
	if err != nil {
		return true, nil, err
	}
	if groupFields == nil {
		return true, nil, errors.New("invalid entry in /etc/passwd (login group does not exist)")
	}
	groupName := groupFields[0]

	//check /etc/group for the supplementary group memberships of this user
	groupNames := []string{}
	_, err = common.Getent(groupFile, func(fields []string) bool {
		if len(fields) <= 3 {
			return false
		}
		//collect groups that contain this user
		users := strings.Split(fields[3], ",")
		for _, user := range users {
			if user == u.name {
				groupNames = append(groupNames, fields[0])
			}
		}
		//keep going
		return false
	})
	if err != nil {
		return true, nil, err
	}

	return true, &User{
		//NOTE: Some fields (name, system, definitionFile) are not set because
		//they are not relevant for the algorithm.
		comment:       fields[4],
		uid:           actualUID,
		homeDirectory: fields[5],
		group:         groupName,
		groups:        groupNames,
		shell:         fields[6],
	}, nil
}

func (u User) callUseradd(report *common.Report) error {
	//assemble arguments for useradd call
	args := []string{}
	if u.system {
		args = append(args, "--system")
	}
	if u.uid > 0 {
		args = append(args, "--uid", strconv.Itoa(u.uid))
	}
	if u.comment != "" {
		args = append(args, "--comment", u.comment)
	}
	if u.homeDirectory != "" {
		args = append(args, "--home-dir", u.homeDirectory)
	}
	if u.group != "" {
		args = append(args, "--gid", u.group)
	}
	if len(u.groups) > 0 {
		args = append(args, "--groups", strings.Join(u.groups, ","))
	}
	if u.shell != "" {
		args = append(args, "--shell", u.shell)
	}
	args = append(args, u.name)

	//call useradd
	_, err := common.ExecProgramOrMock(report, []byte{}, "useradd", args...)
	return err
}

func (u User) callUsermod(report *common.Report) error {
	//assemble arguments for usermod call
	args := []string{}
	if u.uid > 0 {
		args = append(args, "--uid", strconv.Itoa(u.uid))
	}
	if u.comment != "" {
		args = append(args, "--comment", u.comment)
	}
	if u.homeDirectory != "" {
		args = append(args, "--home", u.homeDirectory)
	}
	if u.group != "" {
		args = append(args, "--gid", u.group)
	}
	if len(u.groups) > 0 {
		args = append(args, "--groups", strings.Join(u.groups, ","))
	}
	if u.shell != "" {
		args = append(args, "--shell", u.shell)
	}
	args = append(args, u.name)

	//call usermod
	_, err := common.ExecProgramOrMock(report, []byte{}, "usermod", args...)
	return err
}
