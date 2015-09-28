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

package files

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"../common"
)

type fileType int

const (
	fileUnknown fileType = 0
	fileMissing fileType = 1
	fileRegular fileType = 2
	fileSymlink fileType = 3
)

//RenderDiff creates a unified diff of a target file and its last provisioned
//version, similar to `diff /var/lib/holo/provisioned/$FILE $FILE`, but it also
//handles symlinks and missing files gracefully. The output is always a patch
//that can be applied to last provisioned version into the current version.
func (configFile *ConfigFile) RenderDiff() ([]byte, error) {
	//stat both files (non-existence is not an error here, we handle that later)
	//and check that they are manageable
	fromPath := configFile.ProvisionedPath()
	fromType, fromInfo, err := lstatForDiff(fromPath)
	if err != nil {
		return nil, err
	}

	toPath := configFile.TargetPath()
	toType, toInfo, err := lstatForDiff(toPath)
	if err != nil {
		return nil, err
	}

	//part 1: both files are missing -> empty diff
	if fromType == fileMissing && toType == fileMissing {
		return []byte(nil), nil
	}

	//part 2: different file types -> act similarly to `git diff` and print a
	//deletion diff, followed by a creation diff
	if fromType != toType {
		switch fromType {
		//note: "case fileMissing" was already handled above
		case fileRegular:
			//use `diff` with toPath = /dev/null, fabricate a suitable header
			return makeRegularDeleteDiff(fromPath, toPath, toInfo.Mode()), nil
		case fileSymlink:
			//fabricate the complete diff output
			linkTarget, err := os.Readlink(fromPath)
			if err != nil {
				return nil, err
			}
			return makeSymlinkDeleteDiff(linkTarget, toPath), nil
		}
		switch toType {
		case fileRegular:
			//use `diff` with fromPath = /dev/null, fabricate a suitable header
			return makeRegularCreateDiff(toPath, toPath, toInfo.Mode()), nil
		case fileSymlink:
			//fabricate the complete diff output
			linkTarget, err := os.Readlink(toPath)
			if err != nil {
				return nil, err
			}
			return makeSymlinkCreateDiff(linkTarget, toPath), nil
		}
	}

	//part 3: both files are symlinks - fabricate a modification diff
	if toType == fileSymlink {
		fromLinkTarget, err := os.Readlink(fromPath)
		if err != nil {
			return nil, err
		}
		toLinkTarget, err := os.Readlink(toPath)
		if err != nil {
			return nil, err
		}
		return makeSymlinkModifyDiff(fromLinkTarget, toLinkTarget, toPath), nil
	}

	//part 4: both files are regular - use `diff` and fabricate a suitable header
	return makeRegularModifyDiff(fromPath, toPath, toPath, fromInfo.Mode(), toInfo.Mode()), nil
}

func lstatForDiff(path string) (fileType fileType, fi os.FileInfo, e error) {
	info, err := os.Lstat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fileUnknown, nil, err
		}
		//missing file is not an error
		return fileMissing, nil, nil
	}
	if info.Mode().IsRegular() {
		return fileRegular, info, nil
	}
	if common.IsFileInfoASymbolicLink(info) {
		return fileSymlink, info, nil
	}
	return fileUnknown, nil, fmt.Errorf("%s is not a manageable file", path)
}

func getDiffBody(fromPath, toPath string) []byte {
	//skip the error handling here; a non-empty diff produces a non-zero exit
	//code and we don't want to fail in that case
	output, _ := common.ExecProgram([]byte{}, "diff", "-u", fromPath, toPath)
	//remove the header, up to the first hunk (started by a line like "@@ -1 +0,0")
	return regexp.MustCompile("^(?s:.+?)(?m:^@@)").ReplaceAll(output, []byte("@@"))
}

func makeRegularCreateDiff(path, reportedPath string, mode os.FileMode) []byte {
	header := []byte(strings.Join([]string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
		fmt.Sprintf("new file mode 100%o\n", int(mode)),
		"--- /dev/null\n",
		fmt.Sprintf("+++ b%s\n", reportedPath),
	}, ""))
	return append(header, getDiffBody("/dev/null", path)...)
}

func makeRegularModifyDiff(fromPath, toPath, reportedPath string, fromMode os.FileMode, toMode os.FileMode) []byte {
	headers := []string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
	}
	if fromMode != toMode {
		headers = append(headers,
			fmt.Sprintf("old mode 100%o\n", int(fromMode)),
			fmt.Sprintf("new mode 100%o\n", int(toMode)),
		)
	}
	headers = append(headers,
		fmt.Sprintf("--- a%s\n", reportedPath),
		fmt.Sprintf("+++ b%s\n", reportedPath),
	)
	header := []byte(strings.Join(headers, ""))
	return append(header, getDiffBody(fromPath, toPath)...)
}

func makeRegularDeleteDiff(path, reportedPath string, mode os.FileMode) []byte {
	header := []byte(strings.Join([]string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
		fmt.Sprintf("deleted file mode 100%o\n", int(mode)),
		fmt.Sprintf("--- a%s\n", reportedPath),
		"+++ /dev/null\n",
	}, ""))
	return append(header, getDiffBody(path, "/dev/null")...)
}

func makeSymlinkCreateDiff(linkTarget, reportedPath string) []byte {
	//NOTE: This function makes the reasonable assumption that
	//      !strings.Contains(linkTarget, "\n").
	return []byte(strings.Join([]string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
		"new file mode 120000\n",
		"--- /dev/null\n",
		fmt.Sprintf("+++ b%s\n", reportedPath),
		"@@ -0,0 +1 @@\n",
		fmt.Sprintf("+%s\n", linkTarget),
		"\\ No newline at end of file\n",
	}, ""))
}

func makeSymlinkModifyDiff(fromLinkTarget, toLinkTarget, reportedPath string) []byte {
	//NOTE: This function makes the reasonable assumption that
	//      !strings.Contains(linkTarget, "\n").
	return []byte(strings.Join([]string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
		fmt.Sprintf("--- a%s\n", reportedPath),
		fmt.Sprintf("+++ b%s\n", reportedPath),
		"@@ -1 +1 @@\n",
		fmt.Sprintf("-%s\n", fromLinkTarget),
		"\\ No newline at end of file\n",
		fmt.Sprintf("+%s\n", toLinkTarget),
		"\\ No newline at end of file\n",
	}, ""))
}

func makeSymlinkDeleteDiff(linkTarget, reportedPath string) []byte {
	//NOTE: This function makes the reasonable assumption that
	//      !strings.Contains(linkTarget, "\n").
	return []byte(strings.Join([]string{
		fmt.Sprintf("diff --git a%s b%s\n", reportedPath, reportedPath),
		"deleted file mode 120000\n",
		fmt.Sprintf("--- a%s\n", reportedPath),
		"+++ /dev/null\n",
		"@@ -1 +0,0 @@\n",
		fmt.Sprintf("-%s\n", linkTarget),
		"\\ No newline at end of file\n",
	}, ""))
}
