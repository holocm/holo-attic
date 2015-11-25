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

package plugins

import (
	"../common"
	"../entities"
	"../files"
	"../scripts"
)

//InfoLine represents a line in the information section of an Entity.
type InfoLine struct {
	attribute string
	value     string
}

//Entity represents an entity known to some Holo plugin.
type Entity struct {
	plugin       *Plugin
	id           string
	actionVerb   string
	actionReason string
}

//Scan discovers entities available for the given entity. Errors are reported
//immediately and will result in nil being returned. "No entities found" will
//be reported as a non-nil empty slice.
//there are no entities.
func (p *Plugin) Scan() common.Entities {
	//plugins with the "built-in" flag do their processing in other scan functions
	switch p.ID() {
	case "files":
		return files.ScanRepo()
	case "users-groups":
		return entities.Scan()
	case "run-scripts":
		return scripts.Scan()
	default: //follows below
	}

	//TODO
	return nil
}
