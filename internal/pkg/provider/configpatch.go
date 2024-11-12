// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package provider

// ConfigPatch defines a config patch to be created for each machine managed by the provider.
type ConfigPatch struct {
	// Prefix is the config patch prefix.
	Prefix string `yaml:"prefix"`
	// Data is the raw patch data.
	Data string `yaml:"data"`
}
