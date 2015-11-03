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

// #include <locale.h>
import "C"
import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
)

//This program is used by the holo-build tests to extract generated packages and render
//a textual representation of the package, including the compression and
//archive formats used and all file metadata contained within the archives.
//The program is called like
//
//    ./build/dump-package < $package
//
//And renders output like this:
//
//    $ tar cJf foo.tar.xz foo/
//    $ ./build/dump-package < foo.tar.xz
//    XZ-compressed data
//        POSIX tar archive
//            >> foo/ is directory (mode: 0755, owner: 1000, group: 1000)
//            >> foo/bar is regular file (mode: 0600, owner: 1000, group: 1000), content: data as shown below
//                Hello World!
//            >> foo/baz is symlink to bar
//
//The program is deliberately written very generically so as to make it easy to
//add support for new package formats in the future (when holo-build gains new
//generators).

func main() {
	//Holo requires a neutral locale, esp. for deterministic sorting of file paths
	lcAll := C.int(0)
	C.setlocale(lcAll, C.CString("C"))

	//read the input from stdin
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	//recognize the input, while deconstructing it recursively
	dump, err := recognizeAndDump(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(dump)
}

//general-purpose helper function for pretty-printing of nested data
func indent(dump string) string {
	//indent the first line and all subsequent lines except for the trailing newline
	//(and also ensure a trailing newline, which means that in total we can
	//trim the trailing newline at the start, and put it back at the end)
	dump = strings.TrimSuffix(dump, "\n")
	indent := "    "
	dump = indent + strings.Replace(dump, "\n", "\n"+indent, -1)
	return dump + "\n"
}

func recognizeAndDump(data []byte) (string, error) {
	//Thanks to https://stackoverflow.com/a/19127748/334761 for
	//listing all the magic numbers of the usual compression formats.

	//is it GZip-compressed?
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b, 0x08}) {
		return dumpGZ(data)
	}
	//is it BZip2-compressed?
	if bytes.HasPrefix(data, []byte{0x42, 0x5a, 0x68}) {
		return dumpBZ2(data)
	}
	//is it XZ-compressed?
	if bytes.HasPrefix(data, []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}) {
		return dumpXZ(data)
	}
	//is it a POSIX tar archive?
	if len(data) >= 512 && bytes.Equal(data[257:262], []byte("ustar")) {
		return dumpTar(data)
	}

	return "data as shown below\n" + indent(string(data)), nil
}

func dumpGZ(data []byte) (string, error) {
	//use "compress/gzip" package to decompress the data
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	data2, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	//`data2` now contains the decompressed data
	dump, err := recognizeAndDump(data2)
	return "GZip-compressed data\n" + indent(dump), err
}

func dumpBZ2(data []byte) (string, error) {
	//use "compress/bzip2" package to decompress the data
	r := bzip2.NewReader(bytes.NewReader(data))
	data2, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	//`data2` now contains the decompressed data
	dump, err := recognizeAndDump(data2)
	return "BZip2-compressed data\n" + indent(dump), err
}

func dumpXZ(data []byte) (string, error) {
	//the Go stdlib does not have a compress/xz package, so use the command-line utility
	cmd := exec.Command("xz", "-d")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	//`output` now contains the decompressed data
	dump, err := recognizeAndDump(output)
	return "XZ-compressed data\n" + indent(dump), err
}

func dumpTar(data []byte) (string, error) {
	//use "archive/tar" package to read the tar archive
	tr := tar.NewReader(bytes.NewReader(data))
	dumps := make(map[string]string)
	var names []string

	//iterate through the entries in the archive
	for {
		//get next entry
		header, err := tr.Next()
		if err == io.EOF {
			break //end of archive
		}
		if err != nil {
			return "", err
		}
		info := header.FileInfo()

		//get contents of entry
		data, err := ioutil.ReadAll(tr)
		if err != nil {
			return "", err
		}

		//recognize entry type
		typeStr := ""
		switch info.Mode() & os.ModeType {
		case os.ModeDir:
			typeStr = "directory"
		case os.ModeSymlink:
			typeStr = "symlink"
		case 0:
			typeStr = "regular file"
		default:
			return "", fmt.Errorf("tar entry %s has unrecognized file mode (%o)", header.Name, info.Mode())
		}

		//compile metadata description
		str := fmt.Sprintf(">> %s is %s", header.Name, typeStr)

		if typeStr == "symlink" {
			str += fmt.Sprintf(" to %s", header.Linkname)
		} else {
			str += fmt.Sprintf(" (mode: %o, owner: %d, group: %d)",
				info.Mode()&os.ModePerm, header.Uid, header.Gid,
			)
		}

		//recognizeAndDump contents of regular files with indentation
		if typeStr == "regular file" {
			dump, err := recognizeAndDump(data)
			if err != nil {
				return "", err
			}

			str += ", content is " + dump
		} else {
			str += "\n"
		}

		names = append(names, header.Name)
		dumps[header.Name] = str
	}

	//dump entries ordered by name
	sort.Strings(names)
	dump := ""
	for _, name := range names {
		dump += dumps[name]
	}

	return "POSIX tar archive\n" + indent(dump), nil
}
