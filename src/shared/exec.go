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
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var mock = false

func init() {
	if value := os.Getenv("HOLO_MOCK"); value == "1" {
		mock = true
	}
}

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

//ExecProgramOrMock works like ExecProgram, but when the environment variable
//HOLO_MOCK=1 is set, it will only print the command name and return success
//and empty stdout without executing the command.
func ExecProgramOrMock(report *Report, stdin []byte, command string, arguments ...string) (output []byte, err error) {
	if mock {
		report.AddWarning("MOCK: %s", shellEscapeArgs(append([]string{command}, arguments...)))
		return []byte{}, nil
	}
	o, e := ExecProgram(report, stdin, command, arguments...)
	return o, e
}

func shellEscapeArgs(arguments []string) string {
	//a puny caricature of an actual shell-escape
	var escapedArgs []string
	for _, arg := range arguments {
		if arg == "" || strings.Contains(arg, " ") {
			arg = fmt.Sprintf("'%s'", arg)
		}
		escapedArgs = append(escapedArgs, arg)
	}
	return strings.Join(escapedArgs, " ")
}
