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

package common

import (
	"io/ioutil"
	"strings"
)

//Getent returns an entry from a UNIX user/group database (e.g. /etc/passwd
//or /etc/group) for the given key (usually the name of the user or group in
//question). If no entry exists, an empty string is returned.
func Getent(databaseFile string, key string) (string, error) {
	//read database file
	contents, err := ioutil.ReadFile(databaseFile)
	if err != nil {
		return "", err
	}

	//the line that we're looking for has the key as the first field, and keys
	//are separated by colons
	prefix := key + ":"
	//find the line with the given key
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return line, nil
		}
	}

	//there is no group with that name
	return "", nil
}
