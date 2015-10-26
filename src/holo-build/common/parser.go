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
	"regexp"
	"strings"

	"../../internal/toml"
	"../../shared"
)

//ParsePackageDefinition parses a package definition from the given input.
func ParsePackageDefinition(input io.Reader, r *shared.Report) (result *Package, hasError bool) {
	//prepare a data structure matching the input format
	var p struct {
		Package struct {
			Name          string
			Version       string
			Description   string
			Requires      []string
			Provides      []string
			Conflicts     []string
			Replaces      []string
			SetupScript   string
			CleanupScript string
		}
	}

	//read from input
	blob, err := ioutil.ReadAll(input)
	if err != nil {
		r.AddError(err.Error())
		return nil, true
	}
	_, err = toml.Decode(string(blob), &p)
	if err != nil {
		r.AddError(err.Error())
		return nil, true
	}

	//restructure the parsed data into a common.Package struct
	pkg := Package{
		Name:          p.Package.Name,
		Version:       p.Package.Version,
		Description:   p.Package.Description,
		SetupScript:   p.Package.SetupScript,
		CleanupScript: p.Package.CleanupScript,
	}
	hasError = false

	//do some basic validation on the package name and version since we're
	//going to use these to construct a path
	if strings.ContainsAny(pkg.Name, "/\r\n") {
		r.AddError("Invalid package name \"%s\" (may not contain slashes or newlines)", pkg.Name)
		hasError = true
	}
	if strings.ContainsAny(pkg.Version, "/\r\n") {
		r.AddError("Invalid package version \"%s\" (may not contain slashes or newlines)", pkg.Version)
		hasError = true
	}
	if strings.ContainsAny(pkg.Description, "\r\n") {
		r.AddError("Invalid package description \"%s\" (may not contain newlines)", pkg.Name)
		hasError = true
	}

	//parse relations to other packages
	hasErr := false
	pkg.Requires, hasErr = parseRelatedPackages(p.Package.Requires, r)
	hasError = hasError || hasErr
	pkg.Provides, hasErr = parseRelatedPackages(p.Package.Provides, r)
	hasError = hasError || hasErr
	pkg.Conflicts, hasErr = parseRelatedPackages(p.Package.Conflicts, r)
	hasError = hasError || hasErr
	pkg.Replaces, hasErr = parseRelatedPackages(p.Package.Replaces, r)
	hasError = hasError || hasErr

	return &pkg, hasError
}

var relatedPackageRx = regexp.MustCompile(`^([^\s<=>]+)\s*(?:(<=?|>=?|=)\s*(\S+))?$`)

func parseRelatedPackages(specs []string, r *shared.Report) (result []PackageRelation, hasError bool) {
	rels := make([]PackageRelation, 0, len(specs))
	idxByName := make(map[string]int, len(specs))
	hasErr := false

	for _, spec := range specs {
		//check format of spec
		match := relatedPackageRx.FindStringSubmatch(spec)
		if match == nil {
			r.AddError("Invalid package reference: \"%s\"", spec)
			hasErr = true
			continue
		}

		//do we have a relation to this package already?
		name := match[1]
		idx, exists := idxByName[name]
		if !exists {
			//no, add a new one and remember it for later additional constraints
			idx = len(rels)
			idxByName[name] = idx
			rels = append(rels, PackageRelation{RelatedPackage: name})
		}

		//add version constraint if one was specified
		if match[2] != "" {
			constraint := VersionConstraint{Relation: match[2], Version: match[3]}
			rels[idx].Constraints = append(rels[idx].Constraints, constraint)
		}
	}

	return rels, hasErr
}
