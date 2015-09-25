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

package platform

import "../common"

//dpkgImpl provides the platform.Impl for dpkg-based distributions (Debian and derivatives).
type dpkgImpl struct{}

func (p dpkgImpl) FindUpdatedTargetBase(targetPath string) string {
	dpkgDistPath := targetPath + ".dpkg-dist"
	if common.IsManageableFile(dpkgDistPath) {
		return dpkgDistPath
	}
	return ""
}

func (p dpkgImpl) FindConfigBackup(targetPath string) string {
	dpkgOldPath := targetPath + ".dpkg-old"
	if common.IsManageableFile(dpkgOldPath) {
		return dpkgOldPath
	}
	return ""
}

func (p dpkgImpl) AdditionalCleanupTargets(targetPath string) []string {
	//not used by dpkg
	return []string{}
}
