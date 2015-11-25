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

package plugins

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"../common"
)

//Plugin describes a plugin executable adhering to the holo-plugin-interface(7).
type Plugin struct {
	id             string
	executablePath string
}

//NewPlugin creates a new Plugin.
func NewPlugin(id string) *Plugin {
	executablePath := filepath.Join(common.TargetDirectory(), "usr/lib/holo/holo-"+id)
	return &Plugin{id, executablePath}
}

//NewPluginWithExecutablePath creates a new Plugin whose executable resides in
//a non-standard location. (This is used exclusively for testing plugins before
//they are installed.)
func NewPluginWithExecutablePath(id string, executablePath string) *Plugin {
	return &Plugin{id, executablePath}
}

//ID returns the plugin ID.
func (p *Plugin) ID() string {
	return p.id
}

//ResourceDirectory returns the path to the directory where this plugin may
//find its resources (entity definitions etc.).
func (p *Plugin) ResourceDirectory() string {
	//hard-coded resource directories for builtin plugins
	switch p.id {
	case "files":
		return common.RepoDirectory()
	case "users-groups":
		return common.EntityDirectory()
	case "run-scripts":
		return common.ScriptDirectory()
	}
	return filepath.Join(common.EntityDirectory(), p.id)
}

//CacheDirectory returns the path to the directory where this plugin may
//store temporary data.
func (p *Plugin) CacheDirectory() string {
	return filepath.Join(CachePath(), p.id)
}

//Run runs the plugin with the given arguments, producing output on the given
//output and error channels. Non-zero exit code is reported as a non-nil error.
func (p *Plugin) Run(arguments []string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command(p.executablePath, arguments...)
	cmd.Stdin = nil
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	//setup environment, mapping old-style variable names to those mandated by
	//holo-plugin-interface(7)
	env := os.Environ()
	env = append(env, "HOLO_API_VERSION=1")
	env = append(env, "HOLO_CACHE_DIR="+p.CacheDirectory())
	if common.TargetDirectory() != "/" {
		env = append(env, "HOLO_ROOT_DIR="+common.TargetDirectory())
	}
	cmd.Env = env

	return cmd.Run()
}
