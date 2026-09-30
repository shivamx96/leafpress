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

// fetch downloads family into a hidden staging directory and moves the files
// into place one by one only after every file arrived, so a failed download
// leaves existing files untouched. previous is the family's lock entry when
// an earlier download is incomplete; only files that download owned may be
// replaced. Anything else in the folder, including files added while the
// download runs, is left alone. Every path goes through garden, so a symlink
// cannot redirect writes outside the project.
func fetch(garden *os.Root, client *Client, family string, previous *Family) (*Family, int64, error) {
	slug := Slug(family)
	if slug == "" {
		return nil, 0, fmt.Errorf("%q has no letters or digits to name its folder", family)
	}
	if err := garden.MkdirAll(filepath.FromSlash(Dir), 0755); err != nil {
		return nil, 0, err
	}
	target := filepath.FromSlash(Dir + "/" + slug)
	if previous == nil {
		if _, err := garden.Lstat(target); err == nil {
			return nil, 0, fmt.Errorf("%s/%s/ already exists and was not downloaded by leafpress; declare its files under theme.fonts, or move it and build again", Dir, slug)
		}
	}

	// Hidden names are skipped when static/ is published, so a staging
	// directory left by an interrupted build never reaches the site.
	stage := filepath.FromSlash(Dir + "/." + slug + "-download-" + randomSuffix())
	if err := garden.Mkdir(stage, 0755); err != nil {
		return nil, 0, err
	}
	defer garden.RemoveAll(stage)

	ctx, cancel := context.WithTimeout(context.Background(), perFamilyTimeout)
	defer cancel()
	entry, size, err := client.download(ctx, family, garden, stage)
	if err != nil {
		return nil, 0, err
	}
	if err := placeFiles(garden, stage, target, entry, previous); err != nil {
		return nil, 0, err
	}
	return entry, size, nil
}

// placeFiles moves the downloaded files from stage into target. It first
// parks each file in target under a hidden name, refusing to continue if a
// final name is taken by a file leafpress did not download, then renames the
// parked files into place. Files of the previous download that the new one
// no longer needs are removed; nothing else in target is touched. Until the
// final renames begin, a failure leaves the previous files intact.
func placeFiles(garden *os.Root, stage, target string, fresh, previous *Family) error {
	owned := map[string]bool{}
	if previous != nil {
		for _, file := range previous.files() {
			owned[filepath.FromSlash(file)] = true
		}
	}
	if err := garden.MkdirAll(target, 0755); err != nil {
		return err
	}

	parked := map[string]string{}
	cleanup := func() {
		for _, tmp := range parked {
			garden.Remove(tmp)
		}
	}
	for _, file := range fresh.files() {
		final := filepath.FromSlash(file)
		if _, err := garden.Lstat(final); err == nil && !owned[final] {
			cleanup()
			return fmt.Errorf("%s was added while the family was downloading and was not downloaded by leafpress; move it and build again", file)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanup()
			return err
		}
		tmp := filepath.Join(target, "."+filepath.Base(final)+".new")
		if err := garden.Rename(filepath.Join(stage, filepath.Base(final)), tmp); err != nil {
			cleanup()
			return err
		}
		parked[final] = tmp
	}
	for final, tmp := range parked {
		if err := garden.Rename(tmp, final); err != nil {
			cleanup()
			return err
		}
	}
	for file := range owned {
		if _, replaced := parked[file]; replaced {
			continue
		}
		if err := garden.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
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
