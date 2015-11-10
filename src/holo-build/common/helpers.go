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
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

//FindApparentSizeForPath runs `du -s --apparent-size` on the given path to
//find the apparent size of this directory and everything below it.
//
//The value returned is in bytes.
func FindApparentSizeForPath(path string) (int, error) {
	cmd := exec.Command("du", "-s", "-B", "1", "--apparent-size", path)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	//output is size in bytes + "\t" + path
	match := regexp.MustCompile(`^([0-9]+)\s`).FindSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("invalid output returned from `du -s -B 1 --apparent-size %s`: \"%s\"", path, string(output))
	}
	return strconv.Atoi(string(match[1]))
}
