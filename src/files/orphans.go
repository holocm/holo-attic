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
)

//Some shared logic for `holo scan` and `holo apply` concerning orphaned backup
//files. Find the corresponding target file, and assess the situation.
func ScanOrphanedBackupFile(backupPath string) (targetPath, strategy, assessment string) {
	target := NewConfigFileFromBackupPath(backupPath).TargetPath()
	if IsManageableFile(target) {
		return target, "restore", "all repository files were deleted"
	} else {
		return target, "delete", "target was deleted"
	}
}

//Clean up an orphaned backup file.
func HandleOrphanedBackupFile(backupPath string) {
	targetPath, strategy, assessment := ScanOrphanedBackupFile(backupPath)
	common.PrintInfo(" Scrubbing \x1b[1m%s\x1b[0m (%s)", targetPath, assessment)
	common.PrintInfo("%10s %s", strategy, backupPath)

	switch strategy {
	case "delete":
		//target is gone - delete the backup file
		err := os.Remove(backupPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		//if there was a .pacsave file, we can delete it too
		pacsavePath := targetPath + ".pacsave"
		if IsManageableFile(pacsavePath) {
			common.PrintInfo("    delete %s", pacsavePath)
			err := os.Remove(pacsavePath)
			if err != nil {
				common.PrintError(err.Error())
				return
			}
		}
	case "restore":
		//target is still there - restore the backup file
		err := CopyFile(backupPath, targetPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		err = os.Remove(backupPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
	}

	//TODO: cleanup empty directories below BackupDirectory()
}
