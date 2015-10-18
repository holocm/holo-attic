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
	hasFile, file := false, ""

	//parse arguments
	args := os.Args[1:]
	for _, arg := range args {
		switch arg {
		case "--help":
			printHelp()
			return
		case "--version":
			fmt.Println(shared.VersionString())
			return
		case "--pacman":
			//arg selects pacman package format
			if format != formatAuto {
				errorMultipleFormats()
			}
			format = formatPacman
		default:
			//arg is a file
			if hasFile {
				r := shared.Report{Action: "parse", Target: "arguments"}
				r.AddError("Multiple package description files specified.")
				r.Print()
				os.Exit(1)
			}
			hasFile, file = true, arg
		}
	}
	if !hasFile {
		//missing a file
		printHelp()
		os.Exit(1)
	}

	//TODO: unfinished :)
	fmt.Printf("Building file %s for format %d\n", file, format)
}

func printHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <options> file\n\nOptions:\n", program)
	fmt.Println("  --pacman\t\tBuild a pacman package\n")
	fmt.Println("If no options are given, the package format for the current distribution is selected.\n")
}

func errorMultipleFormats() {
	r := shared.Report{Action: "parse", Target: "arguments"}
	r.AddError("Multiple package formats specified.")
	r.Print()
	os.Exit(1)
}
