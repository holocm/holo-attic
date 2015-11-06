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

package pacman

import (
	"fmt"
	"regexp"
	"strings"

	"../common"
)

//NOTE: pacman does not actually accept dashes in version strings, but
//holo-build does, so we replace these by underscores in fullVersionString()
var packageNameRx = regexp.MustCompile(`^[a-z0-9@._+][a-z0-9@._+-]*$`)
var packageVersionRx = regexp.MustCompile(`^[a-zA-Z0-9.-_]*$`)

func validatePackage(pkg *common.Package) error {
	if !packageNameRx.MatchString(pkg.Name) {
		return fmt.Errorf("Package name \"%s\" is not acceptable for Pacman packages", pkg.Name)
	}
	if !packageVersionRx.MatchString(pkg.Version) {
		//this check is only some Defense in Depth; a stricted version format
		//is already enforced by the generator-independent validation
		return fmt.Errorf("Package version \"%s\" is not acceptable for Pacman packages", pkg.Version)
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
		name := rel.RelatedPackage
		//for requirements, allow special syntaxes "group:foo", "except:bar"
		//and "except:group:qux"
		if relType == "requires" {
			name = strings.TrimPrefix(name, "except:")
			name = strings.TrimPrefix(name, "group:")
		}

		if !packageNameRx.MatchString(name) {
			return fmt.Errorf("Package name \"%s\" is not acceptable for Pacman packages (found in %s)", rel.RelatedPackage, relType)
		}
	}

	return nil
}
