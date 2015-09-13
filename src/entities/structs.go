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

type Group struct {
	name   string
	gid    int
	system bool
}

func (g Group) Name() string   { return g.name }
func (g Group) NumericId() int { return g.gid }
func (g Group) System() bool   { return g.system }

//Definition represents an entity definition file (found below /usr/share/holo)
//and the entity definitions contained within it.
type Definition struct {
	File   string
	Groups []Group
}
