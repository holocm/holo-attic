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

package scripts

import (
	"os"

	"../../shared"
	"../common"
)

//Scan returns a slice with all the provisioning scripts that have been found.
func Scan() common.Entities {
	errorReport := shared.Report{Action: "scan", Target: "provisioning scripts"}

	//look in the script directory for script definitions
	paths, err := common.ScanDirectory(common.ScriptDirectory(), func(fi os.FileInfo) bool {
		return common.IsManageableFileInfo(fi)
	})
	if err != nil {
		errorReport.AddError(err.Error())
		errorReport.Print()
		return nil
	}

	//create entities for each path
	entities := make(common.Entities, 0, len(paths))
	for _, path := range paths {
		entities = append(entities, ProvisioningScript{path})
	}

	return entities
}
