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
	"path/filepath"
	"strings"
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
	str := fmt.Sprintf("%s-%d", pkg.Version, pkg.Release)
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

	//prepare a directory into which to assemble the metadata files for control.tar.gz
	controlTar, err := buildControlTar(pkg, rootPath, buildReproducibly)
	if err != nil {
		return nil, err
	}

	//build ar archive
	return buildArArchive([]arArchiveEntry{
		arArchiveEntry{"debian-binary", []byte("2.0\n")},
		arArchiveEntry{"control.tar.gz", controlTar},
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

func buildControlTar(pkg *common.Package, rootPath string, buildReproducibly bool) ([]byte, error) {
	//prepare a directory into which to put all these files
	controlPath := filepath.Join(rootPath, ".control")
	err := os.MkdirAll(controlPath, 0755)
	if err != nil {
		return nil, err
	}

	//place all the required files in there
	err = writeControlFile(pkg, rootPath, controlPath, buildReproducibly)
	if err != nil {
		return nil, err
	}

	//TODO: add files "conffiles" and "md5sums"

	//compress directory
	cmd := exec.Command(
		//using standardized language settings...
		"env", "LANG=C",
		//...generate a .tar.gz archive...
		"tar", "czf", "-",
		//...of the working directory (== controlPath)
		".",
	)
	cmd.Dir = controlPath
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func writeControlFile(pkg *common.Package, rootPath, controlPath string, buildReproducibly bool) error {
	//reference for this file:
	//https://www.debian.org/doc/debian-policy/ch-controlfields.html#s-binarycontrolfiles
	contents := fmt.Sprintf("Package: %s\n", pkg.Name)
	contents += fmt.Sprintf("Version: %s\n", fullVersionString(pkg))
	contents += "Architecture: all\n"
	contents += "Maintainer: Unknown <unknown@example.org>\n"                             //TODO This needs to be a field in the package definition, and its existence must be validated for Debian packages.
	contents += fmt.Sprintf("Installed-Size: %d\n", int(pkg.InstalledSizeInBytes()/1024)) // convert bytes to KiB
	contents += "Section: misc\n"
	contents += "Priority: optional\n"

	//compile relations
	rels, err := compilePackageRelations("Depends", pkg.Requires)
	if err != nil {
		return err
	}
	contents += rels

	rels, err = compilePackageRelations("Provides", pkg.Provides)
	if err != nil {
		return err
	}
	contents += rels

	rels, err = compilePackageRelations("Conflicts", pkg.Conflicts)
	if err != nil {
		return err
	}
	contents += rels

	rels, err = compilePackageRelations("Replaces", pkg.Replaces)
	if err != nil {
		return err
	}
	contents += rels

	//we have only one description field, which we use both as the synopsis and the extended description
	desc := strings.TrimSpace(strings.Replace(pkg.Description, "\n", " ", -1))
	if desc == "" {
		desc = strings.TrimSpace(pkg.Name) //description field is strictly required
	}
	contents += fmt.Sprintf("Description: %s\n %s\n", desc, desc)

	return common.WriteFile(filepath.Join(controlPath, "control"), []byte(contents), 0644, buildReproducibly)
}

func compilePackageRelations(relType string, rels []common.PackageRelation) (string, error) {
	if len(rels) == 0 {
		return "", nil
	}

	lines := make([]string, 0, len(rels))
	//foreach related package...
	for _, rel := range rels {
		line := fmt.Sprintf("%s: %s", relType, rel.RelatedPackage)

		//...compile constraints into a list like ">= 2.4, << 3.0" (operators "<" and ">" become "<<" and ">>" here)
		if len(rel.Constraints) > 0 {
			if relType == "Provides" {
				return "", fmt.Errorf("version constraints on \"Provides: %s\" are not allowed for Debian packages", rel.RelatedPackage)
			}
			constraints := make([]string, 0, len(rel.Constraints))
			for _, c := range rel.Constraints {
				operator := c.Relation
				if operator == "<" {
					operator = "<<"
				}
				if operator == ">" {
					operator = ">>"
				}
				constraints = append(constraints, fmt.Sprintf("%s %s", operator, c.Version))
			}
			line += fmt.Sprintf(" (%s)", strings.Join(constraints, ", "))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n", nil
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
