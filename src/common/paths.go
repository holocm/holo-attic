/*******************************************************************************
*
*   Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
*   This program is free software; you can redistribute it and/or modify it
*   under the terms of the GNU General Public License as published by the Free
*   Software Foundation; either version 2 of the License, or (at your option)
*   any later version.
*
*   This program is distributed in the hope that it will be useful, but WITHOUT
*   ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
*   FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
*   more details.
*
*   You should have received a copy of the GNU General Public License along
*   with this program; if not, write to the Free Software Foundation, Inc.,
*   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.
*
********************************************************************************/

package common

import (
	"os"
	"path/filepath"
)

var targetDirectory string = "/"
var entityDirectory string = "/usr/share/holo"
var repoDirectory string = "/usr/share/holo/repo"
var backupDirectory string = "/var/lib/holo/backup"

func init() {
	if value := os.Getenv("HOLO_CHROOT_DIR"); value != "" {
		targetDirectory = value
		entityDirectory = filepath.Join(value, entityDirectory[1:])
		repoDirectory = filepath.Join(value, repoDirectory[1:])
		backupDirectory = filepath.Join(value, backupDirectory[1:])
	}
}

//The TargetDirectory (usually the root directory "/") can be set with the
//environment variable HOLO_CHROOT_DIR (usually only within unit tests).
func TargetDirectory() string {
	return targetDirectory
}

//The EntityDirectory is derived from the TargetDirectory() as
//"$target_dir/usr/share/holo".
func EntityDirectory() string {
	return entityDirectory
}

//The RepoDirectory is derived from the TargetDirectory() as
//"$target_dir/usr/share/holo/repo".
func RepoDirectory() string {
	return repoDirectory
}

//The RepoDirectory is derived from the TargetDirectory() as
//"$target_dir/var/lib/holo/backup".
func BackupDirectory() string {
	return backupDirectory
}
