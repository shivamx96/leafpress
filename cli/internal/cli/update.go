package cli

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	githubRepo    = "shivamx96/leafpress"
	githubAPIURL  = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	maxUpdateSize = 256 << 20
)

var updateHTTPClient = &http.Client{Timeout: 2 * time.Minute}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func updateCmd(currentVersion string) *cobra.Command {
	var forceUpdate bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update leafpress to the latest version",
		Long:  `Checks for the latest version and updates the leafpress binary if a newer version is available.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(currentVersion, forceUpdate)
		},
	}

	cmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Force update even if already on latest version")

	return cmd
}

func runUpdate(currentVersion string, force bool) error {
	fmt.Println("Checking for updates...")

	// Fetch latest release info
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestTag := withVPrefix(release.TagName)
	currentTag := withVPrefix(currentVersion)
	install, err := shouldInstallVersion(currentVersion, release.TagName, force)
	if err != nil {
		return err
	}
	if !install {
		fmt.Printf("Already on the latest version (%s)\n", strings.TrimPrefix(currentTag, "v"))
		return nil
	}

	latestVersion := strings.TrimPrefix(latestTag, "v")
	if semver.IsValid(currentTag) && semver.Compare(latestTag, currentTag) == 0 {
		fmt.Printf("Reinstalling version %s...\n", latestVersion)
	} else {
		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
	}

	// Find the right asset for this OS/arch
	assetName := getAssetName(latestTag)
	downloadURL := releaseAssetURL(release, assetName)
	if downloadURL == "" {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	checksumsURL := releaseAssetURL(release, "checksums.txt")
	if checksumsURL == "" {
		return fmt.Errorf("release %s does not include checksums.txt", latestTag)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	archive, err := downloadUpdateAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	checksums, err := downloadUpdateAsset(checksumsURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	if err := verifyAssetChecksum(archive, checksums, assetName); err != nil {
		return fmt.Errorf("update integrity check failed: %w", err)
	}

	// Extract only the expected executable after the archive itself is verified.
	binary, err := extractBinary(archive, assetName)
	if err != nil {
		return fmt.Errorf("failed to extract update: %w", err)
	}

	if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolved
	}
	if err := replaceExecutable(execPath, binary); err != nil {
		return fmt.Errorf("failed to replace binary safely: %w", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latestVersion)
	return nil
}

func extractBinary(archive []byte, assetName string) ([]byte, error) {
	binaryName := "leafpress"
	if strings.HasSuffix(assetName, ".zip") {
		binaryName += ".exe"
		return extractBinaryFromZip(archive, binaryName)
	}
	return extractBinaryFromTarGz(bytes.NewReader(archive), binaryName)
}

func extractBinaryFromTarGz(r io.Reader, binaryName string) ([]byte, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg && path.Base(header.Name) == binaryName {
			return readUpdateBinary(tr, header.Size)
		}
	}

	return nil, fmt.Errorf("%s not found in archive", binaryName)
}

func extractBinaryFromZip(archive []byte, binaryName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, file := range zr.File {
		if !file.FileInfo().Mode().IsRegular() || path.Base(file.Name) != binaryName {
			continue
		}
		if file.UncompressedSize64 > maxUpdateSize {
			return nil, fmt.Errorf("%s exceeds update size limit", binaryName)
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := readUpdateBinary(r, int64(file.UncompressedSize64))
		closeErr := r.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return data, nil
	}
	return nil, fmt.Errorf("%s not found in archive", binaryName)
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := updateHTTPClient.Get(githubAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName(version string) string {
	return getAssetNameFor(version, runtime.GOOS, runtime.GOARCH)
}

func getAssetNameFor(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("leafpress-%s-%s-%s%s", version, goos, goarch, ext)
}

func withVPrefix(version string) string {
	version = strings.TrimSpace(version)
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func shouldInstallVersion(currentVersion, latestVersion string, force bool) (bool, error) {
	latestTag := withVPrefix(latestVersion)
	if !semver.IsValid(latestTag) {
		return false, fmt.Errorf("latest release has invalid semantic version %q", latestVersion)
	}
	currentTag := withVPrefix(currentVersion)
	if !semver.IsValid(currentTag) {
		return true, nil
	}
	switch semver.Compare(latestTag, currentTag) {
	case -1:
		return false, fmt.Errorf("refusing to downgrade from %s to %s", currentTag, latestTag)
	case 0:
		return force, nil
	default:
		return true, nil
	}
}

func releaseAssetURL(release *githubRelease, name string) string {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

func downloadUpdateAsset(downloadURL string) ([]byte, error) {
	resp, err := updateHTTPClient.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxUpdateSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxUpdateSize {
		return nil, fmt.Errorf("asset exceeds %d-byte size limit", maxUpdateSize)
	}
	return data, nil
}

func verifyAssetChecksum(archive, checksums []byte, assetName string) error {
	var expected string
	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		expected = strings.ToLower(fields[0])
		break
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(expected) != sha256.Size*2 {
		return fmt.Errorf("no valid SHA-256 entry for %s", assetName)
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return fmt.Errorf("invalid SHA-256 entry for %s: %w", assetName, err)
	}
	actual := sha256.Sum256(archive)
	if hex.EncodeToString(actual[:]) != expected {
		return fmt.Errorf("SHA-256 mismatch for %s", assetName)
	}
	return nil
}

func readUpdateBinary(r io.Reader, declaredSize int64) ([]byte, error) {
	if declaredSize < 0 || declaredSize > maxUpdateSize {
		return nil, fmt.Errorf("binary exceeds update size limit")
	}
	data, err := io.ReadAll(io.LimitReader(r, maxUpdateSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxUpdateSize || int64(len(data)) != declaredSize {
		return nil, fmt.Errorf("binary size does not match archive metadata")
	}
	return data, nil
}

func replaceExecutable(execPath string, binary []byte) error {
	info, err := os.Stat(execPath)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(execPath), ".leafpress-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := tmp.Write(binary); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		return os.Rename(tmpPath, execPath)
	}

	// Windows cannot replace an existing executable with Rename. Move the old
	// image aside, install the verified file, and roll back if the second move
	// fails. The .old file is removed on the next successful update attempt.
	backupPath := execPath + ".old"
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous backup: %w", err)
	}
	if err := os.Rename(execPath, backupPath); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		if rollbackErr := os.Rename(backupPath, execPath); rollbackErr != nil {
			return fmt.Errorf("install failed: %v (rollback failed: %v)", err, rollbackErr)
		}
		return err
	}

	return nil
}
