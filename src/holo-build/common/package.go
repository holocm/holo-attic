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

//Package contains all information about a single package. This representation
//will be passed into the generator backends.
//
//General note on package references: Various struct fields (Requires,
//Provides, Conflicts, Replaces) are []string where the individual strings are
//other packages. These strings may be just a package name, or the package name
//may have a suffix like /[<>]?=?$version/ to specify a version constraint for
//the referenced package. The package may be referenced multiple times with
//different version constraints.
//
//    pkg.Requires = []string {
//        "foo", # any version of foo required
//        "bar<1.0", # requires a version of bar before 1.0
//        "baz>=
//    }
type Package struct {
	//Name is the package name.
	Name string
	//Version is the package's version, e.g. "1.2.3-1" or "2:20151024-1.1".
	//This field is not structured further in this level since the acceptable
	//version format may depend on the package generator used.
	Version string
	//Description is the optional package description.
	Description string
	//Requires contains a list of other packages that are required dependencies
	//for this package and thus must be installed together with this package.
	//This is called "Depends" by some package managers.
	Requires  []PackageRelation
	Provides  []PackageRelation
	Conflicts []PackageRelation
	Replaces  []PackageRelation
}

//PackageRelation declares a relation to another package. For the related
//package, any number of version constraints may be given. For example, the
//following snippet makes a Package require any version of package "foo", and
//at least version 2.1.2 (but less than version 3.0) of package "bar".
//
//    pkg.Requires := []PackageRelation{
//        PackageRelation { "foo", nil },
//        PackageRelation { "bar", []VersionConstraint{
//            VersionConstraint { ">=", "2.1.2" },
//            VersionConstraint { "<",  "3.0"   },
//        }
//    }
type PackageRelation struct {
	RelatedPackage string
	Constraints    []VersionConstraint
}

//VersionConstraint is used by the PackageRelation struct to specify version
//constraints for a related package.
type VersionConstraint struct {
	//Relation is one of "<", "<=", "=", ">=" or ">".
	Relation string
	//Version is the version on the right side of the Relation, e.g. "1.2.3-1"
	//or "2:20151024-1.1".  This field is not structured further in this level
	//since the acceptable version format may depend on the package generator
	//used.
	Version string
}
