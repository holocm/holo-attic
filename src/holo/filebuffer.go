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
	"io/ioutil"
	"os"
	"path/filepath"
)

//This type represents the contents of a file. It is used in holo.Apply() as
//an intermediary product of application steps.
type FileBuffer struct {
	//set only for regular files
	Contents []byte
	//set only for symlinks
	SymlinkTarget string
	//used by ResolveSymlink (see doc over there)
	BasePath string
}

type FileBufferError string

func (e *FileBufferError) Error() string { return string(*e) }

//NewFileBuffer creates a FileBuffer object by reading the manageable file at
//the given path. The basePath is stored in the FileBuffer for use in
//holo.FileBuffer.ResolveSymlink().
func NewFileBuffer(path string, basePath string) (*FileBuffer, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	//a manageable file is either a symlink...
	if IsFileInfoASymbolicLink(info) {
		target, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}
		return &FileBuffer{
			Contents:      nil,
			SymlinkTarget: target,
			BasePath:      basePath,
		}, nil
	}

	//...or a regular file
	if info.Mode().IsRegular() {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &FileBuffer{
			Contents:      contents,
			SymlinkTarget: "",
			BasePath:      basePath,
		}, nil
	}

	//other types of files are not acceptable
	fberr := FileBufferError("not a manageable file")
	return nil, &os.PathError{
		Op:   "holo.NewFileBuffer",
		Path: path,
		Err:  &fberr,
	}
}

//NewFileBufferFromContents creates a file buffer containing the given byte
//array. The basePath is stored in the FileBuffer for use in
//holo.FileBuffer.ResolveSymlink().
func NewFileBufferFromContents(fileContents []byte, basePath string) *FileBuffer {
	return &FileBuffer{
		Contents:      fileContents,
		SymlinkTarget: "",
		BasePath:      basePath,
	}
}

func (fb *FileBuffer) Write(path string) error {
	//(check that we're not attempting to overwrite unmanageable files
	info, err := os.Lstat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			//abort because the target location is not accessible
			return err
		}
	} else {
		if !(info.Mode().IsRegular() || IsFileInfoASymbolicLink(info)) {
			fberr := FileBufferError("target exists and is not a manageable file")
			return &os.PathError{
				Op:   "holo.FileBuffer.Write",
				Path: path,
				Err:  &fberr,
			}
		}
	}

	//before writing to the target, remove what was there before
	err = os.Remove(path)
	if err != nil {
		return err
	}

	if fb.Contents != nil {
		//a manageable file is either a regular file...
		return ioutil.WriteFile(path, fb.Contents, 600)
	} else {
		//...or a symlink
		return os.Symlink(fb.SymlinkTarget, path)
	}
}

//If the given FileBuffer contains a symlink, ResolveSymlink resolves it and
//returns a new FileBuffer containing the contents of the symlink target. This
//operation is used by application strategies that require text input.
//
//It uses the FileBuffer's BasePath to resolve relative symlinks. Since
//file buffers are usually written to the target path of a `holo apply`
//operation, the BasePath is most likely the target path.
func (fb *FileBuffer) ResolveSymlink() (*FileBuffer, error) {
	//if the buffer has contents already, we can use that
	if fb.Contents != nil {
		return fb, nil
	}

	//if the symlink target is relative, resolve it
	target := fb.SymlinkTarget
	if !filepath.IsAbs(target) {
		baseDir := filepath.Dir(fb.BasePath)
		target = filepath.Join(baseDir, target)
	}

	//read the contents of the target file (NOTE: It's tempting to just use
	//NewFileBuffer here, but that might give us another FileBuffer with a
	//symlink in it, and this time the symlink target might not resolve
	//correctly against the original BasePath. So we explicitly read the file.)
	contents, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	} else {
		return NewFileBufferFromContents(contents, fb.BasePath), nil
	}
}
