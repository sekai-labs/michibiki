package plugin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type InstallOptions struct {
	TargetDir string
	Force     bool
}

func DefaultPluginDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "michibiki", "plugins"), nil
}

func InstallFromLocalFile(srcPath string, opts InstallOptions) (string, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		var err error
		targetDir, err = DefaultPluginDir()
		if err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	baseName := filepath.Base(srcPath)
	if !strings.HasPrefix(baseName, "michibiki-provider-") {
		baseName = "michibiki-provider-" + baseName
	}

	destPath := filepath.Join(targetDir, baseName)
	if _, err := os.Stat(destPath); err == nil && !opts.Force {
		return "", fmt.Errorf("plugin already exists at %s (use --force to overwrite)", destPath)
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", err
	}

	return destPath, nil
}

func InstallFromGoModule(ctx context.Context, goPackage string, opts InstallOptions) (string, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		var err error
		targetDir, err = DefaultPluginDir()
		if err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "michibiki-plugin-build-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	pkgParts := strings.Split(goPackage, "/")
	rawName := pkgParts[len(pkgParts)-1]
	if atIdx := strings.Index(rawName, "@"); atIdx != -1 {
		rawName = rawName[:atIdx]
	}
	binaryName := rawName
	if !strings.HasPrefix(binaryName, "michibiki-provider-") {
		binaryName = "michibiki-provider-" + binaryName
	}

	cmd := exec.CommandContext(ctx, "go", "install", goPackage)
	cmd.Env = append(os.Environ(), "GOBIN="+tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go install failed: %w (output: %s)", err, string(output))
	}

	tempBinaries, err := os.ReadDir(tempDir)
	if err != nil || len(tempBinaries) == 0 {
		return "", errors.New("no binary produced by go install")
	}

	builtBinPath := filepath.Join(tempDir, tempBinaries[0].Name())
	destPath := filepath.Join(targetDir, binaryName)

	if _, err := os.Stat(destPath); err == nil && !opts.Force {
		return "", fmt.Errorf("plugin already exists at %s (use --force to overwrite)", destPath)
	}

	builtData, err := os.ReadFile(builtBinPath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destPath, builtData, 0755); err != nil {
		return "", err
	}

	return destPath, nil
}

func InstallFromURL(ctx context.Context, downloadURL string, opts InstallOptions) (string, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		var err error
		targetDir, err = DefaultPluginDir()
		if err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	urlLower := strings.ToLower(downloadURL)
	if strings.HasSuffix(urlLower, ".tar.gz") || strings.HasSuffix(urlLower, ".tgz") {
		return extractTarGz(resp.Body, targetDir, opts.Force)
	}
	if strings.HasSuffix(urlLower, ".zip") {
		return extractZip(resp.Body, targetDir, opts.Force)
	}

	parts := strings.Split(downloadURL, "/")
	rawName := parts[len(parts)-1]
	if qIdx := strings.Index(rawName, "?"); qIdx != -1 {
		rawName = rawName[:qIdx]
	}
	if !strings.HasPrefix(rawName, "michibiki-provider-") {
		rawName = "michibiki-provider-" + rawName
	}

	destPath := filepath.Join(targetDir, rawName)
	if _, err := os.Stat(destPath); err == nil && !opts.Force {
		return "", fmt.Errorf("plugin already exists at %s (use --force to overwrite)", destPath)
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, resp.Body); err != nil {
		return "", err
	}

	return destPath, nil
}

func extractTarGz(r io.Reader, targetDir string, force bool) (string, error) {
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

		baseName := filepath.Base(header.Name)
		if strings.HasPrefix(baseName, "michibiki-provider-") && !header.FileInfo().IsDir() {
			destPath := filepath.Join(targetDir, baseName)
			if _, err := os.Stat(destPath); err == nil && !force {
				return "", fmt.Errorf("plugin already exists at %s (use --force)", destPath)
			}
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			defer f.Close()
			if _, err := io.Copy(f, tr); err != nil {
				return "", err
			}
			return destPath, nil
		}
	}
	return "", errors.New("no executable starting with 'michibiki-provider-' found in archive")
}

func extractZip(r io.Reader, targetDir string, force bool) (string, error) {
	tempZip, err := os.CreateTemp("", "plugin-*.zip")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempZip.Name())
	defer tempZip.Close()

	if _, err := io.Copy(tempZip, r); err != nil {
		return "", err
	}

	stat, err := tempZip.Stat()
	if err != nil {
		return "", err
	}

	zr, err := zip.NewReader(tempZip, stat.Size())
	if err != nil {
		return "", err
	}

	for _, f := range zr.File {
		baseName := filepath.Base(f.Name)
		if strings.HasPrefix(baseName, "michibiki-provider-") && !f.FileInfo().IsDir() {
			destPath := filepath.Join(targetDir, baseName)
			if _, err := os.Stat(destPath); err == nil && !force {
				return "", fmt.Errorf("plugin already exists at %s (use --force)", destPath)
			}
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			df, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			defer df.Close()

			if _, err := io.Copy(df, rc); err != nil {
				return "", err
			}
			return destPath, nil
		}
	}
	return "", errors.New("no executable starting with 'michibiki-provider-' found in zip")
}
