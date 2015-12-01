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
	"errors"
	"fmt"
	"os"
	"syscall"

	"../shared"
	"./common"
	"./debian"
	"./pacman"
)

func main() {
	//holo-build needs to run in a fakeroot(1)
	if os.Getenv("FAKEROOTKEY") != "" {
		//already running in fakeroot, commence normal operation
		actualMain()
		return
	}

	//not running in fakeroot -> exec self with fakeroot
	args := append([]string{"/usr/bin/fakeroot"}, os.Args...)
	syscall.Exec(args[0], args, os.Environ())
}

const (
	formatAuto = iota
	formatPacman
	formatDebian
)

type options struct {
	format        int
	printToStdout bool
	reproducible  bool
	filenameOnly  bool
}

func actualMain() {
	opts, earlyExit := parseArgs()
	if earlyExit {
		return
	}
	generator := findGenerator(opts.format)

	//read package definition from stdin
	pkg, errs := common.ParsePackageDefinition(os.Stdin)

	//try to validate package
	var validateErrs []error
	if pkg != nil {
		validateErrs = generator.Validate(pkg)
	}
	errs = append(errs, validateErrs...)

	//did that go wrong?
	if len(errs) > 0 {
		for _, err := range errs {
			showError(err)
		}
		os.Exit(1)
	}

	//print filename instead of building package, if requested
	if opts.filenameOnly {
		fmt.Println(generator.RecommendedFileName(pkg))
		return
	}

	//build package
	err := pkg.Build(generator, opts.printToStdout, opts.reproducible)
	if err != nil {
		showError(fmt.Errorf("cannot build %s: %s\n",
			generator.RecommendedFileName(pkg), err.Error(),
		))
		os.Exit(2)
	}
}

func parseArgs() (result options, exit bool) {
	//default settings
	opts := options{
		format:        formatAuto,
		printToStdout: false,
		reproducible:  false,
	}

	//parse arguments
	args := os.Args[1:]
	hasArgsError := false
	for _, arg := range args {
		switch arg {
		case "--help":
			printHelp()
			return opts, true
		case "--version":
			fmt.Println(common.VersionString())
			return opts, true
		case "--stdout":
			opts.printToStdout = true
		case "--no-stdout":
			opts.printToStdout = false
		case "--suggest-filename":
			opts.filenameOnly = true
		case "--reproducible":
			opts.reproducible = true
		case "--no-reproducible":
			opts.reproducible = false
		case "--pacman":
			if opts.format != formatAuto {
				showError(errors.New("Multiple package formats specified."))
				hasArgsError = true
			}
			opts.format = formatPacman
		case "--debian":
			if opts.format != formatAuto {
				showError(errors.New("Multiple package formats specified."))
				hasArgsError = true
			}
			opts.format = formatDebian
		default:
			showError(fmt.Errorf("Unrecognized argument: '%s'", arg))
			hasArgsError = true
		}
	}
	if hasArgsError {
		printHelp()
		os.Exit(1)
	}

	return opts, false
}

func printHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <options> < definitionfile > packagefile\n\nOptions:\n", program)
	fmt.Println("  --stdout\t\tPrint resulting package on stdout")
	fmt.Println("  --no-stdout\t\tWrite resulting package to the working directory (default)")
	fmt.Println("  --reproducible\tBuild a reproducible package with bogus timestamps etc.")
	fmt.Println("  --no-reproducible\tBuild a non-reproducible package with actual timestamps etc. (default)")
	fmt.Println("  --debian\t\tBuild a debian package\n")
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
		case isDist["debian"]:
			return &debian.Generator{}
		default:
			shared.ReportUnsupportedDistribution(isDist)
			return nil
		}
	case formatPacman:
		return &pacman.Generator{}
	case formatDebian:
		return &debian.Generator{}
	default:
		panic("Impossible format")
	}
}

func showError(err error) {
	fmt.Fprintf(os.Stderr, "\x1b[31m\x1b[1m!!\x1b[0m %s\n", err.Error())
}
