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
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

func IsRegularFile(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

//Returns true if the file at firstPath is newer than the file at secondPath.
//Panics on error. (Compare implementation of walkRepo.)
func IsNewerThan(path1, path2 string) bool {
	info1, err := os.Stat(path1)
	if err != nil {
		panic(err.Error())
	}
	info2, err := os.Stat(path2)
	if err != nil {
		panic(err.Error())
	}
	return info1.ModTime().After(info2.ModTime())
}

//Panics on error. (Compare implementation of walkRepo.)
func CopyFile(fromPath, toPath string) {
	if err := copyFileImpl(fromPath, toPath); err != nil {
		panic(fmt.Sprintf("Cannot copy %s to %s: %s", fromPath, toPath, err.Error()))
	}
}

func copyFileImpl(fromPath, toPath string) error {
	//copy contents
	data, err := ioutil.ReadFile(fromPath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(toPath, data, 0600)
	if err != nil {
		return err
	}

	return ApplyFilePermissions(fromPath, toPath)
}

func ApplyFilePermissions(fromPath, toPath string) error {
	//apply permissions, ownership, modification date from source file to target file
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
	err = os.Chtimes(toPath, info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}

	return nil
}
