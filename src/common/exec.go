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

package common

import (
	"bytes"
	"os/exec"
	"strings"
)

//ExecProgram is a wrapper around exec.Command that reports any stderr output
//of the child process automatically.
func ExecProgram(stdin []byte, command string, arguments ...string) (output []byte, err error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(command, arguments...)
	cmd.Stdin = bytes.NewBuffer(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		PrintWarning("execution of %s produced error output:", command)
		stderrLines := strings.Split(strings.Trim(stderr.String(), "\n"), "\n")
		for _, stderrLine := range stderrLines {
			PrintWarning("    %s", stderrLine)
		}
	}
	return stdout.Bytes(), err
}
