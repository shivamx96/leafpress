package fonts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/shivamx96/leafpress/core/config"
)

// Options controls how Resolve treats families that are not downloaded yet.
type Options struct {
	// Download allows fetching missing families from Google Fonts.
	Download bool
	// NoDownloadReason completes the warning for a missing family when
	// Download is false, for example "downloads are off (--offline), so
	// system fonts are used".
	NoDownloadReason string
	// Client is the Google Fonts client; nil uses the public service.
	Client *Client
	// Log receives progress messages.
	Log io.Writer
}

// Result is what Resolve found for the requested families.
type Result struct {
	// Faces declares every available family for the build.
	Faces []config.FontFace
	// Warnings explains, per family, why it is still unavailable. Those
	// families fall back to the reader's system fonts.
	Warnings map[string]string
}

// perFamilyTimeout bounds one family's download so an unresponsive network
// cannot stall a build indefinitely.
const perFamilyTimeout = 2 * time.Minute

// Resolve makes each family available from the garden's lock, downloading
// missing ones when allowed. Download and lookup failures become warnings so
// the build can continue with system fonts. Only a problem with files already
// in the garden, such as an edited font file, is returned as an error.
func Resolve(root string, families []string, opts Options) (Result, error) {
	result := Result{Warnings: map[string]string{}}
	if len(families) == 0 {
		return result, nil
	}
	lock, err := ReadLock(root)
	if err != nil {
		return result, err
	}
	client := opts.Client
	if client == nil {
		client = Google()
	}
	log := opts.Log
	if log == nil {
		log = io.Discard
	}

	downloaded := false
	for _, family := range families {
		entry, managed := lock.Families[family]
		if managed {
			complete, err := entry.check(root)
			if err != nil {
				return result, fmt.Errorf("font %q: %w", family, err)
			}
			if complete {
				result.Faces = append(result.Faces, entry.FontFaces(family)...)
				continue
			}
		}
		if !opts.Download {
			result.Warnings[family] = fmt.Sprintf("font family %q is not downloaded yet: %s", family, opts.NoDownloadReason)
			continue
		}

		fmt.Fprintf(log, "Downloading %q from Google Fonts...\n", family)
		fresh, size, err := fetch(root, client, family, entry)
		if err != nil {
			var notFound *NotFoundError
			if errors.As(err, &notFound) {
				result.Warnings[family] = "font family " + notFound.Error() + "; using system fonts"
			} else {
				result.Warnings[family] = fmt.Sprintf("could not download font family %q from Google Fonts: %v; using system fonts until a later build succeeds", family, err)
			}
			continue
		}
		lock.Families[family] = fresh
		if err := lock.write(root); err != nil {
			return result, err
		}
		downloaded = true
		fmt.Fprintf(log, "  saved %s/%s/ (%d files, %s, %s)\n", Dir, Slug(family), len(fresh.Faces)+1, formatSize(size), fresh.License)
		result.Faces = append(result.Faces, fresh.FontFaces(family)...)
	}
	if downloaded {
		fmt.Fprintf(log, "  Commit %s/ so later builds work offline.\n", Dir)
	}
	return result, nil
}

// fetch downloads family into a hidden staging directory and moves it into
// place only when every file arrived. previous is the family's old lock
// entry, whose files may be replaced; any other existing directory is left
// alone.
func fetch(root string, client *Client, family string, previous *Family) (*Family, int64, error) {
	fontsDir := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(fontsDir, 0755); err != nil {
		return nil, 0, err
	}
	slug := Slug(family)
	if slug == "" {
		return nil, 0, fmt.Errorf("%q has no letters or digits to name its folder", family)
	}
	target := filepath.Join(fontsDir, slug)
	if previous != nil {
		if err := removeFiles(root, previous); err != nil {
			return nil, 0, err
		}
	}
	if _, err := os.Stat(target); err == nil {
		return nil, 0, fmt.Errorf("%s/%s/ already exists and was not downloaded by leafpress; declare its files under theme.fonts, or move it and build again", Dir, slug)
	}

	// Hidden names are skipped when static/ is published, so a staging
	// directory left by an interrupted build never reaches the site.
	stage, err := os.MkdirTemp(fontsDir, "."+slug+"-download-")
	if err != nil {
		return nil, 0, err
	}
	defer os.RemoveAll(stage)

	ctx, cancel := context.WithTimeout(context.Background(), perFamilyTimeout)
	defer cancel()
	entry, size, err := client.download(ctx, family, stage)
	if err != nil {
		return nil, 0, err
	}
	if err := os.Rename(stage, target); err != nil {
		return nil, 0, err
	}
	return entry, size, nil
}

// removeFiles deletes the files a previous download of the family owned,
// then its directory if nothing else is in it.
func removeFiles(root string, family *Family) error {
	dirs := map[string]bool{}
	for _, file := range family.files() {
		full := filepath.Join(root, filepath.FromSlash(file))
		if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		dirs[filepath.Dir(full)] = true
	}
	for dir := range dirs {
		if err := os.Remove(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s holds files leafpress did not download; move them and build again", dir)
		}
	}
	return nil
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%d KB", (bytes+512)/1024)
}
