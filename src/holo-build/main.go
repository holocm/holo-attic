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

package main

import (
	"fmt"
	"os"

	"../shared"
)

const (
	formatAuto = iota
	formatPacman
)

func main() {
	//what can be in the arguments?
	format := formatAuto

	//parse arguments
	args := os.Args[1:]
	r := shared.Report{Action: "parse", Target: "arguments"}
	hasArgsError := false
	for _, arg := range args {
		switch arg {
		case "--help":
			printHelp()
			return
		case "--version":
			fmt.Println(shared.VersionString())
			return
		case "--pacman":
			if format != formatAuto {
				r.AddError("Multiple package formats specified.")
				hasArgsError = true
			}
			format = formatPacman
		default:
			r.AddError("Unrecognized argument: '%s'", arg)
			hasArgsError = true
		}
	}
	if hasArgsError {
		r.Print()
		printHelp()
		os.Exit(1)
	}

	//TODO: unfinished :)
	fmt.Printf("Building for format %d\n", format)
}

func printHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <options> < definitionfile > packagefile\n\nOptions:\n", program)
	fmt.Println("  --pacman\t\tBuild a pacman package\n")
	fmt.Println("If no options are given, the package format for the current distribution is selected.\n")
}
