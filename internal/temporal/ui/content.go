// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

//go:build !dashboard_production

package ui

import (
	"embed"
	"io/fs"
)

// standin keeps Go-only builds independent of the production UI assets.
//
//go:embed standin/*
var standin embed.FS

func content() fs.FS {
	return mustFiles(fs.Sub(standin, "standin"))
}
