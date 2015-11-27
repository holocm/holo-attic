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

package shared

import (
	"bytes"
	"os/exec"
	"strings"
)

//ExecProgram is a wrapper around exec.Command that reports any stderr output
//of the child process to the given Report automatically.
func ExecProgram(report *Report, stdin []byte, command string, arguments ...string) (output []byte, err error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(command, arguments...)
	cmd.Stdin = bytes.NewBuffer(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		report.AddWarning("execution of %s produced error output:", command)
		stderrLines := strings.Split(strings.Trim(stderr.String(), "\n"), "\n")
		for _, stderrLine := range stderrLines {
			report.AddWarning("    " + stderrLine)
		}
	}
	return stdout.Bytes(), err
}
