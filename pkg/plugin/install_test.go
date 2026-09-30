package plugin

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallFromLocalFile(t *testing.T) {
	tempSrc := t.TempDir()
	tempDst := t.TempDir()

	srcFile := filepath.Join(tempSrc, "michibiki-provider-mydevice")
	if err := os.WriteFile(srcFile, []byte("#!/bin/sh\necho ok"), 0755); err != nil {
		t.Fatalf("failed to write test binary: %v", err)
	}

	installedPath, err := InstallFromLocalFile(srcFile, InstallOptions{
		TargetDir: tempDst,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("InstallFromLocalFile failed: %v", err)
	}

	if filepath.Base(installedPath) != "michibiki-provider-mydevice" {
		t.Errorf("unexpected installed file: %s", installedPath)
	}

	info, err := os.Stat(installedPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("installed file should be executable: %v", info.Mode())
	}
}
func TestExtractTarGz_Security(t *testing.T) {
	tempDst := t.TempDir()

	t.Run("rejects path traversal entry", func(t *testing.T) {
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gzw)

		data := []byte("#!/bin/sh\n")
		hdr := &tar.Header{
			Name: "../../../tmp/michibiki-provider-escape",
			Mode: 0755,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatalf("failed to write data: %v", err)
		}
		_ = tw.Close()
		_ = gzw.Close()

		_, err := extractTarGz(&buf, tempDst, true)
		if err == nil {
			t.Fatal("expected extractTarGz to reject path traversal entry")
		}
		if !strings.Contains(err.Error(), "insecure archive entry") && !strings.Contains(err.Error(), "escapes") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects symlinks", func(t *testing.T) {
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gzw)

		hdr := &tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     "michibiki-provider-symlink",
			Linkname: "/etc/passwd",
			Mode:     0777,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		_ = tw.Close()
		_ = gzw.Close()

		_, err := extractTarGz(&buf, tempDst, true)
		if err == nil {
			t.Fatal("expected extractTarGz to ignore or reject symlink")
		}
	})
}
func TestInstallFromURL_Security(t *testing.T) {
	tempDst := t.TempDir()

	t.Run("rejects unsupported URL schemes", func(t *testing.T) {
		badURLs := []string{
			"file:///etc/passwd",
			"ftp://example.com/plugin.tar.gz",
			"gopher://example.com/plugin",
			"data:text/plain;base64,SGVsbG8=",
		}
		for _, u := range badURLs {
			_, err := InstallFromURL(t.Context(), u, InstallOptions{TargetDir: tempDst})
			if err == nil {
				t.Fatalf("expected InstallFromURL to reject scheme for %q", u)
			}
			if !strings.Contains(err.Error(), "unsupported download scheme") {
				t.Fatalf("unexpected error message for %q: %v", u, err)
			}
		}
	})
}
