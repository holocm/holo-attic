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

package holo

import "os"

func getenvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	} else {
		return value
	}
}

//The target directory (usually the root directory "/") can be set with the
//environment variable HOLO_TARGET_DIR (usually only within unit tests).
func TargetDirectory() string {
	return getenvOrDefault("HOLO_TARGET_DIR", "/")
}

//The repo directory (usually "/holo/repo") can be set with the environment
//variable HOLO_REPO_DIR (usually only within unit tests).
func RepoDirectory() string {
	return getenvOrDefault("HOLO_REPO_DIR", "/holo/repo")
}

//The backup directory (usually "/holo/backup") can be set with the environment
//variable HOLO_BACKUP_DIR (usually only within unit tests).
func BackupDirectory() string {
	return getenvOrDefault("HOLO_BACKUP_DIR", "/holo/backup")
}
