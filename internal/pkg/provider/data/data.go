// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package data defines the structures and interfaces for custom provider data.
// Each provider can have its own machine configuration schema.
// When a provider starts, it reports its data schema back to Omni.
// Omni then uses this schema to render the appropriate UI and validate MachineRequests
package data

import (
	_ "embed"
)

//go:embed schema.json
var Schema []byte

// Data and schema.json should be in sync.
type Data struct {
	Cores            int    `yaml:"cores"`
	Memory           uint64 `yaml:"memory"`
	Architecture     string `yaml:"architecture"`
	DiskSize         int    `yaml:"disk_size"`
	StorageClassName string `yaml:"storage_class_name"`
	NetworkBinding   string `yaml:"network_binding,omitempty"`
}
