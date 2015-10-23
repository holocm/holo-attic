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
	"./common"
	"./pacman"
)

const (
	formatAuto = iota
	formatPacman
)

type options struct {
	format        int
	printToStdout bool
}

func main() {
	opts, earlyExit := parseArgs()
	if earlyExit {
		return
	}
	generator := findGenerator(opts.format)

	//read package definition from stdin
	r := shared.Report{Action: "read", Target: "package definition"}
	pkg, hasError := common.ParsePackageDefinition(os.Stdin, &r)
	if hasError {
		r.Print()
		os.Exit(1)
	}

	//build package
	err := pkg.Build(generator, opts.printToStdout)
	if err != nil {
		r = shared.Report{Action: "build", Target: fmt.Sprintf("%s-%s", pkg.Name, pkg.Version)}
		r.AddError(err.Error())
		r.Print()
		os.Exit(2)
	}
}

func parseArgs() (result options, exit bool) {
	//default settings
	opts := options{
		format:        formatAuto,
		printToStdout: false,
	}

	//parse arguments
	args := os.Args[1:]
	r := shared.Report{Action: "parse", Target: "arguments"}
	hasArgsError := false
	for _, arg := range args {
		switch arg {
		case "--help":
			printHelp()
			return opts, true
		case "--version":
			fmt.Println(shared.VersionString())
			return opts, true
		case "--print":
			opts.printToStdout = true
		case "--no-print":
			opts.printToStdout = false
		case "--pacman":
			if opts.format != formatAuto {
				r.AddError("Multiple package formats specified.")
				hasArgsError = true
			}
			opts.format = formatPacman
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

	return opts, false
}

func printHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <options> < definitionfile > packagefile\n\nOptions:\n", program)
	fmt.Println("  --pacman\t\tBuild a pacman package\n")
	fmt.Println("If no options are given, the package format for the current distribution is selected.\n")
}

func findGenerator(format int) common.Generator {
	switch format {
	case formatAuto:
		//which distribution are we running on?
		isDist := shared.GetCurrentDistribution()
		switch {
		case isDist["arch"]:
			return &pacman.Generator{}
		default:
			shared.ReportUnsupportedDistribution(isDist)
			return nil
		}
	case formatPacman:
		return &pacman.Generator{}
	default:
		panic("Impossible format")
	}
}
