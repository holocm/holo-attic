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

import "os"

var targetDirectory string = "/"
var repoDirectory string = "/holo/repo"
var backupDirectory string = "/holo/backup"

func init() {
	if value := os.Getenv("HOLO_TARGET_DIR"); value != "" {
		targetDirectory = value
	}
	if value := os.Getenv("HOLO_REPO_DIR"); value != "" {
		repoDirectory = value
	}
	if value := os.Getenv("HOLO_BACKUP_DIR"); value != "" {
		backupDirectory = value
	}
}

//The target directory (usually the root directory "/") can be set with the
//environment variable HOLO_TARGET_DIR (usually only within unit tests).
func TargetDirectory() string {
	return targetDirectory
}

//The repo directory (usually "/holo/repo") can be set with the environment
//variable HOLO_REPO_DIR (usually only within unit tests).
func RepoDirectory() string {
	return repoDirectory
}

//The backup directory (usually "/holo/backup") can be set with the environment
//variable HOLO_BACKUP_DIR (usually only within unit tests).
func BackupDirectory() string {
	return backupDirectory
}
