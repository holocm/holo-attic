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
	"io/ioutil"
	"path/filepath"
	"strings"

	"../../shared"
	"../common"
)

//Configuration contains the parsed contents of /etc/holorc.
type Configuration struct {
	Plugins []string
}

//ReadConfiguration reads the configuration file /etc/holorc.
func ReadConfiguration() *Configuration {
	path := filepath.Join(common.TargetDirectory(), "etc/holorc")

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		r := shared.Report{Action: "read", Target: path}
		r.AddError(err.Error())
		r.Print()
		return nil
	}

	var result Configuration
	lines := strings.SplitN(strings.TrimSpace(string(contents)), "\n", -1)
	for _, line := range lines {
		//ignore comments and empty lines
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		//collect plugin IDs
		if strings.HasPrefix(line, "plugin ") {
			pluginID := strings.TrimSpace(strings.TrimPrefix(line, "plugin"))
			result.Plugins = append(result.Plugins, pluginID)
			continue
		} else {
			//unknown line
			r := shared.Report{Action: "read", Target: path}
			r.AddError("unknown command: %s", line)
			r.Print()
			return nil
		}
	}

	return &result
}
