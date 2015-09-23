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

package files

import (
	"os"
	"path/filepath"

	"../common"
	"../platform"
)

//Apply performs the complete application algorithm for the given ConfigFile.
//This includes taking a backup if necessary, applying all repo files, and
//saving the result in the target path with the correct file metadata.
func Apply(file ConfigFile, withForce bool) {
	//determine the related paths
	targetPath := file.TargetPath()
	backupPath := file.BackupPath()

	//step 1: will only install files from repo if there is a corresponding
	//regular file in the target location (that file comes from the application
	//package, the repo file from the holo metapackage)
	common.PrintInfo("Working on \x1b[1m%s\x1b[0m", targetPath)
	if !common.IsManageableFile(targetPath) {
		common.PrintError("  skipped: target is not a manageable file")
		return
	}

	//step 2: we know that a file exists at installPath; if we don't have a
	//backup of the original file, the file at installPath *is* the original
	//file which we have to backup now
	if !common.IsManageableFile(backupPath) {
		common.PrintInfo("  store at %s", backupPath)

		backupDir := filepath.Dir(backupPath)
		err := os.MkdirAll(backupDir, 0755)
		if err != nil {
			common.PrintError("Cannot create directory %s: %s", backupDir, err.Error())
			return
		}

		err = common.CopyFile(targetPath, backupPath)
		if err != nil {
			common.PrintError("Cannot copy %s to %s: %s", targetPath, backupPath, err.Error())
			return
		}
	}

	//step 2.5: if an updated version of the stock configuration file exists
	//next to the targetPath, move it to the backup location and use it as the
	//basis for step 3
	updatePath := platform.Implementation().FindUpdatedTargetBase(targetPath)
	if updatePath != "" {
		common.PrintInfo("    update %s -> %s", updatePath, backupPath)
		err := common.CopyFile(updatePath, backupPath)
		if err != nil {
			common.PrintError("Cannot copy %s to %s: %s", updatePath, backupPath, err.Error())
			return
		}
		_ = os.Remove(updatePath) //this can fail silently
	}

	//step 3: apply the repo files *if* the version at targetPath is the one
	//installed by the package (which can be found at backupPath); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)
	if !withForce {
		isNewer, err := common.IsNewerThan(targetPath, backupPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		if isNewer {
			common.PrintError("  skipped: target file has been modified by user (use --force to overwrite)")
			return
		}
	}

	//step 3a: load the backup file into a buffer as the start for the
	//application algorithm
	buffer, err := NewFileBuffer(backupPath, targetPath)
	if err != nil {
		common.PrintError(err.Error())
		return
	}

	//step 3b: apply all the applicable repo files in order
	repoFiles := file.RepoFiles()
	for _, repoFile := range repoFiles {
		common.PrintInfo("%10s %s", repoFile.ApplicationStrategy(), repoFile.Path())
		buffer, err = GetApplyImpl(repoFile)(buffer)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
	}

	//step 3c: write the result buffer to the target location and copy
	//permissions/timestamps from backup file to target file, in order to be
	//able to detect manual modifications in the next holo-apply run
	err = buffer.Write(targetPath)
	if err != nil {
		common.PrintError(err.Error())
		return
	}
	err = common.ApplyFilePermissions(backupPath, targetPath)
	if err != nil {
		common.PrintError(err.Error())
		return
	}
}
