// Package loader reads ArgFuscator-compatible JSON profile files and returns
// parsed ProfileFile values ready for use by the engine.
//
// Profiles are embedded at compile time from data/models/*.json using go:embed,
// so the binary is fully self-contained. Additional profiles can be loaded from
// an arbitrary fs.FS (e.g. os.DirFS) at runtime.
package loader

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"cmdFuscator/models"
)

// LoadFS reads every *.json file from the provided fs.FS and returns a slice of
// parsed ProfileFiles. The Name field of each ProfileFile is set to the base
// filename without the .json extension (e.g. "certutil").
//
// Files that fail to parse are skipped and their errors are collected; a non-nil
// error is returned only when no files could be loaded at all.
func LoadFS(fsys fs.FS) ([]*models.ProfileFile, error) {
	entries, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return nil, fmt.Errorf("loader: glob: %w", err)
	}

	var (
		profiles []*models.ProfileFile
		errs     []string
	)

	for _, entry := range entries {
		pf, err := loadFile(fsys, entry)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", entry, err))
			continue
		}
		profiles = append(profiles, pf)
	}

	if len(profiles) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("loader: all files failed:\n%s", strings.Join(errs, "\n"))
	}

	return profiles, nil
}

// loadFile reads and parses a single JSON profile file from fsys.
func loadFile(fsys fs.FS, name string) (*models.ProfileFile, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var pf models.ProfileFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	// Derive the executable name from the filename (strip directory + extension).
	base := filepath.Base(name)
	pf.Name = strings.TrimSuffix(base, filepath.Ext(base))

	return &pf, nil
}

// IndexByName returns a map from executable name (lowercase) to its ProfileFile.
// When multiple profiles share the same name the last one wins.
func IndexByName(profiles []*models.ProfileFile) map[string]*models.ProfileFile {
	idx := make(map[string]*models.ProfileFile, len(profiles))
	for _, pf := range profiles {
		idx[strings.ToLower(pf.Name)] = pf
	}
	return idx
}

// GroupByPlatform partitions a slice of ProfileFiles into per-platform buckets.
// The key is the lowercased platform string from the first profile in each file
// (e.g. "windows", "linux", "macos").
//
// A ProfileFile may contain profiles for more than one platform; in that case it
// appears in all matching buckets.
func GroupByPlatform(profiles []*models.ProfileFile) map[string][]*models.ProfileFile {
	groups := make(map[string][]*models.ProfileFile)
	for _, pf := range profiles {
		seen := make(map[string]bool)
		for _, p := range pf.Profiles {
			plat := strings.ToLower(p.Platform)
			if plat == "" {
				plat = "other"
			}
			if !seen[plat] {
				groups[plat] = append(groups[plat], pf)
				seen[plat] = true
			}
		}
	}
	return groups
}
