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

package common

import (
	"fmt"
	"strings"
)

func msg(color, message string) {
	fmt.Printf("\x1b[%sm\x1b[1m[holo]\x1b[0m %s\n", color, message)
}

//PrintError formats the given error message on stdout, similar to fmt.Printf.
func PrintError(message string, a ...interface{}) {
	msg("31", fmt.Sprintf(message, a...))
}

//PrintInfo formats the given info message on stdout, similar to fmt.Printf.
func PrintInfo(message string, a ...interface{}) {
	msg("38", fmt.Sprintf(message, a...))
}

//PrintWarning formats the given warning message on stdout, similar to fmt.Printf.
func PrintWarning(message string, a ...interface{}) {
	msg("33", fmt.Sprintf(message, a...))
}

type reportLine struct {
	key   string
	value string
}

//Report formats information for an action taken on a single target, including
//warning and error messages.
type Report struct {
	Action    string
	Target    string
	State     string
	infoLines []string
	msgText   string
}

//AddLine adds an information line to the given Report.
func (r *Report) AddLine(key, value string) {
	//format contents appropriately
	line := fmt.Sprintf("%12s %s", key, value)
	r.infoLines = append(r.infoLines, line)
}

func (r *Report) addMessage(color, text string) {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	r.msgText += fmt.Sprintf("\x1b[%sm%s\x1b[0m", color, text)
}

//AddWarning adds a warning message to the given Report.
func (r *Report) AddWarning(text string) { r.addMessage("33", text) }

//AddError adds an error message to the given Report.
func (r *Report) AddError(text string) { r.addMessage("31", text) }

//Print prints the full report on stdout.
func (r *Report) Print() {
	//print initial line with Action, Target and State
	if r.Action == "" {
		fmt.Printf("\x1b[1m%s\x1b[0m", r.Target)
	} else {
		fmt.Printf("%s \x1b[1m%s\x1b[0m", r.Action, r.Target)
	}
	if r.State == "" {
		fmt.Println()
	} else {
		fmt.Printf(" (%s)\n", r.State)
	}

	//print infoLines
	for _, line := range r.infoLines {
		fmt.Println(line)
	}
	fmt.Println()

	//print message text, if any
	if r.msgText != "" {
		fmt.Println(r.msgText) //including trailing newline
	}
}
