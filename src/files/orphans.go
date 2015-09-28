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

	"../common"
	"../platform"
)

//ScanOrphanedBackupFile locates a target file for a given orphaned backup file
//and assesses the situation. This logic is grouped in one function because
//it's used by both `holo scan` and `holo apply`.
func ScanOrphanedBackupFile(backupPath string) (targetPath, strategy, assessment string) {
	target := NewConfigFileFromBackupPath(backupPath).TargetPath()
	if common.IsManageableFile(target) {
		return target, "restore", "all repository files were deleted"
	}
	return target, "delete", "target was deleted"
}

//HandleOrphanedBackupFile cleans up an orphaned backup file.
func HandleOrphanedBackupFile(backupPath string) {
	targetPath, strategy, assessment := ScanOrphanedBackupFile(backupPath)
	common.PrintInfo(" Scrubbing \x1b[1m%s\x1b[0m (%s)", targetPath, assessment)
	common.PrintInfo("%10s %s", strategy, backupPath)

	switch strategy {
	case "delete":
		//target is gone - delete the computed target and the backup file
		computedPath := NewConfigFileFromBackupPath(backupPath).ComputedPath()
		err := os.Remove(computedPath)
		if err != nil && !os.IsNotExist(err) {
			common.PrintError(err.Error())
			return
		}
		err = os.Remove(backupPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		//if the package management left behind additional cleanup targets
		//(most likely a backup of our custom configuration), we can delete
		//these too
		cleanupTargets := platform.Implementation().AdditionalCleanupTargets(targetPath)
		for _, otherFile := range cleanupTargets {
			common.PrintInfo("    delete %s", otherFile)
			err := os.Remove(otherFile)
			if err != nil {
				common.PrintError(err.Error())
				return
			}
		}
	case "restore":
		//target is still there - restore the backup file
		err := common.CopyFile(backupPath, targetPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		//target is not managed by Holo anymore, so delete the computed target and the backup file
		computedPath := NewConfigFileFromBackupPath(backupPath).ComputedPath()
		err = os.Remove(computedPath)
		if err != nil && !os.IsNotExist(err) {
			common.PrintError(err.Error())
			return
		}
		err = os.Remove(backupPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
	}

	//TODO: cleanup empty directories below BackupDirectory() and ComputedDirectory()
}
