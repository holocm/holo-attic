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

package entities

import (
	"fmt"
	"strings"
)

//Entity provides a common interface for configuration entities that are not
//files, such as user accounts and user groups.
type Entity interface {
	//EntityID returns a string that uniquely identifies the entity, usually in
	//the form "type:name". This is how the entity can be addressed as a target
	//in the argument list foe "holo apply", e.g. "holo apply /etc/sudoers
	//group:foo" will apply /etc/sudoers and the group "foo". Therefore, entity
	//IDs should not contain whitespaces or characters that have a special
	//meaning on the shell.
	EntityID() string
	//DefinitionFile returns the path to the file containing the definition of this entity.
	DefinitionFile() string
	//Attributes returns a string describing additional attributes set for this entity,
	//alternatively an empty string.
	Attributes() string
}

//Group represents a UNIX group (as registered in /etc/group). It implements
//the Entity interface and is handled accordingly.
type Group struct {
	name           string
	gid            int
	system         bool
	definitionFile string
}

//Name returns the group name (the first field in /etc/group).
func (g Group) Name() string { return g.name }

//NumericID returns the GID (the third field in /etc/group).
func (g Group) NumericID() int { return g.gid }

//System returns true if the group shall be created as a system group.
func (g Group) System() bool { return g.system }

//EntityID implements the Entity interface for Group.
func (g Group) EntityID() string { return "group:" + g.name }

//DefinitionFile implements the Entity interface for Group.
func (g Group) DefinitionFile() string { return g.definitionFile }

//Attributes implements the Entity interface for Group.
func (g Group) Attributes() string {
	attrs := []string{}
	if g.system {
		attrs = append(attrs, "type: system")
	}
	if g.gid > 0 {
		attrs = append(attrs, fmt.Sprintf("gid: %d", g.gid))
	}
	return strings.Join(attrs, ", ")
}

//Entities holds a slice of Entity instances, and implements some methods to
//satisfy the sort.Interface interface.
type Entities []Entity

func (e Entities) Len() int           { return len(e) }
func (e Entities) Less(i, j int) bool { return e[i].DefinitionFile() < e[j].DefinitionFile() }
func (e Entities) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
