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

//Package platform implements integration points with the platform that Holo is
//running (most notably the package manager).
package platform

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"../common"
)

//Impl provides integration points with a distribution's toolchain.
type Impl interface {
	//FindUpdatedTargetBase is called as part of the repo file application
	//algorithm. If the system package manager updates a file which has been
	//modified by Holo, it will usually place the new stock configuration next
	//to the targetPath (usually with a special suffix). If such a file exists,
	//this method must return its name, so that Holo can pick it up and use it
	//as a new base configuration.
	//
	//The reportedPath is usually the same as the actualPath, but some
	//implementations have to move files around, in which case the reportedPath
	//is the original path to the updated target base, and the actualPath is
	//where Holo will find the file.
	FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error)
	//AdditionalCleanupTargets is called as part of the orphan handling. When
	//an application package is removed, but one of its configuration files has
	//been modified by Holo, the system package manager will usually retain a
	//copy next to the targetPath (usually with a special suffix). If such a
	//file exists, this method must return its name, for Holo to clean it up.
	AdditionalCleanupTargets(targetPath string) []string
}

var impl Impl

//Implementation returns the most suitable platform implementation for the
//current system.
func Implementation() Impl {
	return impl
}

func init() {
	//get list of distribution IDs
	dists := getCurrentDistribution()
	//convert into hash for easier lookup
	isDist := make(map[string]bool)
	for _, dist := range dists {
		isDist[dist] = true
	}

	//which distribution are we running on?
	switch {
	case isDist["arch"]:
		impl = archImpl{}
	case isDist["debian"]:
		impl = dpkgImpl{}
	case isDist["fedora"], isDist["suse"]:
		impl = rpmImpl{}
	case isDist["unittest"]:
		//set via HOLO_CURRENT_DISTRIBUTION=unittest only
		impl = genericImpl{}
	default:
		common.PrintError("Running on an unrecognized distribution. Distribution IDs: %s", strings.Join(dists, ","))
		common.PrintWarning("Please report this error at <https://github.com/holocm/holo/issues/new>")
		common.PrintWarning("and include the contents of your /etc/os-release file.")
		impl = genericImpl{}
	}
}

//Returns a list of distribution IDs, drawing on the ID= and ID_LIKE= fields of
//os-release(5).
func getCurrentDistribution() []string {
	//check if a unit test override is active
	if value := os.Getenv("HOLO_CURRENT_DISTRIBUTION"); value != "" {
		return []string{value}
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
	result := []string{variables["ID"]}
	if idLike, ok := variables["ID_LIKE"]; ok {
		result = append(result, strings.Split(idLike, " ")...)
	}
	return result
}
