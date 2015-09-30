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

package platform

import "../common"

//rpmImpl provides the platform.Impl for RPM-based distributions.
type rpmImpl struct{}

func (p rpmImpl) FindUpdatedTargetBase(targetPath string) string {
	rpmnewPath := targetPath + ".rpmnew"
	if common.IsManageableFile(rpmnewPath) {
		return rpmnewPath
	}
	return ""
}

func (p rpmImpl) FindConfigBackup(targetPath string) string {
	rpmsavePath := targetPath + ".rpmsave"
	if common.IsManageableFile(rpmsavePath) {
		return rpmsavePath
	}
	return ""
}

func (p rpmImpl) AdditionalCleanupTargets(targetPath string) []string {
	//not used by RPM
	return []string{}
}
