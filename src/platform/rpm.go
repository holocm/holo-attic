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

//rpmImpl provides the platform.Impl for RPM-based distributions.
type rpmImpl struct{}

func (p rpmImpl) FindUpdatedTargetBase(targetPath string) string {
	rpmnewPath := targetPath + ".rpmnew"
	if common.IsManageableFile(rpmnewPath) {
		return rpmnewPath
	}
	return ""
}

func (p rpmImpl) AdditionalCleanupTargets(targetPath string) []string {
	//there seems to be no RPM equivalent to Arch's .pacsave
	//(.rpmsave is *not* the same thing, it's a reverse variant of
	//.rpmnew that doesn't seem to be commonly used anymore)
	return []string{}
}
