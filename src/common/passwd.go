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
	"io/ioutil"
	"strings"
)

//Getent reads entries from a UNIX user/group database (e.g. /etc/passwd
//or /etc/group) and returns the first entry matching the given predicate.
//For example, to locate the user with name "foo":
//
//    fields, err := Getent("/etc/passwd", func(fields []string) bool {
//        return fields[0] == "foo"
//    })
func Getent(databaseFile string, predicate func([]string) bool) ([]string, error) {
	//read database file
	contents, err := ioutil.ReadFile(databaseFile)
	if err != nil {
		return nil, err
	}

	//each entry is one line
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	for _, line := range lines {
		//fields inside the entries are separated by colons
		fields := strings.Split(strings.TrimSpace(line), ":")
		if predicate(fields) {
			return fields, nil
		}
	}

	//no entry matches
	return nil, nil
}
