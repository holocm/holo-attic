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

// The idea of holo is to rely as much on package management as possible for
// the whole system setup and maintenance process. This is achieved by using
// metapackages to define personal package selections for all systems or for
// certain types of systems.
//
// Metapackages go only 90% of the way, though. A prominent shortcoming is that
// the holo metapackages cannot install custom configuration files where the
// original packages already installed stock configuration files.
//
// Instead, the custom configuration file is placed under the /holo/repo
// directory, e.g. /holo/repo/etc/foobar.conf. This tool, holo-apply, is then
// run by the metapackage's post-install and post-update hook to place the
// custom configuration file at its designated position
// (e.g. /etc/foobar.conf), while simultaneously retaining a copy of the stock
// configuration file in /holo/backup (e.g. /holo/backup/etc/foobar.conf) for
// reference.

package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

var msgError = "31"
var msgInfo = "38"

func msg(color string, message string) {
	fmt.Printf("\x1b[%sm\x1b[1m[holo-apply]\x1b[0m %s\n", color, message)
}

func main() {
	//check that /holo/repo exists
	repoInfo, err := os.Lstat("/holo/repo")
	if err != nil {
		msg(msgError, fmt.Sprintf("Cannot open /holo/repo: %s", err.Error()))
		return
	}
	if !repoInfo.IsDir() {
		msg(msgError, "Cannot open /holo/repo: not a directory!")
		return
	}

	//do the work :)
	filepath.Walk("/holo/repo", walkRepo)
}

func walkRepo(repoPath string, repoInfo os.FileInfo, err error) (resultError error) {
	//skip over unaccessible stuff
	if err != nil {
		return err
	}
	//only look at files
	if !repoInfo.Mode().IsRegular() {
		return nil
	}

	//when anything of the following panics, display the error and continue
	//with the next file
	defer func() {
		if message := recover(); message != nil {
			msg(msgError, message.(string))
			resultError = nil
		}
	}()

	//determine the related paths
	relPath, _ := filepath.Rel("/holo/repo", repoPath)
	targetPath := filepath.Join("/", relPath)
	backupPath := filepath.Join("/holo/backup", relPath)
	pacnewPath := targetPath + ".pacnew"

	//step 1: will only install files from repo if there is a corresponding
	//regular file in the target location (that file comes from the application
	//package, the repo file from the holo metapackage)
	if !isRegularFile(targetPath) {
		panic(fmt.Sprintf("%s is not a regular file", targetPath))
	}

	//step 2: we know that a file exists at installPath; if we don't have a
	//backup of the original file, the file at installPath *is* the original
	//file which we have to backup now
	if !isRegularFile(backupPath) {
		msg(msgInfo, fmt.Sprintf("Saving %s in /holo/backup", targetPath))

		backupDir := filepath.Dir(backupPath)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			panic(fmt.Sprintf("Cannot create directory %s: %s", backupDir, err.Error()))
		}

		copyFile(targetPath, backupPath)
	}

	//step 3: overwrite targetPath with repoPath *if* the version at targetPath
	//is the one installed by the package (which can be found at backupPath);
	//complain if the user made any changes to config files governed by holo
	targetSha := sha256ForFile(targetPath)
	if targetSha != sha256ForFile(repoPath) {
		if targetSha != sha256ForFile(backupPath) {
			panic(fmt.Sprintf("Skipping %s: has been modified by user", targetPath))
		}
		msg(msgInfo, fmt.Sprintf("Installing %s", targetPath))
		copyFile(repoPath, targetPath)
	}

	//step 4: if a .pacnew file exists next to the targetPath, the base package
	//was updated and the .pacnew is the newer version of the original config
	//file; move it to the backup location
	if isRegularFile(pacnewPath) {
		msg(msgInfo, fmt.Sprintf("Saving %s in /holo/backup", pacnewPath))
		copyFile(pacnewPath, backupPath)
	}

	return nil
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

//Panics on error. (Compare implementation of walkRepo.)
func sha256ForFile(path string) [32]byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err.Error())
	}
	return sha256.Sum256(data)
}

//Panics on error. (Compare implementation of walkRepo.)
func copyFile(fromPath, toPath string) {
	if err := copyFileImpl(fromPath, toPath); err != nil {
		panic(fmt.Sprintf("Cannot copy %s to %s: %s", fromPath, toPath, err.Error()))
	}
}

func copyFileImpl(fromPath, toPath string) (result error) {
	//copy contents
	data, err := ioutil.ReadFile(fromPath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(toPath, data, 0600)
	if err != nil {
		return err
	}

	//apply permissions and ownership from source file to target file
	//NOTE: We cannot just pass the FileMode in WriteFile(), because its
	//FileMode argument is only applied when a new file is created, not when
	//an existing one is truncated.
	info, err := os.Stat(fromPath)
	if err != nil {
		return err
	}
	err = os.Chmod(toPath, info.Mode())
	if err != nil {
		return err
	}
	stat_t := info.Sys().(*syscall.Stat_t) // UGLY
	err = os.Chown(toPath, int(stat_t.Uid), int(stat_t.Gid))
	if err != nil {
		return err
	}

	return nil
}
