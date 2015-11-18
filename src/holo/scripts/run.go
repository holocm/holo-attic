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
	"bytes"
	"os/exec"
	"path/filepath"

	"../../shared"
)

//ProvisioningScript represents a provisioning script that can be run by Holo
//(an executable file below /usr/share/holo/provision).
type ProvisioningScript struct {
	path string
}

//EntityID implements the common.Entity interface for ProvisioningScript.
func (s ProvisioningScript) EntityID() string {
	return "script:" + filepath.Base(s.path)
}

//Report implements the common.Entity interface for ProvisioningScript.
func (s ProvisioningScript) Report() *shared.Report {
	r := shared.Report{Target: s.EntityID()}
	r.AddLine("found at", s.path)
	return &r
}

//Apply implements the common.Entity interface for ProvisioningScript.
func (s ProvisioningScript) Apply(withForce bool) {
	report := s.Report()
	report.Action = "Executing"

	//run program, buffer stdout/stderr
	var stdout bytes.Buffer
	cmd := exec.Command(s.path)
	cmd.Stdin = nil
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	err := cmd.Run()
	if err != nil {
		report.AddError(err.Error())
	}

	report.AddLog(string(stdout.Bytes()))
	report.Print()
}

//RenderDiff implements the common.Entity interface for ProvisioningScript.
func (s ProvisioningScript) RenderDiff() ([]byte, error) {
	//diffs are not applicable to scripts, so always return an empty diff
	return nil, nil
}
