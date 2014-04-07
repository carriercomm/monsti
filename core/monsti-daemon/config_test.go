// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"os"
	"path/filepath"
	"testing"

	mtest "pkg.monsti.org/monsti/api/util/testing"
)

func TestLoadConfigs(t *testing.T) {
	files := map[string]string{
		"/configs/document.yaml": `
namespace: core
nodetypes:
  - id: document
    name: {en: "Document", de: "Dokument"}
    fields:
      - id: body
        name: {en: "Body", de: "Rumpf"}
        type: widgets
`,
		"/image.yaml": `
namespace: core
nodetypes:
  - id: image
    name: {en: "Image", de: "Bild"}
    fields:
      - id: file
        name: {en: "File", de: "Datei"}
        required: yes
        type: file
`}
	root, cleanup, err := mtest.CreateDirectoryTree(files, "TestLoadConfigs")
	if err != nil {
		t.Fatalf("Could not create test files: %v", err)
	}
	defer cleanup()
	err = os.Symlink(filepath.Join(root, "image.yaml"),
		filepath.Join(root, "configs", "image.yaml"))
	if err != nil {
		t.Fatalf("Could not create symlink to config: %v", err)
	}
	config, err := loadConfigs(filepath.Join(root, "configs"))
	if err != nil {
		t.Fatalf("Could not load configs: %v", err)
	}
	if len(config.NodeTypes) != 2 {
		t.Fatalf("Should have two node types, but found %v", len(config.NodeTypes))
	}
	nodeType, ok := config.NodeTypes["core.image"]
	if !ok {
		t.Fatalf("Missing node type core.image")
	}
	if nodeType.Id != "core.image" {
		t.Fatalf("Id field of core.image is %q, should be core.image", nodeType.Id)
	}
}
