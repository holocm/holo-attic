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

//Apply performs the complete application algorithm for the given TargetFile.
//This includes taking a copy of the target base if necessary, applying all
//repository entries, and saving the result in the target path with the correct
//file metadata.
func apply(target *TargetFile, report *common.Report, withForce bool) (skipReport bool) {
	//determine the related paths
	targetPath := target.PathIn(common.TargetDirectory())
	targetBasePath := target.PathIn(common.TargetBaseDirectory())

	//step 1: will only apply targets if:
	//option 1: there is a manageable file in the target location (this target
	//file is either the target base from the application package or the
	//product of a previous Apply run)
	//option 2: the target file was deleted, but we have a target base that we
	//can start from
	if !common.IsManageableFile(targetPath) {
		if !common.IsManageableFile(targetBasePath) {
			report.AddError("skipping target: not a manageable file")
			return false
		}
		if !withForce {
			report.AddError("skipping target: file has been deleted by user (use --force to restore)")
			return false
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
			return false
		}

		err = common.CopyFile(targetPath, targetBasePath)
		if err != nil {
			report.AddError("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
			return false
		}
	}

	//step 3: check if a system update installed a new version of the stock
	//configuration
	updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(targetPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}
	if updatedTBPath != "" {
		//an updated stock configuration is available at updatedTBPath
		report.ReplaceLine(0, "update", fmt.Sprintf("%s -> %s", reportedTBPath, targetBasePath))
		err := common.CopyFile(updatedTBPath, targetBasePath)
		if err != nil {
			report.AddError("Cannot copy %s to %s: %s", updatedTBPath, targetBasePath, err.Error())
			return false
		}
		_ = os.Remove(updatedTBPath) //this can fail silently
	}

	//step 4: apply the repo files *if* the version at targetPath is the one
	//installed by the package (which can be found at targetBasePath); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)
	var lastProvisionedBuffer *FileBuffer
	lastProvisionedPath := target.PathIn(common.ProvisionedDirectory())
	if !withForce && common.IsManageableFile(lastProvisionedPath) {
		targetBuffer, err := NewFileBuffer(targetPath, targetPath)
		if err != nil {
			report.AddError(err.Error())
			return false
		}
		lastProvisionedBuffer, err = NewFileBuffer(lastProvisionedPath, targetPath)
		if err != nil {
			report.AddError(err.Error())
			return false
		}
		if !targetBuffer.EqualTo(lastProvisionedBuffer) {
			report.AddError("skipping target: file has been modified by user (use --force to overwrite)")
			return false
		}
	}

	//step 4a: load the target base into a buffer as the start for the
	//application algorithm
	buffer, err := NewFileBuffer(targetBasePath, targetPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}

	//step 4b: apply all the applicable repo files in order
	repoEntries := target.RepoEntries()
	for _, repoFile := range repoEntries {
		buffer, err = GetApplyImpl(repoFile)(buffer, report)
		if err != nil {
			report.AddError(err.Error())
			return false
		}
	}

	//step 4c: don't do anything more if nothing has changed
	if !withForce && lastProvisionedBuffer != nil {
		if buffer.EqualTo(lastProvisionedBuffer) {
			//since we did not do anything, don't report this
			return true
		}
	}

	//step 4c: save a copy of the provisioned config file to check for manual
	//modifications in the next Apply() run
	provisionedDir := filepath.Dir(lastProvisionedPath)
	err = os.MkdirAll(provisionedDir, 0755)
	if err != nil {
		report.AddError("Cannot write %s: %s", lastProvisionedPath, err.Error())
		return false
	}
	err = buffer.Write(lastProvisionedPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}
	err = common.ApplyFilePermissions(targetBasePath, lastProvisionedPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}

	//step 4d: write the result buffer to the target location and copy
	//owners/permissions from target base to target file
	err = buffer.Write(targetPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}
	err = common.ApplyFilePermissions(targetBasePath, targetPath)
	if err != nil {
		report.AddError(err.Error())
		return false
	}

	return false
}
