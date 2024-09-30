// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package provider

// Data is the provider custom machine config.
type Data struct {
	Architecture string `yaml:"architecture"`
	Cores        int    `yaml:"cores"`
	DiskSize     int    `yaml:"disk_size"`
	Memory       uint64 `yaml:"memory"`
}
