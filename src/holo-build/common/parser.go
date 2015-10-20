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
	"io"
	"io/ioutil"

	"../../internal/toml"
	"../../shared"
)

//ParsePackageDefinition parses a package definition from the given input.
func ParsePackageDefinition(input io.Reader, r *shared.Report) (p *Package, hasError bool) {
	//prepare a data structure matching the input format
	var pkg struct {
		Package struct {
			Name    string
			Version string
		}
	}

	//read from input
	blob, err := ioutil.ReadAll(input)
	if err != nil {
		r.AddError(err.Error())
		return nil, true
	}
	_, err = toml.Decode(string(blob), &pkg)
	if err != nil {
		r.AddError(err.Error())
		return nil, true
	}

	//restructure the parsed data into a common.Package struct
	return &Package{
		Name:    pkg.Package.Name,
		Version: pkg.Package.Version,
	}, false
}
