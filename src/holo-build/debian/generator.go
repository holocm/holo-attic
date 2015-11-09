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

package debian

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"../common"
)

//Generator is the common.Generator for Debian packages.
type Generator struct{}

//RecommendedFileName implements the common.Generator interface.
func (g *Generator) RecommendedFileName(pkg *common.Package) string {
	//this is called after Build(), so we can assume that package name,
	//version, etc. were already validated
	return fmt.Sprintf("%s_%s_any.deb", pkg.Name, fullVersionString(pkg))
}

func fullVersionString(pkg *common.Package) string {
	str := fmt.Sprintf("%s_%d", pkg.Version, pkg.Release)
	if pkg.Epoch > 0 {
		str = fmt.Sprintf("%d:%s", pkg.Epoch, str)
	}
	return str
}

type arArchiveEntry struct {
	Name string
	Data []byte
}

//Build implements the common.Generator interface.
func (g *Generator) Build(pkg *common.Package, rootPath string, buildReproducibly bool) ([]byte, error) {
	//validate package
	err := validatePackage(pkg)
	if err != nil {
		return nil, err
	}

	//compress data.tar.xz
	dataTar, err := buildDataTar(rootPath)
	if err != nil {
		return nil, err
	}

	//TODO: build control.tar.gz

	//build ar archive
	return buildArArchive([]arArchiveEntry{
		arArchiveEntry{"debian-binary", []byte("2.0\n")},
		//arArchiveEntry{"control.tar.gz", controlTar},
		arArchiveEntry{"data.tar.xz", dataTar},
	})
}

func buildDataTar(rootPath string) ([]byte, error) {
	cmd := exec.Command(
		//using standardized language settings...
		"env", "LANG=C",
		//...generate a .tar.xz archive...
		"tar", "cJf", "-",
		//...of the working directory (== rootPath)
		".",
	)
	cmd.Dir = rootPath
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func buildArArchive(entries []arArchiveEntry) ([]byte, error) {
	//we only need a very small subset of the ar archive format, so we can
	//directly construct it without requiring an extra library
	buf := bytes.NewBuffer([]byte("!<arch>\n"))

	//most fields are static
	now := time.Now().Unix()
	headerFormat := "%-16s"
	headerFormat += fmt.Sprintf("%-12d", now) //modification time = now
	headerFormat += "0     "                  //owner ID = root
	headerFormat += "0     "                  //group ID = root
	headerFormat += "100644  "                //file mode = regular file, rw-r--r--
	headerFormat += "%-10d"                   //file size in bytes
	headerFormat += "\x60\n"                  //magic header separator

	for _, entry := range entries {
		fmt.Fprintf(buf, headerFormat, entry.Name, len(entry.Data))
		buf.Write(entry.Data)
		//pad data to 2-byte boundary
		if len(entry.Data)%2 == 1 {
			buf.Write([]byte{'\n'})
		}
	}

	return buf.Bytes(), nil
}
