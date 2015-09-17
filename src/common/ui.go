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

import "fmt"

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
