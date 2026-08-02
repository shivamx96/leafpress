package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestShouldInstallVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		force   bool
		want    bool
		wantErr bool
	}{
		{name: "upgrade", current: "v1.0.0", latest: "v1.1.0", want: true},
		{name: "same", current: "v1.0.0", latest: "v1.0.0"},
		{name: "forced reinstall", current: "v1.0.0", latest: "v1.0.0", force: true, want: true},
		{name: "downgrade", current: "v1.1.0", latest: "v1.0.0", wantErr: true},
		{name: "development build", current: "dev", latest: "v1.0.0", want: true},
		{name: "invalid release", current: "v1.0.0", latest: "latest", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shouldInstallVersion(tt.current, tt.latest, tt.force)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("install = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAssetNameForPlatform(t *testing.T) {
	if got := getAssetNameFor("v1.2.3", "linux", "amd64"); got != "leafpress-v1.2.3-linux-amd64.tar.gz" {
		t.Errorf("linux asset = %q", got)
	}
	if got := getAssetNameFor("v1.2.3", "windows", "arm64"); got != "leafpress-v1.2.3-windows-arm64.zip" {
		t.Errorf("windows asset = %q", got)
	}
}

func TestVerifyAssetChecksum(t *testing.T) {
	archive := []byte("verified archive")
	sum := sha256.Sum256(archive)
	assetName := "leafpress-v1.2.3-linux-amd64.tar.gz"
	checksums := []byte(hex.EncodeToString(sum[:]) + "  " + assetName + "\n")

	if err := verifyAssetChecksum(archive, checksums, assetName); err != nil {
		t.Fatalf("valid checksum rejected: %v", err)
	}
	if err := verifyAssetChecksum([]byte("tampered"), checksums, assetName); err == nil {
		t.Fatal("tampered archive passed checksum verification")
	}
	if err := verifyAssetChecksum(archive, checksums, "other.tar.gz"); err == nil {
		t.Fatal("missing checksum entry was accepted")
	}
}

func TestExtractBinaryFromReleaseArchives(t *testing.T) {
	want := []byte("leafpress binary")

	var tarBytes bytes.Buffer
	gzw := gzip.NewWriter(&tarBytes)
	tw := tar.NewWriter(gzw)
	if err := tw.WriteHeader(&tar.Header{Name: "dist/leafpress", Mode: 0755, Size: int64(len(want)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(want); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}

	var zipBytes bytes.Buffer
	zw := zip.NewWriter(&zipBytes)
	zf, err := zw.Create("leafpress.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zf.Write(want); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	for name, test := range map[string]struct {
		archive   []byte
		assetName string
	}{
		"tar.gz": {archive: tarBytes.Bytes(), assetName: "leafpress-v1.2.3-linux-amd64.tar.gz"},
		"zip":    {archive: zipBytes.Bytes(), assetName: "leafpress-v1.2.3-windows-amd64.zip"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := extractBinary(test.archive, test.assetName)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("binary = %q, want %q", got, want)
			}
		})
	}
}

func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	execPath := filepath.Join(dir, "leafpress")
	if err := os.WriteFile(execPath, []byte("old"), 0751); err != nil {
		t.Fatal(err)
	}
	if err := replaceExecutable(execPath, []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("executable = %q, want new", got)
	}
	info, err := os.Stat(execPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0751 {
		t.Errorf("mode = %o, want 751", info.Mode().Perm())
	}
}
