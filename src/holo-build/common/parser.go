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

package common

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"../../internal/toml"
)

//PackageDefinition only needs a nice exported name for the TOML parser to
//produce more meaningful error messages on malformed input data.
type PackageDefinition struct {
	Package   PackageSection
	File      []FileSection
	Directory []DirectorySection
	Symlink   []SymlinkSection
}

//PackageSection only needs a nice exported name for the TOML parser to produce
//more meaningful error messages on malformed input data.
type PackageSection struct {
	Name          string
	Version       string
	Description   string
	Requires      []string
	Provides      []string
	Conflicts     []string
	Replaces      []string
	SetupScript   string
	CleanupScript string
}

//FileSection only needs a nice exported name for the TOML parser to produce
//more meaningful error messages on malformed input data.
type FileSection struct {
	Path        string
	Content     string
	ContentFrom string
	Mode        string      //TOML does not support octal number literals, so we have to write: mode = "0666"
	Owner       interface{} //either string (name) or integer (ID)
	Group       interface{} //same
	//NOTE: We could use custom types implementing TextUnmarshaler for Mode,
	//Owner and Group, but then toml.Decode would accept any primitive type.
	//But for Mode, we need the type enforcement to prevent the "mode = 0666"
	//error (which would be 666 in decimal = something else in octal). And for
	//Owner and Group, we need to distinguish IDs from names using the type.
}

//DirectorySection only needs a nice exported name for the TOML parser to
//produce more meaningful error messages on malformed input data.
type DirectorySection struct {
	Path  string
	Mode  string      //see above
	Owner interface{} //see above
	Group interface{} //see above
}

//SymlinkSection only needs a nice exported name for the TOML parser to produce
//more meaningful error messages on malformed input data.
type SymlinkSection struct {
	Path   string
	Target string
}

type errorCollector struct {
	errors []error
}

func (c *errorCollector) add(err error) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

func (c *errorCollector) addf(format string, args ...interface{}) {
	if len(args) > 0 {
		c.errors = append(c.errors, fmt.Errorf(format, args...))
	} else {
		c.errors = append(c.errors, errors.New(format))
	}
}

//ParsePackageDefinition parses a package definition from the given input.
//The operation is successful if the returned []error is nil or empty.
func ParsePackageDefinition(input io.Reader) (*Package, []error) {
	//read from input
	blob, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, []error{err}
	}
	var p PackageDefinition
	_, err = toml.Decode(string(blob), &p)
	if err != nil {
		return nil, []error{err}
	}

	//restructure the parsed data into a common.Package struct
	fsEntryCount := len(p.Directory) + len(p.File) + len(p.Symlink)
	pkg := Package{
		Name:          strings.TrimSpace(p.Package.Name),
		Version:       strings.TrimSpace(p.Package.Version),
		Description:   strings.TrimSpace(p.Package.Description),
		SetupScript:   strings.TrimSpace(p.Package.SetupScript),
		CleanupScript: strings.TrimSpace(p.Package.CleanupScript),
		FSEntries:     make([]FSEntry, 0, fsEntryCount),
	}

	//do some basic validation on the package name and version since we're
	//going to use these to construct a path
	ec := &errorCollector{}
	if pkg.Name == "" {
		ec.addf("Missing package name", pkg.Name)
	}
	if strings.ContainsAny(pkg.Name, "/\r\n") {
		ec.addf("Invalid package name \"%s\" (may not contain slashes or newlines)", pkg.Name)
	}
	if pkg.Version == "" {
		ec.addf("Missing package version", pkg.Name)
	}
	if strings.ContainsAny(pkg.Version, "/\r\n") {
		ec.addf("Invalid package version \"%s\" (may not contain slashes or newlines)", pkg.Version)
	}
	if strings.ContainsAny(pkg.Description, "\r\n") {
		ec.addf("Invalid package description \"%s\" (may not contain newlines)", pkg.Name)
	}

	//parse relations to other packages
	pkg.Requires = parseRelatedPackages(p.Package.Requires, ec)
	pkg.Provides = parseRelatedPackages(p.Package.Provides, ec)
	pkg.Conflicts = parseRelatedPackages(p.Package.Conflicts, ec)
	pkg.Replaces = parseRelatedPackages(p.Package.Replaces, ec)

	//parse and validate FS entries
	wasPathSeen := make(map[string]bool, fsEntryCount)

	for idx, dirSection := range p.Directory {
		path := dirSection.Path
		if !validatePath(path, &wasPathSeen, ec, "directory", idx) {
			continue
		}

		entryDesc := fmt.Sprintf("directory \"%s\"", path)
		pkg.FSEntries = append(pkg.FSEntries, FSEntry{
			Type:  FSEntryTypeDirectory,
			Path:  path,
			Mode:  parseFileMode(dirSection.Mode, 0755, ec, entryDesc),
			Owner: parseUserOrGroupRef(dirSection.Owner, ec, entryDesc),
			Group: parseUserOrGroupRef(dirSection.Group, ec, entryDesc),
		})
	}

	for idx, fileSection := range p.File {
		path := fileSection.Path
		if !validatePath(path, &wasPathSeen, ec, "file", idx) {
			continue
		}

		entryDesc := fmt.Sprintf("file \"%s\"", path)
		pkg.FSEntries = append(pkg.FSEntries, FSEntry{
			Type:    FSEntryTypeRegular,
			Path:    path,
			Content: parseFileContent(fileSection.Content, fileSection.ContentFrom, ec, entryDesc),
			Mode:    parseFileMode(fileSection.Mode, 0755, ec, entryDesc),
			Owner:   parseUserOrGroupRef(fileSection.Owner, ec, entryDesc),
			Group:   parseUserOrGroupRef(fileSection.Group, ec, entryDesc),
		})
	}

	for idx, symlinkSection := range p.Symlink {
		path := symlinkSection.Path
		if !validatePath(path, &wasPathSeen, ec, "symlink", idx) {
			continue
		}

		if symlinkSection.Target == "" {
			ec.addf("symlink \"%s\" is invalid: missing target", path)
		}

		pkg.FSEntries = append(pkg.FSEntries, FSEntry{
			Type:    FSEntryTypeSymlink,
			Path:    path,
			Content: symlinkSection.Target,
		})
	}

	return &pkg, ec.errors
}

var relatedPackageRx = regexp.MustCompile(`^([^\s<=>]+)\s*(?:(<=?|>=?|=)\s*(\S+))?$`)

func parseRelatedPackages(specs []string, ec *errorCollector) []PackageRelation {
	rels := make([]PackageRelation, 0, len(specs))
	idxByName := make(map[string]int, len(specs))

	for _, spec := range specs {
		//check format of spec
		match := relatedPackageRx.FindStringSubmatch(spec)
		if match == nil {
			ec.addf("Invalid package reference: \"%s\"", spec)
			continue
		}

		//do we have a relation to this package already?
		name := match[1]
		idx, exists := idxByName[name]
		if !exists {
			//no, add a new one and remember it for later additional constraints
			idx = len(rels)
			idxByName[name] = idx
			rels = append(rels, PackageRelation{RelatedPackage: name})
		}

		//add version constraint if one was specified
		if match[2] != "" {
			constraint := VersionConstraint{Relation: match[2], Version: match[3]}
			rels[idx].Constraints = append(rels[idx].Constraints, constraint)
		}
	}

	return rels
}

//path is the path to be validated.
//wasPathSeen tracks usage of paths to detect duplicate entries.
//ec collects errors.
//entryType and entryIdx are
func validatePath(path string, wasPathSeen *map[string]bool, ec *errorCollector, entryType string, entryIdx int) bool {
	if path == "" {
		ec.addf("%s %d is invalid: missing \"path\" attribute", entryType, entryIdx)
		return false
	}
	if !strings.HasPrefix(path, "/") {
		ec.addf("%s \"%s\" is invalid: must be an absolute path", entryType, path)
		return false
	}
	if strings.HasSuffix(path, "/") {
		ec.addf("%s \"%s\" is invalid: trailing slash(es)", entryType, path)
		return false
	}
	if (*wasPathSeen)[path] {
		ec.addf("multiple entries for path \"%s\"", path)
		return false
	}
	(*wasPathSeen)[path] = true
	return true
}

func parseFileMode(modeStr string, defaultMode os.FileMode, ec *errorCollector, entryDesc string) os.FileMode {
	//default value
	if modeStr == "" {
		return defaultMode
	}

	//parse modeStr as uint in base 8 to uint32 (== os.FileMode)
	value, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		ec.addf("%s is invalid: cannot parse mode \"%s\" (%s)", entryDesc, modeStr, err.Error())
	}
	return os.FileMode(value)
}

//this regexp copied from useradd(8) manpage
var userOrGroupRx = regexp.MustCompile(`^[a-z_][a-z0-9_-]*\$?$`)

func parseUserOrGroupRef(value interface{}, ec *errorCollector, entryDesc string) *IntOrString {
	//default value
	if value == nil {
		return nil
	}

	switch val := value.(type) {
	case int64:
		if val < 0 {
			ec.addf("%s is invalid: user or group ID \"%d\" may not be negative", entryDesc, val)
		}
		if val >= 1<<32 {
			ec.addf("%s is invalid: user or group ID \"%d\" does not fit in uint32", entryDesc, val)
		}
		return &IntOrString{Int: uint32(val)}
	case string:
		if !userOrGroupRx.MatchString(val) {
			ec.addf("%s is invalid: \"%s\" is not an acceptable user or group name", entryDesc, val)
		}
		return &IntOrString{Str: val}
	default:
		ec.addf("%s is invalid: \"owner\"/\"group\" attributes must be strings or integers, found type %t", entryDesc, value)
		return nil
	}
}

func parseFileContent(content string, contentFrom string, ec *errorCollector, entryDesc string) string {
	//option 1: content given verbatim in "content" field
	if content != "" {
		return content
	}

	//option 2: content referenced in "contentFrom" field
	if contentFrom == "" {
		ec.addf("%s is invalid: missing content", entryDesc)
		return ""
	}
	bytes, err := ioutil.ReadFile(contentFrom)
	ec.add(err)
	return string(bytes)
}
