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

package files

import (
	"path/filepath"
	"sort"

	"../common"
)

//TargetFile represents a configuration file that can be provisioned by Holo.
type TargetFile struct {
	relTargetPath string //the target path relative to the common.TargetDirectory()
	orphaned      bool   //default: false
	repoEntries   RepoFiles
}

//NewTargetFileFromPathIn creates a TargetFile instance for which a path
//relative to a known location is known.
//
//    target := NewTargetFileFromPathIn(common.TargetDirectory(), targetPath)
//    target := NewTargetFileFromPathIn(common.ProvisionedDirectory(), provisionedPath)
func NewTargetFileFromPathIn(directory, path string) *TargetFile {
	//make path relative
	relTargetPath, _ := filepath.Rel(directory, path)
	return &TargetFile{relTargetPath: relTargetPath}
}

//PathIn returns the path to this target file relative to the given directory.
//
//    targetPath := target.pathIn(common.TargetDirectory())           // e.g. "/etc/foo.conf"
//    targetBasePath := target.pathIn(common.TargetBaseDirectory())   // e.g. "/var/lib/holo/base/etc/foo.conf"
//    provisionedPath := target.pathIn(common.ProvisionedDirectory()) // e.g. "/var/lib/holo/provisioned/etc/foo.conf"
//
func (target *TargetFile) PathIn(directory string) string {
	return filepath.Join(directory, target.relTargetPath)
}

//AddRepoEntry registers a new repository entry in this TargetFile instance.
func (target *TargetFile) AddRepoEntry(entry RepoFile) {
	target.repoEntries = append(target.repoEntries, entry)
}

//RepoEntries returns an ordered list of all repository entries for this
//TargetFile.
func (target *TargetFile) RepoEntries() RepoFiles {
	sort.Sort(target.repoEntries)
	return target.repoEntries
}

//EntityID implements the common.Entity interface.
func (target *TargetFile) EntityID() string {
	return target.PathIn(common.TargetDirectory())
}

//Report implements the common.Entity interface.
func (target *TargetFile) Report() *common.Report {
	r := common.Report{Target: target.EntityID()}

	if target.orphaned {
		_, strategy, assessment := target.scanOrphanedTargetBase()
		r.State = assessment
		r.AddLine(strategy, target.PathIn(common.TargetBaseDirectory()))
	} else {
		r.AddLine("store at", target.PathIn(common.TargetBaseDirectory()))
		for _, entry := range target.repoEntries {
			r.AddLine(entry.ApplicationStrategy(), entry.Path())
		}
	}

	return &r
}

//Apply implements the common.Entity interface.
func (target *TargetFile) Apply(withForce bool) {
	if target.orphaned {
		target.handleOrphanedTargetBase()
	} else {
		Apply(target, withForce)
	}
}
