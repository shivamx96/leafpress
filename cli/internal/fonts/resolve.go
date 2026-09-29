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
	garden, err := os.OpenRoot(root)
	if err != nil {
		return result, err
	}
	defer garden.Close()
	lock, err := readLock(garden)
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
			complete, err := entry.check(garden)
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
		fresh, size, err := fetch(garden, client, family, entry)
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
		if err := lock.write(garden); err != nil {
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

// fetch downloads family into a hidden staging directory and swaps it into
// place only when every file arrived, so a failed download leaves any
// existing files untouched. previous is the family's lock entry when an
// earlier download is incomplete; its folder may be replaced only if it
// holds nothing but that download's files. Any other existing folder is
// left alone. Every path goes through garden, so a symlink cannot redirect
// writes outside the project.
func fetch(garden *os.Root, client *Client, family string, previous *Family) (*Family, int64, error) {
	slug := Slug(family)
	if slug == "" {
		return nil, 0, fmt.Errorf("%q has no letters or digits to name its folder", family)
	}
	if err := garden.MkdirAll(filepath.FromSlash(Dir), 0755); err != nil {
		return nil, 0, err
	}
	target := filepath.FromSlash(Dir + "/" + slug)
	if previous != nil {
		if err := onlyFamilyFiles(garden, target, previous); err != nil {
			return nil, 0, err
		}
	} else if _, err := garden.Lstat(target); err == nil {
		return nil, 0, fmt.Errorf("%s/%s/ already exists and was not downloaded by leafpress; declare its files under theme.fonts, or move it and build again", Dir, slug)
	}

	// Hidden names are skipped when static/ is published, so a staging
	// directory left by an interrupted build never reaches the site.
	stage := Dir + "/." + slug + "-download-" + randomSuffix()
	if err := garden.Mkdir(filepath.FromSlash(stage), 0755); err != nil {
		return nil, 0, err
	}
	defer garden.RemoveAll(filepath.FromSlash(stage))

	ctx, cancel := context.WithTimeout(context.Background(), perFamilyTimeout)
	defer cancel()
	entry, size, err := client.download(ctx, family, garden, stage)
	if err != nil {
		return nil, 0, err
	}
	if err := swapIn(garden, filepath.FromSlash(stage), target); err != nil {
		return nil, 0, err
	}
	return entry, size, nil
}

// onlyFamilyFiles checks that target holds nothing but files a previous
// download of the family owned, so replacing the folder cannot discard
// anyone else's files.
func onlyFamilyFiles(garden *os.Root, target string, previous *Family) error {
	dir, err := garden.Open(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return err
	}
	owned := map[string]bool{}
	for _, file := range previous.files() {
		owned[filepath.FromSlash(file)] = true
	}
	for _, entry := range entries {
		if !owned[filepath.Join(target, entry.Name())] {
			return fmt.Errorf("%s holds files leafpress did not download; move them and build again", filepath.ToSlash(target))
		}
	}
	return nil
}

// swapIn moves stage to target. An existing target is moved aside first and
// restored if the swap fails, then removed once the new folder is in place.
func swapIn(garden *os.Root, stage, target string) error {
	if _, err := garden.Lstat(target); errors.Is(err, os.ErrNotExist) {
		return garden.Rename(stage, target)
	} else if err != nil {
		return err
	}
	backup := filepath.Join(filepath.Dir(target), "."+filepath.Base(target)+"-previous-"+randomSuffix())
	if err := garden.Rename(target, backup); err != nil {
		return err
	}
	if err := garden.Rename(stage, target); err != nil {
		if restoreErr := garden.Rename(backup, target); restoreErr != nil {
			return errors.Join(err, fmt.Errorf("restore %s: %w", filepath.ToSlash(target), restoreErr))
		}
		return err
	}
	return garden.RemoveAll(backup)
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%d KB", (bytes+512)/1024)
}
