package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	githubRepo   = "shivamx96/leafpress"
	githubAPIURL = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

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

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	if latestVersion == currentVersion && !force {
		fmt.Printf("Already on the latest version (%s)\n", currentVersion)
		return nil
	}

	if latestVersion == currentVersion && force {
		fmt.Printf("Reinstalling version %s...\n", latestVersion)
	} else {
		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
	}

	// Find the right asset for this OS/arch
	assetName := getAssetName(release.TagName)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download tarball
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: HTTP %d", resp.StatusCode)
	}

	// Extract binary from tarball
	tmpPath, err := extractBinaryFromTarGz(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to extract update: %w", err)
	}
	defer os.Remove(tmpPath)

	// Replace current binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		// If rename fails (cross-device), try copy
		if err := copyFile(tmpPath, execPath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	}

	fmt.Printf("Successfully updated to version %s\n", latestVersion)
	return nil
}

func extractBinaryFromTarGz(r io.Reader) (string, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the leafpress binary
		if header.Typeflag == tar.TypeReg && (header.Name == "leafpress" || strings.HasSuffix(header.Name, "/leafpress")) {
			tmpFile, err := os.CreateTemp("", "leafpress-update-*")
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return "", err
			}
			tmpFile.Close()

			// Make executable
			if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("leafpress binary not found in archive")
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(githubAPIURL)
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
	// Format: leafpress-{version}-{os}-{arch}.tar.gz
	return fmt.Sprintf("leafpress-%s-%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
