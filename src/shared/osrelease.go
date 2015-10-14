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

package shared

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

//GetCurrentDistribution returns a set of distribution IDs, drawing on the ID=
//and ID_LIKE= fields of os-release(5).
func GetCurrentDistribution() map[string]bool {
	//check if a unit test override is active
	if value := os.Getenv("HOLO_CURRENT_DISTRIBUTION"); value != "" {
		return map[string]bool{value: true}
	}

	//read /etc/os-release, fall back to /usr/lib/os-release if not available
	bytes, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		if os.IsNotExist(err) {
			bytes, err = ioutil.ReadFile("/usr/lib/os-release")
		}
	}
	if err != nil {
		panic("Cannot read os-release: " + err.Error())
	}

	//parse os-release syntax (a harshly limited subset of shell script)
	variables := make(map[string]string)
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		//ignore comments
		if line == "" || line[0] == '#' {
			continue
		}
		//line format is key=value
		if !strings.Contains(line, "=") {
			continue
		}
		split := strings.SplitN(line, "=", 2)
		key, value := split[0], split[1]
		//value may be enclosed in quotes
		switch {
		case strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""):
			value = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
		case strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'"):
			value = strings.TrimPrefix(strings.TrimSuffix(value, "'"), "'")
		}
		//special characters may be escaped
		value = regexp.MustCompile(`\\(.)`).ReplaceAllString(value, "$1")
		//store assignment
		variables[key] = value
	}

	//the distribution IDs we're looking for are in ID= (single value) or ID_LIKE= (space-separated list)
	result := map[string]bool{variables["ID"]: true}
	if idLike, ok := variables["ID_LIKE"]; ok {
		ids := strings.Split(idLike, " ")
		for _, id := range ids {
			result[id] = true
		}
	}
	return result
}
