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

package files

import (
	"os"

	"../common"
	"../platform"
)

//ScanOrphanedTargetBase locates a target file for a given orphaned target base
//and assesses the situation. This logic is grouped in one function because
//it's used by both `holo scan` and `holo apply`.
func ScanOrphanedTargetBase(targetBasePath string) (targetPath, strategy, assessment string) {
	target := NewConfigFileFromTargetBasePath(targetBasePath).TargetPath()
	if common.IsManageableFile(target) {
		return target, "restore", "all repository files were deleted"
	}
	return target, "delete", "target was deleted"
}

//HandleOrphanedTargetBase cleans up an orphaned target base.
func HandleOrphanedTargetBase(targetBasePath string) {
	targetPath, strategy, assessment := ScanOrphanedTargetBase(targetBasePath)
	common.PrintInfo(" Scrubbing \x1b[1m%s\x1b[0m (%s)", targetPath, assessment)
	common.PrintInfo("%10s %s", strategy, targetBasePath)

	switch strategy {
	case "delete":
		//target is gone - delete the provisioned target and the target base
		provisionedPath := NewConfigFileFromTargetBasePath(targetBasePath).ProvisionedPath()
		err := os.Remove(provisionedPath)
		if err != nil && !os.IsNotExist(err) {
			common.PrintError(err.Error())
			return
		}
		err = os.Remove(targetBasePath)
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
		//target is still there - restore the target base
		err := common.CopyFile(targetBasePath, targetPath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
		//target is not managed by Holo anymore, so delete the provisioned target and the target base
		provisionedPath := NewConfigFileFromTargetBasePath(targetBasePath).ProvisionedPath()
		err = os.Remove(provisionedPath)
		if err != nil && !os.IsNotExist(err) {
			common.PrintError(err.Error())
			return
		}
		err = os.Remove(targetBasePath)
		if err != nil {
			common.PrintError(err.Error())
			return
		}
	}

	//TODO: cleanup empty directories below TargetBaseDirectory() and ProvisionedDirectory()
}
