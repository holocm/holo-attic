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
	"os"
	"path/filepath"

	"../../shared"
)

var targetDirectory = "/"
var entityDirectory = "/usr/share/holo/users-groups"
var repoDirectory = "/usr/share/holo/files"
var targetBaseDirectory = "/var/lib/holo/files/base"
var provisionedDirectory = "/var/lib/holo/files/provisioned"

func init() {
	if value := os.Getenv("HOLO_ROOT_DIR"); value != "" {
		targetDirectory = value
		entityDirectory = filepath.Join(value, entityDirectory[1:])
		repoDirectory = filepath.Join(value, repoDirectory[1:])
		targetBaseDirectory = filepath.Join(value, targetBaseDirectory[1:])
		provisionedDirectory = filepath.Join(value, provisionedDirectory[1:])
	}

	//all these directories need to exist (most directories are checked for
	//existence in plugins.ReadConfiguration now, so we limit ourselves to the
	//stuff in /var)
	dirs := []string{targetBaseDirectory, provisionedDirectory}
	errorReport := shared.Report{Action: "Errors occurred during", Target: "startup"}
	hasError := false
	for _, dir := range dirs {
		fi, err := os.Stat(dir)
		switch {
		case err != nil:
			errorReport.AddError("Cannot open %s: %s", dir, err.Error())
			hasError = true
		case !fi.IsDir():
			errorReport.AddError("Cannot open %s: not a directory!", dir)
			hasError = true
		}
	}
	if hasError {
		errorReport.Print()
		panic("startup failed")
	}
}

//TargetDirectory is usually the root directory "/" and can be set with the
//environment variable HOLO_ROOT_DIR (usually only within unit tests).
func TargetDirectory() string {
	return targetDirectory
}

//EntityDirectory is derived from the TargetDirectory() as
//"$target_dir/usr/share/holo".
func EntityDirectory() string {
	return entityDirectory
}

//RepoDirectory is derived from the TargetDirectory() as
//"$target_dir/usr/share/holo/files".
func RepoDirectory() string {
	return repoDirectory
}

//TargetBaseDirectory is derived from the TargetDirectory() as
//"$target_dir/var/lib/holo/files/base".
func TargetBaseDirectory() string {
	return targetBaseDirectory
}

//ProvisionedDirectory is derived from the TargetDirectory() as
//"$target_dir/var/lib/holo/files/provisioned".
func ProvisionedDirectory() string {
	return provisionedDirectory
}
