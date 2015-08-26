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

package holo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Apply(file ConfigFile, withForce bool) {
	//when anything of the following panics, display the error and continue
	//with the next file
	defer func() {
		if message := recover(); message != nil {
			PrintError(message.(string))
		}
	}()

	//determine the related paths
	repoFile := file.RepoFile()
	repoPath := repoFile.Path()
	targetPath := file.TargetPath()
	backupPath := file.BackupPath()
	pacnewPath := targetPath + ".pacnew"

	//application strategy is determined by the file suffix (TODO: make this mess object-oriented)
	var applicationStrategy func(string, string, string)
	if repoFile.ApplicationStrategy() == "passthru" {
		//repoPath ends in ".holoscript" -> the repo file is a script that
		//converts the backup file into the target file
		applicationStrategy = applyProgram
	} else {
		//repoPath does not have special suffix -> the repo file is applied by
		//copying it to the target location
		applicationStrategy = applyCopy
	}

	//step 1: will only install files from repo if there is a corresponding
	//regular file in the target location (that file comes from the application
	//package, the repo file from the holo metapackage)
	PrintInfo("Working on \x1b[1m%s\x1b[0m", targetPath)
	if !IsManageableFile(targetPath) {
		panic(fmt.Sprintf("%s is not a regular file", targetPath))
	}

	//step 2: we know that a file exists at installPath; if we don't have a
	//backup of the original file, the file at installPath *is* the original
	//file which we have to backup now
	if !IsManageableFile(backupPath) {
		PrintInfo("  store at %s", backupPath)

		backupDir := filepath.Dir(backupPath)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			panic(fmt.Sprintf("Cannot create directory %s: %s", backupDir, err.Error()))
		}

		CopyFile(targetPath, backupPath)
	}

	//step 2.5: if a .pacnew file exists next to the targetPath, the base
	//package was updated and the .pacnew is the newer version of the original
	//config file; move it to the backup location
	if IsManageableFile(pacnewPath) {
		PrintInfo("    update %s -> %s", pacnewPath, backupPath)
		CopyFile(pacnewPath, backupPath)
		_ = os.Remove(pacnewPath) //this can fail silently
	}

	//step 3: overwrite targetPath with repoPath *if* the version at targetPath
	//is the one installed by the package (which can be found at backupPath);
	//complain if the user made any changes to config files governed by holo
	//(this check is overridden by the --force option)
	if !withForce && IsNewerThan(targetPath, backupPath) {
		panic(fmt.Sprintf("  skipped: target file has been modified by user (use --force to overwrite)"))
	}
	PrintInfo("%10s %s", repoFile.ApplicationStrategy(), repoPath)
	applicationStrategy(repoPath, backupPath, targetPath)

	//step 4: copy permissions/timestamps from backup file to target file, in order to
	//be able to detect manual modifications in the next holo-apply run
	ApplyFilePermissions(backupPath, targetPath)
}

func applyCopy(repoPath, backupPath, targetPath string) {
	CopyFile(repoPath, targetPath)
}

func applyProgram(repoPath, backupPath, targetPath string) {
	//apply repoPath by executing it in the form
	//$ exec repoPath < backupPath > targetPath
	cmd := exec.Command(repoPath)

	//prepare standard input
	var err error
	cmd.Stdin, err = os.Open(backupPath)
	if err != nil {
		panic(err.Error())
	}

	//run command, fetch result file into buffer (not into the targetPath
	//directly, in order not to corrupt the file there if the script run fails)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if stderr.Len() > 0 {
		PrintWarning("execution of %s produced error output:", repoPath)
		stderrLines := strings.Split(strings.Trim(stderr.String(), "\n"), "\n")
		for _, stderrLine := range stderrLines {
			PrintWarning("    %s", stderrLine)
		}
	}
	if err != nil {
		panic(err.Error())
	}

	//write result file and apply permissions from backup path
	err = ioutil.WriteFile(targetPath, stdout.Bytes(), 600)
	if err != nil {
		panic(err.Error())
	}
}
