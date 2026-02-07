// Package upgrade provides functionality for self-updating the dotenv-tui binary.
package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	repoOwner       = "jellydn"
	repoName        = "dotenv-tui"
	githubAPIURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	downloadBaseURL = "https://github.com/" + repoOwner + "/" + repoName + "/releases/download"
)

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
}

// Upgrade performs the upgrade to the latest version.
func Upgrade(currentVersion string) error {
	latestVersion, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	if currentVersion == "dev" {
		fmt.Printf("Current version: dev\n")
		fmt.Printf("Latest version: %s\n", latestVersion)
		fmt.Println("Cannot upgrade dev build. Please install using go install or build from source.")
		return nil
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version: %s\n", latestVersion)

	if currentVersion == latestVersion {
		fmt.Println("Already up to date!")
		return nil
	}

	fmt.Printf("Upgrade to %s? [y/N] ", latestVersion)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.ToLower(response)
	if response != "y" && response != "yes" {
		fmt.Println("Upgrade cancelled.")
		return nil
	}

	osType, arch := detectPlatform()

	binaryName := fmt.Sprintf("dotenv-tui-%s-%s", osType, arch)
	if osType == "windows" {
		binaryName += ".exe"
	}
	checksumName := binaryName + ".sha256"

	downloadURL := fmt.Sprintf("%s/v%s/%s", downloadBaseURL, latestVersion, binaryName)
	checksumURL := fmt.Sprintf("%s/v%s/%s", downloadBaseURL, latestVersion, checksumName)

	fmt.Printf("Downloading %s...\n", binaryName)

	tmpFile, tmpChecksum, err := downloadBinaryAndChecksum(downloadURL, checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()
	if tmpChecksum != "" {
		defer func() { _ = os.Remove(tmpChecksum) }()
	}

	if tmpChecksum != "" {
		fmt.Println("Verifying checksum...")
		if err := verifyChecksum(tmpFile, tmpChecksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("Checksum verified!")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := replaceBinary(tmpFile, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Successfully upgraded to %s!\n", latestVersion)
	return nil
}

func detectPlatform() (string, string) {
	osType := runtime.GOOS
	arch := runtime.GOARCH

	switch osType {
	case "linux", "darwin", "windows":
	default:
		osType = "linux"
	}

	switch arch {
	case "amd64", "386", "arm", "arm64":
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	default:
		arch = "amd64"
	}

	return osType, arch
}

// getLatestVersion fetches the latest release version from GitHub.
func getLatestVersion() (string, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag name in release")
	}

	return release.TagName, nil
}

func downloadBinaryAndChecksum(binaryURL, checksumURL string) (string, string, error) {
	binaryFile, err := downloadFile(binaryURL, "dotenv-tui-upgrade-*")
	if err != nil {
		return "", "", err
	}

	if err := os.Chmod(binaryFile, 0755); err != nil {
		_ = os.Remove(binaryFile)
		return "", "", err
	}

	checksumFile, err := downloadFile(checksumURL, "dotenv-tui-upgrade-checksum-*")
	if err != nil {
		fmt.Println("Warning: Checksum file not available, skipping verification")
		return binaryFile, "", nil
	}

	return binaryFile, checksumFile, nil
}

// downloadFile downloads a file from the given URL and saves it to a temp file.
func downloadFile(url, pattern string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return "", err
	}

	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func verifyChecksum(binaryPath, checksumPath string) error {
	expectedChecksum, err := readChecksumFile(checksumPath)
	if err != nil {
		return err
	}

	actualChecksum, err := calculateFileSHA256(binaryPath)
	if err != nil {
		return err
	}

	if expectedChecksum != actualChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func readChecksumFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty checksum file")
	}
	return fields[0], nil
}

// calculateFileSHA256 calculates the SHA256 hash of a file.
func calculateFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func replaceBinary(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		if err := copyFile(src, dst); err != nil {
			return err
		}
		return os.Remove(src)
	}
	return nil
}

// copyFile copies the contents of src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
