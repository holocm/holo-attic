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
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"../../internal/toml"
)

//PackageDefinition only needs a nice exported name for the TOML parser to
//produce more meaningful error messages on malformed input data.
type PackageDefinition struct {
	Package PackageSection
}

//PackageSection only needs a nice exported name for the TOML parser to produce
//more meaningful error messages on malformed input data.
type PackageSection struct {
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

//ParsePackageDefinition parses a package definition from the given input.
//The operation is successful if the returned []error is nil or empty.
func ParsePackageDefinition(input io.Reader) (*Package, []error) {
	//read from input
	blob, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, []error{err}
	}
	var p PackageDefinition
	_, err = toml.Decode(string(blob), &p)
	if err != nil {
		return nil, []error{err}
	}

	//restructure the parsed data into a common.Package struct
	pkg := Package{
		Name:          strings.TrimSpace(p.Package.Name),
		Version:       strings.TrimSpace(p.Package.Version),
		Description:   strings.TrimSpace(p.Package.Description),
		SetupScript:   strings.TrimSpace(p.Package.SetupScript),
		CleanupScript: strings.TrimSpace(p.Package.CleanupScript),
	}

	//do some basic validation on the package name and version since we're
	//going to use these to construct a path
	var errors []error
	if pkg.Name == "" {
		errors = append(errors, fmt.Errorf("Missing package name", pkg.Name))
	}
	if strings.ContainsAny(pkg.Name, "/\r\n") {
		errors = append(errors, fmt.Errorf("Invalid package name \"%s\" (may not contain slashes or newlines)", pkg.Name))
	}
	if pkg.Version == "" {
		errors = append(errors, fmt.Errorf("Missing package version", pkg.Name))
	}
	if strings.ContainsAny(pkg.Version, "/\r\n") {
		errors = append(errors, fmt.Errorf("Invalid package version \"%s\" (may not contain slashes or newlines)", pkg.Version))
	}
	if strings.ContainsAny(pkg.Description, "\r\n") {
		errors = append(errors, fmt.Errorf("Invalid package description \"%s\" (may not contain newlines)", pkg.Name))
	}

	//parse relations to other packages
	var errs []error
	pkg.Requires, errs = parseRelatedPackages(p.Package.Requires)
	errors = append(errors, errs...)
	pkg.Provides, errs = parseRelatedPackages(p.Package.Provides)
	errors = append(errors, errs...)
	pkg.Conflicts, errs = parseRelatedPackages(p.Package.Conflicts)
	errors = append(errors, errs...)
	pkg.Replaces, errs = parseRelatedPackages(p.Package.Replaces)
	errors = append(errors, errs...)

	return &pkg, errors
}

var relatedPackageRx = regexp.MustCompile(`^([^\s<=>]+)\s*(?:(<=?|>=?|=)\s*(\S+))?$`)

func parseRelatedPackages(specs []string) ([]PackageRelation, []error) {
	rels := make([]PackageRelation, 0, len(specs))
	idxByName := make(map[string]int, len(specs))
	var errors []error

	for _, spec := range specs {
		//check format of spec
		match := relatedPackageRx.FindStringSubmatch(spec)
		if match == nil {
			errors = append(errors, fmt.Errorf("Invalid package reference: \"%s\"", spec))
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

	return rels, errors
}
