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
	"fmt"

	"../../shared"
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
	infoLines    []InfoLine
}

//EntityID implements the common.Entity interface.
func (e *Entity) EntityID() string { return e.id }

//Report implements the common.Entity interface.
func (e *Entity) Report() *shared.Report {
	r := shared.Report{Target: e.id, State: e.actionReason}
	for _, infoLine := range e.infoLines {
		r.AddLine(infoLine.attribute, infoLine.value)
	}
	return &r
}

//Apply implements the common.Entity interface.
func (e *Entity) Apply(withForce bool) {
	r := e.Report()
	r.Action = e.actionVerb

	r.AddError("TODO: apply %s\n", e.id)
	r.Print()
}

//RenderDiff implements the common.Entity interface.
func (e *Entity) RenderDiff() ([]byte, error) {
	fmt.Printf("TODO: diff %s\n", e.id)
	return nil, nil
}
