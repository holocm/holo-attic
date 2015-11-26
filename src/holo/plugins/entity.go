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
	"bytes"
	"fmt"
	"os"

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
	r.Print()

	command := "apply"
	if withForce {
		command = "apply-force"
	}

	err := e.plugin.Run([]string{command, e.id}, os.Stdout, os.Stderr)

	if err != nil {
		fmt.Printf("apply %s failed: %s\n", e.id, err.Error())
	}
	fmt.Println() //ensure newline between output and next report
}

//RenderDiff implements the common.Entity interface.
func (e *Entity) RenderDiff() ([]byte, error) {
	var buffer bytes.Buffer
	err := e.plugin.Run([]string{"diff", e.id}, &buffer, os.Stderr)
	return buffer.Bytes(), err
}
