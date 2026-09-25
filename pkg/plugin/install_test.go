package plugin

import (
	"os"
	"path/filepath"
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
