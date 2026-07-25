package terminal

import (
	"strings"
	"testing"
)

// TestAutoInstallToolMappings locks in the correct installer for tools whose
// mapping was previously wrong or missing:
//   - paramspider is a Python tool (pipx), not a Go module, so
//     "go install github.com/devanshbatham/paramspider@latest" always failed.
//   - arjun and uro are Python tools (pipx) that were absent from packageMap,
//     so resolvePackage returned "" and they were never auto-installed even
//     though the agent prompt recommends them.
//   - x8 and findomain are Rust tools (cargo): x8 was absent from packageMap;
//     findomain was documented as a cargo tool but only lived in the generic
//     apt path (not in apt repos -> install always failed).
func TestAutoInstallToolMappings(t *testing.T) {
	// Each tool must resolve via packageMap; otherwise installPackage is never
	// reached and the tool can never be auto-installed.
	for _, tool := range []string{"paramspider", "arjun", "x8", "uro", "findomain"} {
		if resolvePackage(tool) == "" {
			t.Errorf("resolvePackage(%q) = %q; tool missing from packageMap, so it can never be auto-installed", tool, "")
		}
	}

	// paramspider must NOT be a Go tool, and paramspider/arjun/uro must be pipx.
	if _, ok := goTools["paramspider"]; ok {
		t.Error(`paramspider must not be in goTools: it is a Python tool and "go install" of that path always fails`)
	}
	for _, tool := range []string{"paramspider", "arjun", "uro"} {
		if _, ok := pipxTools[tool]; !ok {
			t.Errorf("%s must be in pipxTools (pip/pipx-installed Python tool)", tool)
		}
	}

	// x8 and findomain must be cargo tools.
	for _, tool := range []string{"x8", "findomain"} {
		if _, ok := cargoTools[tool]; !ok {
			t.Errorf("%s must be in cargoTools (Rust/cargo tool)", tool)
		}
	}

	// A tool must never be classified under two installers at once.
	all := []map[string]string{pipxTools, cargoTools, goTools, npmTools}
	for _, tool := range []string{"paramspider", "arjun", "x8", "uro", "findomain"} {
		hits := 0
		for _, m := range all {
			if _, ok := m[tool]; ok {
				hits++
			}
		}
		if hits != 1 {
			t.Errorf("%s is classified by %d installer maps; want exactly 1", tool, hits)
		}
	}
}

// TestAptGetInstallRefreshesLists guards the regression where package installs
// ran `apt-get install` without a preceding `apt-get update`, which fails with
// "Unable to locate package" on a stale/empty apt cache.
func TestAptGetInstallRefreshesLists(t *testing.T) {
	cmd := aptGetInstall("whois")

	updateAt := strings.Index(cmd, "apt-get update")
	installAt := strings.Index(cmd, "apt-get install -y -q whois")

	if updateAt < 0 {
		t.Fatalf("aptGetInstall must run `apt-get update` before install; got %q", cmd)
	}
	if installAt < 0 {
		t.Fatalf("aptGetInstall must install the requested package; got %q", cmd)
	}
	if updateAt > installAt {
		t.Errorf("`apt-get update` must precede install; got %q", cmd)
	}
}
