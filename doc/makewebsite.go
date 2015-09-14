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

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type InputData struct {
	Title    string
	Contents template.HTML
}

func main() {
	//first argument is the stem of the Makefile target (e.g. "doc")
	//from which all other paths are computed
	stem := os.Args[1]
	inputFile := "doc/website-" + stem + ".pod"
	outputFile := "website/" + stem + ".html"

	//run pod2html on the input file and read the result for post-processing
	cmd := exec.Command("pod2html", "--noheader", "--noindex", "--infile="+inputFile)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot execute pod2html: %s\n", err.Error())
		return
	}
	contents := string(stdout.Bytes())

	//use only <body>...</body> contents
	contents = (regexp.MustCompile("<body>((?s:.+))</body>").FindStringSubmatch(contents))[1]
	contents = strings.TrimSpace(contents)

	//clean whitespace in <pre> elements
	contents = regexp.MustCompile("<pre>((?s:.+?))</pre>").ReplaceAllStringFunc(contents, cleanWhitespaceInPre)

	//divide into <section>s along <h1>
	opener := "<section><div class=\"fixed-width clearfix\">"
	closer := "</div></section>"
	contents = opener + strings.Replace(contents, "<h1 ", closer+opener+"<h1 ", -1) + closer
	//if output starts with a <h1>, we have a superfluous opener+closer at the start
	contents = strings.TrimPrefix(contents, opener+closer)

	//prepare input data for template processing
	inputData := InputData{
		Title:    "Holo - Minimalistic Configuration Management", //TODO: title per page
		Contents: template.HTML(contents),
	}

	//read output template
	templateString, err := ioutil.ReadFile("doc/template.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read doc/template.html: %s\n", err.Error())
		return
	}
	template, err := template.New("html").Parse(string(templateString))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse doc/template.html: %s\n", err.Error())
		return
	}

	//write result to output file
	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open %s: %s\n", outputFile, err.Error())
		return
	}
	err = template.Execute(file, inputData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot write %s: %s\n", outputFile, err.Error())
		return
	}
}

func cleanWhitespaceInPre(text string) string {
	// trim leading whitespace in code snippets, e.g.
	// input = "<pre><code>    foo\n      bar\n</code></pre>"
	// output = "<pre><code>foo\n  bar\n</code></pre>"
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "<pre>")
	text = strings.TrimSuffix(text, "</pre>")
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "<code>")
	text = strings.TrimSuffix(text, "</code>")
	lines := strings.Split(text, "\n")

	//find shortest leading spaces run
	minLeadingSpaces := 1000
	for _, line := range lines {
		for idx, runeVal := range line {
			if runeVal != ' ' {
				if minLeadingSpaces > idx {
					minLeadingSpaces = idx
				}
				break
			}
		}
	}

	var trimmedLines []string
	for _, line := range lines {
		trimmedLine := ""
		if len(line) >= minLeadingSpaces {
			trimmedLine = line[minLeadingSpaces:]
		}
		trimmedLines = append(trimmedLines, trimmedLine)
	}
	text = strings.Join(trimmedLines, "\n")

	//while we're at it, we also fix that pod2html wrongly double escapes "&gt;" etc. in <pre>
	text = strings.Replace(text, "&amp;gt;", "&gt;", -1)
	text = strings.Replace(text, "&amp;lt;", "&lt;", -1)

	return "<pre><code>" + text + "</code></pre>"
}
