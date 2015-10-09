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
	"strings"

	"./common"
	"./entities"
	"./files"
)

//Note: This line is parsed by the Makefile to get the version string. If you
//change the format, adjust the Makefile too.
var version = "v0.6.1"
var codename = "Providence"

func main() {
	//a command word must be given as first argument
	if len(os.Args) < 2 {
		commandHelp()
		return
	}

	//check that it is a known command word
	var command func(common.Entities)
	switch os.Args[1] {
	case "apply":
		command = commandApply
	case "diff":
		command = commandDiff
	case "scan":
		command = commandScan
	case "version", "--version":
		fmt.Printf("%s \"%s\"\n", version, codename)
		return
	default:
		commandHelp()
		return
	}

	//scan the repo
	fileEntities := files.ScanRepo()
	if fileEntities == nil {
		//some fatal error occurred while scanning the repo - it was already
		//reported, so just exit
		return
	}

	//scan for entity definitions
	entities := entities.Scan()
	if entities == nil {
		//some fatal error occurred while scanning the repo - it was already
		//reported, so just exit
		return
	}

	//execute command
	command(append(fileEntities, entities...))
}

func commandHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <operation> [...]\nOperations:\n", program)
	fmt.Printf("    %s apply [-f|--force] [target(s)]\n", program)
	fmt.Printf("    %s diff [file(s)]\n", program)
	fmt.Printf("    %s scan [-s|--short]\n", program)
	fmt.Printf("\nSee `man 8 holo` for details.\n")
}

func commandApply(entities common.Entities) {
	//parse arguments after "holo apply" (either files or "--force")
	withForce := false
	withTargets := false
	targets := make(map[string]bool)

	args := os.Args[2:]
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-f", "--force":
				withForce = true
			default:
				fmt.Println("Unrecognized option: " + arg)
				return
			}
		} else {
			targets[arg] = true
			withTargets = true
		}
	}

	//apply all declared entities (or only some if the args contain a limited subset)
	for _, entity := range entities {
		if !withTargets || targets[entity.EntityID()] {
			entity.Apply(withForce)
		}
	}
}

func commandScan(entities common.Entities) {
	//check args
	args := os.Args[2:]
	isShort := false
	for _, arg := range args {
		//"--short" shows only the target names, not the strategy
		switch arg {
		case "-s", "--short":
			isShort = true
		default:
			fmt.Println("Unrecognized argument: " + arg)
			return
		}
	}

	//report declared entities
	for _, entity := range entities {
		if isShort {
			fmt.Println(entity.EntityID())
		} else {
			entity.Report().Print()
		}
	}
}

func commandDiff(entities common.Entities) {
	//parse arguments after "holo diff" (targets)
	withTargets := false
	targets := make(map[string]bool)
	args := os.Args[2:]
	for _, arg := range args {
		targets[arg] = true
		withTargets = true
	}

	//apply all declared entities (or only some if the args contain a limited subset)
	for _, entity := range entities {
		if !withTargets || targets[entity.EntityID()] {
			//ignore entities that are not files (TODO: allow diffs for users/groups, too)
			if target, ok := entity.(*files.TargetFile); ok {
				output, err := target.RenderDiff()
				if err != nil {
					report := common.Report{Action: "diff", Target: target.EntityID()}
					report.AddError(err.Error())
					report.Print()
				}
				os.Stdout.Write(output)
			}
		}
	}
}
