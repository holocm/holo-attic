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

package impl

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

//DumpRpm dumps RPM packages.
func DumpRpm(data []byte) (string, error) {
	//We don't have a library for the RPM format, and unfortunately, it's an utter mess.
	//The main reference that I used (apart from sample RPMs from Fedora, Mageia, and Suse)
	//is <http://www.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html>.
	reader := bytes.NewReader(data)

	//decode the various header structures
	leadDump, err := dumpRpmLead(reader)
	if err != nil {
		return "", err
	}
	signatureDump, err := dumpRpmHeader(reader, "signature", true)
	if err != nil {
		return "", err
	}
	headerDump, err := dumpRpmHeader(reader, "header", false)
	if err != nil {
		return "", err
	}

	//decode payload
	payloadData, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	payloadDump, err := RecognizeAndDump(payloadData)
	if err != nil {
		return "", err
	}

	return "RPM package\n" + Indent(leadDump) + Indent(signatureDump) + Indent(headerDump) + Indent(">> payload: "+payloadDump), nil
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
	return ">> lead section:\n" + Indent(strings.Join(lines, "\n")), nil
}

//IndexEntry represents an entry in the index of an RPM header.
type IndexEntry struct {
	Tag    uint32 //defines the semantics of the value in this field
	Type   uint32 //data type
	Offset uint32 //relative to the beginning of the store
	Count  uint32 //number of data items in this field
}

func dumpRpmHeader(reader io.Reader, sectionIdent string, readAligned bool) (string, error) {
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

	if readAligned {
		//next structure in reader is aligned to 4-byte boundary -- skip over padding
		_, err = io.ReadFull(reader, make([]byte, 4-header.DataSize%4))
		if err != nil {
			return "", err
		}
	}

	//decode and dump entries
	lines := make([]string, 0, len(indexEntries)) //lower estimate only
	for _, entry := range indexEntries {
		//seek to start of entry
		_, err := bufferedReader.Seek(int64(entry.Offset), 0)
		if err != nil {
			return "", err
		}

		var sublines []string
		if entry.Type == 7 {
			//for entry.Type = 7 (BIN), entry.Count is the number of bytes to be read
			data := make([]byte, entry.Count)
			_, err = io.ReadFull(bufferedReader, data)
			if err != nil {
				return "", err
			}
			sublines = []string{hex.Dump(data)}
		} else {
			//for all other types, entry.Count tells the number of records to read
			sublines = make([]string, 0, entry.Count)
			for idx := uint32(0); idx < entry.Count; idx++ {
				repr, err := decodeIndexEntry(entry.Type, bufferedReader)
				if err != nil {
					return "", err
				}
				sublines = append(sublines, repr)
			}
		}

		line := fmt.Sprintf("tag %d: length %d\n", entry.Tag, entry.Count)
		lines = append(lines, line+strings.TrimSuffix(Indent(strings.Join(sublines, "\n")), "\n"))
	}

	return identifier + Indent(strings.Join(lines, "\n")), nil
}

func decodeIndexEntry(dataType uint32, reader io.Reader) (string, error) {
	//check data type
	switch dataType {
	case 0: //NULL
		return "null", nil
	case 1: //CHAR
		var value uint8
		err := binary.Read(reader, binary.BigEndian, &value)
		return fmt.Sprintf("char: %c", rune(value)), err
	case 2: //INT8
		var value int8
		err := binary.Read(reader, binary.BigEndian, &value)
		return fmt.Sprintf("int8: %d", value), err
	case 3: //INT16
		var value int16
		err := binary.Read(reader, binary.BigEndian, &value)
		return fmt.Sprintf("int16: %d", value), err
	case 4: //INT32
		var value int32
		err := binary.Read(reader, binary.BigEndian, &value)
		return fmt.Sprintf("int32: %d", value), err
	case 5: //INT64
		var value int64
		err := binary.Read(reader, binary.BigEndian, &value)
		return fmt.Sprintf("int64: %d", value), err
	case 7: //BIN
		panic("Cannot be reached")
	case 6, 8: //STRING or STRING_ARRAY (not different at this point)
		//read NUL-terminated string (byte-wise)
		var result []byte
		buffer := make([]byte, 1)
		for {
			_, err := reader.Read(buffer)
			if err != nil {
				return "", nil
			}
			if buffer[0] == 0 {
				break
			} else {
				result = append(result, buffer[0])
			}
		}
		return fmt.Sprintf("string: %s", string(result)), nil
	default:
		return fmt.Sprintf("don't know how to decode data type %d", dataType), nil
	}
}
