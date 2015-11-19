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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"../internal/ar"
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
	if len(data) == 0 {
		return "empty file", nil
	}

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
	//is it an mtree archive?
	if bytes.HasPrefix(data, []byte("#mtree")) {
		return dumpMtree(data)
	}
	//is it an ar archive?
	if bytes.HasPrefix(data, []byte("!<arch>\n")) {
		return dumpAr(data)
	}
	//is it an RPM package?
	if bytes.HasPrefix(data, []byte{0xed, 0xab, 0xee, 0xdb}) {
		return dumpRpm(data)
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

func dumpAr(data []byte) (string, error) {
	//use "github.com/blakesmith/ar" package to read the ar archive
	ar := ar.NewReader(bytes.NewReader(data))
	dumps := make(map[string]string)
	var names []string

	//iterate through the entries in the archive
	idx := -1
	for {
		idx++

		//get next entry
		header, err := ar.Next()
		if err == io.EOF {
			break //end of archive
		}
		if err != nil {
			return "", err
		}

		//get contents of entry
		data, err := ioutil.ReadAll(ar)
		if err != nil {
			return "", err
		}

		//our ar parser only works with a small subset of all the varieties of
		//ar files (large enough to handle Debian packages whose toplevel ar
		//packages contain just plain files with short names), so we assume
		//that everything that it reads without crashing is a regular file
		str := fmt.Sprintf(">> %s is regular file (mode: %o, owner: %d, group: %d)",
			header.Name, header.Mode, header.Uid, header.Gid,
		)

		//for Debian packages, we need to check that the file "debian-binary"
		//is the first entry
		if header.Name == "debian-binary" {
			str += fmt.Sprintf(" at archive position %d", idx)
		}

		//recognizeAndDump contents of regular files with indentation
		dump, err := recognizeAndDump(data)
		if err != nil {
			return "", err
		}
		str += ", content is " + dump

		names = append(names, header.Name)
		dumps[header.Name] = str
	}

	//dump entries ordered by name
	sort.Strings(names)
	dump := ""
	for _, name := range names {
		dump += dumps[name]
	}

	return "ar archive\n" + indent(dump), nil
}

func dumpMtree(data []byte) (string, error) {
	//We don't have a library for the mtree(5) format, but it's relatively simple.
	//NOTE: We don't support absolute paths ("mtree v2.0") and we don't track the cwd.
	//All we do is resolve duplicate entries and "/set" and "/unset" commands.
	lines := strings.Split(string(data), "\n")

	//go through each entry and resolve "/set"
	globalOpts := make(map[string]string)
	entries := make(map[string]map[string]string)

	for _, line := range lines {
		//ignore comments
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		//lines look like "name option option option"
		options := strings.Split(line, " ")
		name := options[0]
		options = options[1:]

		//parse options (option = "key=value")
		opts := make(map[string]string, len(options))
		for _, option := range options {
			pair := strings.SplitN(option, "=", 2)
			if len(pair) == 1 {
				opts[pair[0]] = ""
			} else {
				opts[pair[0]] = pair[1]
			}
		}

		//name can either be a special command or a filename
		switch name {
		case "/set":
			//set the opts globally
			for key, value := range opts {
				globalOpts[key] = value
			}
		case "/unset":
			//unset the opts globally
			for key := range opts {
				delete(globalOpts, key)
			}
		default:
			//create (if missing) an entry for this file and add the opts to it
			entry, ok := entries[name]
			if !ok {
				entry = make(map[string]string, len(opts)+len(globalOpts))
				//apply globalOpts
				for key, value := range globalOpts {
					entry[key] = value
				}
				entries[name] = entry
			}
			for key, value := range opts {
				entry[key] = value
			}
		}
	}

	//sort entries by name
	entryNames := make([]string, 0, len(entries))
	for name := range entries {
		entryNames = append(entryNames, name)
	}
	sort.Strings(entryNames)

	outputLines := make([]string, 0, len(entries))
	for _, name := range entryNames {
		//sort options for entry by key
		entry := entries[name]
		keys := make([]string, 0, len(entry))
		for key := range entry {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		options := ""
		for _, key := range keys {
			options += fmt.Sprintf(" %s=%s", key, entry[key])
		}

		outputLines = append(outputLines, ">> "+name+options)
	}

	return "mtree metadata archive\n" + indent(strings.Join(outputLines, "\n")), nil
}

func dumpRpm(data []byte) (string, error) {
	//We don't have a library for the RPM format, and unfortunately, it's an utter mess.
	//The main reference that I used (apart from sample RPMs from Fedora, Mageia, and Suse)
	//is <http://www.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html>.
	reader := bytes.NewReader(data)

	leadDump, err := dumpRpmLead(reader)
	if err != nil {
		return "", err
	}
	signatureDump, err := dumpRpmHeader(reader, "signature")
	if err != nil {
		return "", err
	}
	headerDump, err := dumpRpmHeader(reader, "header")
	if err != nil {
		return "", err
	}

	return "RPM package\n" + indent(leadDump) + indent(signatureDump) + indent(headerDump) + indent(">> TODO: header entries, payload"), nil
}

func dumpRpmLead(reader io.Reader) (string, error) {
	//read the lead (the initial fixed-size header)
	var lead struct {
		Magic         uint32
		MajorVersion  uint8
		MinorVersion  uint8
		Type          uint16
		Architecture  uint16
		Name          [66]byte
		OSNum         uint16
		SignatureType uint16
		Reserved      [16]byte
	}
	err := binary.Read(reader, binary.BigEndian, &lead)
	if err != nil {
		return "", err
	}

	lines := []string{
		fmt.Sprintf("RPM format version %d.%d", lead.MajorVersion, lead.MinorVersion),
		fmt.Sprintf("Type: %d (0 = binary, 1 = source)", lead.Type),
		fmt.Sprintf("Architecture: %d (0 = noarch, 1 = x86, ...)", lead.Architecture),
		//lead.Name is a NUL-terminated (and NUL-padded) string; trim all the NULs at the end
		fmt.Sprintf("Name: %s", strings.TrimRight(string(lead.Name[:]), "\x00")),
		fmt.Sprintf("Built for OS: %d (1 = Linux, ...)", lead.OSNum),
		fmt.Sprintf("Signature type: %d", lead.SignatureType),
	}
	return ">> lead section:\n" + indent(strings.Join(lines, "\n")), nil
}

func dumpRpmHeader(reader io.Reader, sectionIdent string) (string, error) {
	//the header has a header (I'm So Meta, Even This Acronym)
	var header struct {
		Magic      [3]byte
		Version    uint8
		Reserved   [4]byte
		EntryCount uint32 //supports 4 billion header entries... Now that's planning ahead! :)
		DataSize   uint32 //size of the store (i.e. the data section, everything after the index until the end of the header section)
	}
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		return "", err
	}
	if header.Magic != [3]byte{0x8e, 0xad, 0xe8} {
		return "", fmt.Errorf(
			"did not find RPM header structure header at expected position (saw 0x%s instead of 0x8eade8)",
			hex.EncodeToString(header.Magic[:]),
		)
	}
	identifier := fmt.Sprintf(">> %s section: format version %d, %d entries, %d bytes of data\n",
		sectionIdent, header.Version, header.EntryCount, header.DataSize,
	)

	//read index of fields
	type IndexEntry struct {
		Tag    uint32 //defines the semantics of the value in this field
		Type   uint32 //data type
		Offset uint32 //relative to the beginning of the store
		Count  uint32 //number of data items in this field
	}
	indexEntries := make([]IndexEntry, 0, header.EntryCount)
	for idx := uint32(0); idx < header.EntryCount; idx++ {
		var entry IndexEntry
		err := binary.Read(reader, binary.BigEndian, &entry)
		if err != nil {
			return "", err
		}
		indexEntries = append(indexEntries, entry)
	}

	//read remaining part of header (the data store) into a buffer for random access
	buffer := make([]byte, header.DataSize)
	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return "", err
	}
	bufferedReader := bytes.NewReader(buffer)

	//next structure in reader is aligned to 4-byte boundary -- skip over padding
	_, err = io.ReadFull(reader, make([]byte, 4-header.DataSize%4))
	if err != nil {
		return "", err
	}

	_ = bufferedReader

	//TODO: header entries need to be parsed; for now, just display the index entries for validation purposes
	lines := []string{}
	for _, entry := range indexEntries {
		lines = append(lines, fmt.Sprintf(
			"entry: tag %d, type %d, offset %d, count %d",
			entry.Tag, entry.Type, entry.Offset, entry.Count,
		))
	}
	return identifier + indent(strings.Join(lines, "\n")), nil
}
