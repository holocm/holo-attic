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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"./holo"
)

func main() {
	//check that /holo/repo exists
	repoPath := holo.RepoDirectory()
	repoInfo, err := os.Lstat(repoPath)
	if err != nil {
		holo.PrintError("Cannot open %s: %s", repoPath, err.Error())
		return
	}
	if !repoInfo.IsDir() {
		holo.PrintError("Cannot open %s: not a directory!", repoPath)
		return
	}

	//do the work :)
	filepath.Walk(repoPath, walkRepo)
}

func walkRepo(repoPath string, repoInfo os.FileInfo, err error) (resultError error) {
	//skip over unaccessible stuff
	if err != nil {
		return err
	}
	//only look at files
	if !(repoInfo.Mode().IsRegular() || holo.IsFileInfoASymbolicLink(repoInfo)) {
		return nil
	}

	//when anything of the following panics, display the error and continue
	//with the next file
	defer func() {
		if message := recover(); message != nil {
			holo.PrintError(message.(string))
			resultError = nil
		}
	}()

	//application strategy is determined by the file suffix (TODO: make this mess object-oriented)
	var strategyName string
	var applicationStrategy func(string, string, string)
	if repoInfo.Mode().IsRegular() {
		switch {
		case strings.HasSuffix(repoPath, ".holoscript"):
			//repoPath ends in ".holoscript" -> the repo file is a script that
			//converts the backup file into the target file
			strategyName = "program"
			applicationStrategy = applyProgram
		default:
			//repoPath does not have special suffix -> the repo file is applied by
			//copying it to the target location
			strategyName = "copy"
			applicationStrategy = applyCopy
		}
	} else {
		//for symbolic links, always use the copy strategy
		strategyName = "copy"
		applicationStrategy = applyCopy
	}

	//determine the related paths
	configFile := holo.NewConfigFileFromRepoPath(repoPath)
	targetPath := configFile.TargetPath()
	backupPath := configFile.BackupPath()
	pacnewPath := targetPath + ".pacnew"

	//step 1: will only install files from repo if there is a corresponding
	//regular file in the target location (that file comes from the application
	//package, the repo file from the holo metapackage)
	if !holo.IsManageableFile(targetPath) {
		panic(fmt.Sprintf("%s is not a regular file", targetPath))
	}

	//step 2: we know that a file exists at installPath; if we don't have a
	//backup of the original file, the file at installPath *is* the original
	//file which we have to backup now
	if !holo.IsManageableFile(backupPath) {
		holo.PrintInfo("Saving %s in %s", targetPath, holo.BackupDirectory())

		backupDir := filepath.Dir(backupPath)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			panic(fmt.Sprintf("Cannot create directory %s: %s", backupDir, err.Error()))
		}

		holo.CopyFile(targetPath, backupPath)
	}

	//step 2.5: if a .pacnew file exists next to the targetPath, the base
	//package was updated and the .pacnew is the newer version of the original
	//config file; move it to the backup location
	if holo.IsManageableFile(pacnewPath) {
		holo.PrintInfo("Saving %s in %s", pacnewPath, holo.BackupDirectory())
		holo.CopyFile(pacnewPath, backupPath)
		_ = os.Remove(pacnewPath) //this can fail silently
	}

	//step 3: overwrite targetPath with repoPath *if* the version at targetPath
	//is the one installed by the package (which can be found at backupPath);
	//complain if the user made any changes to config files governed by holo
	if holo.IsNewerThan(targetPath, backupPath) {
		//NOTE: this check works because holo.CopyFile() copies the mtime
		panic(fmt.Sprintf("Skipping %s: has been modified by user (application strategy: %s)", targetPath, strategyName))
	}
	holo.PrintInfo("Installing %s with application strategy: %s", targetPath, strategyName)
	applicationStrategy(repoPath, backupPath, targetPath)

	//step 4: copy permissions/timestamps from backup file to target file, in order to
	//be able to detect manual modifications in the next holo-apply run
	holo.ApplyFilePermissions(backupPath, targetPath)

	return nil
}

func applyCopy(repoPath, backupPath, targetPath string) {
	holo.CopyFile(repoPath, targetPath)
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
		holo.PrintWarning("execution of %s produced error output:", repoPath)
		stderrLines := strings.Split(strings.Trim(stderr.String(), "\n"), "\n")
		for _, stderrLine := range stderrLines {
			holo.PrintWarning("    %s", stderrLine)
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
