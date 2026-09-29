// Package fonts downloads Google Fonts families into a garden so a theme can
// use any family by name while the site stays self-hosted.
//
// A family is downloaded once, stored under static/fonts/<slug>/ with its
// license, and recorded in static/fonts/fonts.lock.json with a SHA-256 for
// every file. Later builds read the lock and never touch the network, so a
// committed garden builds offline and reproducibly.
package fonts

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/shivamx96/leafpress/core/config"
)

const (
	// Dir is the garden directory that holds downloaded families.
	Dir = "static/fonts"
	// LockFile records every downloaded family and its files.
	LockFile = Dir + "/fonts.lock.json"

	lockVersion = 1
)

// Lock is the contents of LockFile.
type Lock struct {
	Version  int                `json:"version"`
	Families map[string]*Family `json:"families"`
}

// Family is one downloaded font family.
type Family struct {
	Source      string `json:"source"`
	License     string `json:"license"`
	LicenseFile string `json:"licenseFile"`
	Faces       []Face `json:"faces"`
}

// Face is one downloaded font file: a style and weight range for one
// character subset.
type Face struct {
	File         string `json:"file"`
	Weight       string `json:"weight"`
	Style        string `json:"style"`
	UnicodeRange string `json:"unicodeRange"`
	SHA256       string `json:"sha256"`
}

// ReadLock loads the lock file of the garden at root. A missing file is an
// empty lock.
func ReadLock(root string) (*Lock, error) {
	garden, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer garden.Close()
	return readLock(garden)
}

// readLock reads the lock through garden, which refuses paths that leave the
// project, such as a static/fonts symlink to another directory.
func readLock(garden *os.Root) (*Lock, error) {
	data, err := garden.ReadFile(filepath.FromSlash(LockFile))
	if errors.Is(err, os.ErrNotExist) {
		return &Lock{Version: lockVersion, Families: map[string]*Family{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", LockFile, err)
	}
	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse %s: %w", LockFile, err)
	}
	if lock.Version != lockVersion {
		return nil, fmt.Errorf("%s has version %d; this leafpress reads version %d", LockFile, lock.Version, lockVersion)
	}
	if lock.Families == nil {
		lock.Families = map[string]*Family{}
	}
	for name, family := range lock.Families {
		if family == nil {
			return nil, fmt.Errorf("%s: family %q has no entry; delete it from the lock and build again to download it", LockFile, name)
		}
		if err := family.validatePaths(); err != nil {
			return nil, fmt.Errorf("%s: family %q: %w", LockFile, name, err)
		}
	}
	return &lock, nil
}

// write saves the lock atomically so an interrupted build never leaves a
// truncated file behind.
func (l *Lock) write(garden *os.Root) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := filepath.FromSlash(Dir + "/.fonts.lock-" + randomSuffix() + ".json")
	if err := garden.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", LockFile, err)
	}
	if err := garden.Rename(tmp, filepath.FromSlash(LockFile)); err != nil {
		garden.Remove(tmp)
		return fmt.Errorf("write %s: %w", LockFile, err)
	}
	return nil
}

// randomSuffix names temporary files and directories.
func randomSuffix() string {
	var b [8]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// FontFaces converts the family into theme font declarations, so the build
// self-hosts it exactly like a family declared under theme.fonts.
func (f *Family) FontFaces(name string) []config.FontFace {
	faces := make([]config.FontFace, 0, len(f.Faces))
	for _, face := range f.Faces {
		faces = append(faces, config.FontFace{
			Family:       name,
			File:         face.File,
			Weight:       face.Weight,
			Style:        face.Style,
			Display:      "swap",
			UnicodeRange: face.UnicodeRange,
		})
	}
	return faces
}

// files lists every garden-relative path the family owns.
func (f *Family) files() []string {
	files := make([]string, 0, len(f.Faces)+1)
	for _, face := range f.Faces {
		files = append(files, face.File)
	}
	return append(files, f.LicenseFile)
}

// validatePaths keeps a hand-edited lock from pointing leafpress at files
// outside the fonts directory.
func (f *Family) validatePaths() error {
	for _, file := range f.files() {
		clean := path.Clean(file)
		if clean != file || !isUnder(Dir, clean) {
			return fmt.Errorf("file %q is not under %s/", file, Dir)
		}
	}
	return nil
}

// check reports whether every file of the family is present and unchanged.
// Missing files mean the family should be downloaded again. A file whose
// contents changed is an error: leafpress will not silently replace a file
// someone edited. Every surviving file is checked, so a missing file cannot
// hide an edited one that a new download would overwrite.
func (f *Family) check(garden *os.Root) (complete bool, err error) {
	complete = true
	for _, face := range f.Faces {
		data, err := garden.ReadFile(filepath.FromSlash(face.File))
		if errors.Is(err, os.ErrNotExist) {
			complete = false
			continue
		}
		if err != nil {
			return false, err
		}
		if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != face.SHA256 {
			return false, fmt.Errorf("%s does not match %s; delete %s/ and build again to download it afresh", face.File, LockFile, path.Dir(face.File))
		}
	}
	if _, err := garden.Stat(filepath.FromSlash(f.LicenseFile)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		complete = false
	}
	return complete, nil
}

func isUnder(dir, file string) bool {
	return len(file) > len(dir)+1 && file[:len(dir)+1] == dir+"/"
}

// hasPrefix reports whether data starts with prefix, for file signatures.
func hasPrefix(data []byte, prefix string) bool {
	return bytes.HasPrefix(data, []byte(prefix))
}
