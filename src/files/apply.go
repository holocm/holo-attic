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
	"fmt"
	"os"
	"path/filepath"

	"../common"
	"../platform"
)

//Apply performs the complete application algorithm for the given ConfigFile.
//This includes taking a copy of the target base if necessary, applying all
//repo files, and saving the result in the target path with the correct file
//metadata.
func Apply(target *TargetFile, report *common.Report, withForce bool) {
	//determine the related paths
	targetPath := target.PathIn(common.TargetDirectory())
	targetBasePath := target.PathIn(common.TargetBaseDirectory())

	//step 1: will only install files from repo if:
	//option 1: there is a corresponding regular file in the target location
	//(that file comes from the application package, the repo file from the
	//holo metapackage)
	//option 2: the target file was deleted, but we have a target base that we can start from
	if !common.IsManageableFile(targetPath) {
		if !common.IsManageableFile(targetBasePath) {
			report.AddError("skipping target: not a manageable file")
			return
		}
		if !withForce {
			report.AddError("skipping target: file has been deleted by user (use --force to overwrite)")
			return
		}
	}

	//step 2: if we don't have a target base yet, the file at targetPath *is*
	//the targetBase which we have to copy now
	if common.IsManageableFile(targetBasePath) {
		report.ReplaceLine(0, "", "") //remove the "store at $target_base" line because we did not do that
	} else {
		targetBaseDir := filepath.Dir(targetBasePath)
		err := os.MkdirAll(targetBaseDir, 0755)
		if err != nil {
			report.AddError("Cannot create directory %s: %s", targetBaseDir, err.Error())
			return
		}

		err = common.CopyFile(targetPath, targetBasePath)
		if err != nil {
			report.AddError("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
			return
		}
	}

	//step 3: check if a system update installed a new version of the stock
	//configuration
	updateBackupPath := platform.Implementation().FindConfigBackup(targetPath)
	lastInstalledTargetPath := targetPath
	if updateBackupPath != "" {
		//case 1: yes, the targetPath is an updated stock configuration and the
		//old targetPath (last written by Holo) was moved to updateBackupPath
		//(this code path is used for .rpmsave and .dpkg-old files)
		report.ReplaceLine(0, "update", fmt.Sprintf("%s -> %s", targetPath, targetBasePath))
		err := common.CopyFile(targetPath, targetBasePath)
		if err != nil {
			report.AddError("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
			return
		}
		//since the target file that we wrote last time has been moved to a
		//different place, we need to use the changed path for the mtime check
		//in the next step
		lastInstalledTargetPath = updateBackupPath
	} else {
		updatePath := platform.Implementation().FindUpdatedTargetBase(targetPath)
		if updatePath != "" {
			//case 2: yes, an updated stock configuration is available at updatePath
			//(this code path is used for .rpmnew, .dpkg-dist and .pacnew files)
			report.ReplaceLine(0, "update", fmt.Sprintf("%s -> %s", updatePath, targetBasePath))
			err := common.CopyFile(updatePath, targetBasePath)
			if err != nil {
				report.AddError("Cannot copy %s to %s: %s", updatePath, targetBasePath, err.Error())
				return
			}
			_ = os.Remove(updatePath) //this can fail silently
		}
	}

	//step 4: apply the repo files *if* the version at targetPath is the one
	//installed by the package (which can be found at targetBasePath); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)
	provisionedPath := target.PathIn(common.ProvisionedDirectory())
	if !withForce && common.IsManageableFile(provisionedPath) {
		targetBuffer, err := NewFileBuffer(lastInstalledTargetPath, targetPath)
		if err != nil {
			report.AddError(err.Error())
			return
		}
		lastProvisionedBuffer, err := NewFileBuffer(provisionedPath, targetPath)
		if err != nil {
			report.AddError(err.Error())
			return
		}
		if !targetBuffer.EqualTo(lastProvisionedBuffer) {
			report.AddError("skipping target: file has been modified by user (use --force to overwrite)")
			return
		}
	}

	//step 4a: load the target base into a buffer as the start for the
	//application algorithm
	buffer, err := NewFileBuffer(targetBasePath, targetPath)
	if err != nil {
		report.AddError(err.Error())
		return
	}

	//step 4b: apply all the applicable repo files in order
	repoEntries := target.RepoEntries()
	for _, repoFile := range repoEntries {
		buffer, err = GetApplyImpl(repoFile)(buffer)
		if err != nil {
			report.AddError(err.Error())
			return
		}
	}

	//step 4c: save a copy of the provisioned config file to check for manual
	//modifications in the next Apply() run
	provisionedDir := filepath.Dir(provisionedPath)
	err = os.MkdirAll(provisionedDir, 0755)
	if err != nil {
		report.AddError("Cannot write %s: %s", provisionedPath, err.Error())
		return
	}
	err = buffer.Write(provisionedPath)
	if err != nil {
		report.AddError(err.Error())
		return
	}
	err = common.ApplyFilePermissions(targetBasePath, provisionedPath)
	if err != nil {
		report.AddError(err.Error())
		return
	}

	//step 4d: write the result buffer to the target location and copy
	//owners/permissions from target base to target file
	err = buffer.Write(targetPath)
	if err != nil {
		report.AddError(err.Error())
		return
	}
	err = common.ApplyFilePermissions(targetBasePath, targetPath)
	if err != nil {
		report.AddError(err.Error())
		return
	}

	//step 5: cleanup the updateBackupPath now that we successfully generated a
	//new version of the desired target
	if updateBackupPath != "" {
		report.AddLine("delete", updateBackupPath)
		_ = os.Remove(updateBackupPath) //this can fail silently
	}
}
