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

import "fmt"

//Note: This line is parsed by the Makefile to get the version string. If you
//change the format, adjust the Makefile too.
var version = "v0.8-pre"
var codename = "Enthusiasm"

//VersionString returns a string including the version and the release codename.
func VersionString() string {
	return fmt.Sprintf("%s \"%s\"", version, codename)
}
