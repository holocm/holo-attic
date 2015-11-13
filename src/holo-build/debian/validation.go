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

package debian

import (
	"errors"
	"fmt"
	"regexp"

	"../common"
)

//reference: https://www.debian.org/doc/debian-policy/ch-controlfields.html
var packageNameRx = regexp.MustCompile(`^[a-z0-9][a-z0-9+-.]+$`)
var packageVersionRx = regexp.MustCompile(`^[0-9][A-Za-z0-9.+:~-]*$`)

//Validate implements the common.Generator interface.
func (g *Generator) Validate(pkg *common.Package) []error {
	//TODO: refactor to show all errors at once
	err := validatePackage(pkg)
	if err != nil {
		return []error{err}
	}
	return nil
}

func validatePackage(pkg *common.Package) error {
	if !packageNameRx.MatchString(pkg.Name) {
		return fmt.Errorf("Package name \"%s\" is not acceptable for Debian packages", pkg.Name)
	}
	if !packageVersionRx.MatchString(pkg.Version) {
		//this check is only some Defense in Depth; a stricted version format
		//is already enforced by the generator-independent validation
		return fmt.Errorf("Package version \"%s\" is not acceptable for Debian packages", pkg.Version)
	}
	if pkg.Author == "" {
		return errors.New("The \"package.author\" field is required for Debian packages")
	}

	err := validatePackageRelations("requires", pkg.Requires)
	if err != nil {
		return err
	}
	err = validatePackageRelations("provides", pkg.Provides)
	if err != nil {
		return err
	}
	err = validatePackageRelations("conflicts", pkg.Conflicts)
	if err != nil {
		return err
	}
	err = validatePackageRelations("replaces", pkg.Replaces)
	if err != nil {
		return err
	}

	return nil
}

func validatePackageRelations(relType string, rels []common.PackageRelation) error {
	for _, rel := range rels {
		if !packageNameRx.MatchString(rel.RelatedPackage) {
			return fmt.Errorf("Package name \"%s\" is not acceptable for Debian packages (found in %s)", rel.RelatedPackage, relType)
		}

		for _, constraint := range rel.Constraints {
			if !packageVersionRx.MatchString(constraint.Version) {
				return fmt.Errorf(
					"Version in \"%s %s %s\" is not acceptable for Debian packages (found in %s)",
					rel.RelatedPackage, constraint.Relation, constraint.Version, relType,
				)
			}
		}
	}

	return nil
}
