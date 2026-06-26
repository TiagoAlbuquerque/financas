package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// AppVersion is the current version of the application
const AppVersion = "1.0.0"

// UpdateInfo holds information about a potential update
type UpdateInfo struct {
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	Body        string `json:"body"`
	DownloadURL string `json:"downloadUrl"`
	AssetName   string `json:"assetName"`
	AssetID     int64  `json:"assetId"`
}

// GitHubRelease represents the structure returned by GitHub Release API
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Body    string        `json:"body"`
	Assets  []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a file in a GitHub Release
type GitHubAsset struct {
	Name               string `json:"name"`
	ID                 int64  `json:"id"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string `json:"url"` // api url for downloading
}

// CheckForUpdate checks if there is a newer version available on GitHub
func CheckForUpdate() (*UpdateInfo, error) {
	repo := "TiagoAlbuquerque/financas"

	// 1. Try public/token HTTP API first
	info, err := getLatestReleaseAPI(repo)
	if err == nil {
		return info, nil
	}

	// 2. Fallback: try local gh CLI
	info, err = getLatestReleaseGH(repo)
	if err == nil {
		return info, nil
	}

	return nil, fmt.Errorf("não foi possível verificar atualizações: %w", err)
}

func getLatestReleaseAPI(repo string) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "FinancasPersonalApp-Updater")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Optional token from env
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return parseReleaseToUpdateInfo(&release)
}

func getLatestReleaseGH(repo string) (*UpdateInfo, error) {
	// Check if gh CLI is in path
	_, err := exec.LookPath("gh")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("gh", "release", "view", "--repo", repo, "--json", "tagName,body,assets")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh cli error: %s", stderr.String())
	}

	// Parsing gh CLI JSON format (which uses camelCase keys)
	var ghResult struct {
		TagName string `json:"tagName"`
		Body    string `json:"body"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"assets"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &ghResult); err != nil {
		return nil, err
	}

	// Map to our standard struct
	var release GitHubRelease
	release.TagName = ghResult.TagName
	release.Body = ghResult.Body
	for _, a := range ghResult.Assets {
		release.Assets = append(release.Assets, GitHubAsset{
			Name:               a.Name,
			BrowserDownloadURL: a.URL,
		})
	}

	return parseReleaseToUpdateInfo(&release)
}

func parseReleaseToUpdateInfo(release *GitHubRelease) (*UpdateInfo, error) {
	latestVer := release.TagName
	if !isNewerVersion(AppVersion, latestVer) {
		return &UpdateInfo{Available: false, Version: latestVer}, nil
	}

	// Find the correct asset for current platform
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "financas.exe"
	} else {
		assetName = "financas"
	}

	var matchedAsset *GitHubAsset
	for _, asset := range release.Assets {
		// Matches exact names or matches platform names (e.g. financas-linux, financas-windows.exe)
		nameLower := strings.ToLower(asset.Name)
		if asset.Name == assetName ||
			(runtime.GOOS == "windows" && strings.HasSuffix(nameLower, ".exe") && strings.Contains(nameLower, "financas")) ||
			(runtime.GOOS == "linux" && !strings.HasSuffix(nameLower, ".exe") && strings.Contains(nameLower, "financas") && !strings.Contains(nameLower, ".zip") && !strings.Contains(nameLower, ".tar")) {
			matchedAsset = &asset
			break
		}
	}

	if matchedAsset == nil {
		return nil, fmt.Errorf("nenhum binário compatível encontrado na release %s para %s", latestVer, runtime.GOOS)
	}

	return &UpdateInfo{
		Available:   true,
		Version:     latestVer,
		Body:        release.Body,
		DownloadURL: matchedAsset.BrowserDownloadURL,
		AssetName:   matchedAsset.Name,
		AssetID:     matchedAsset.ID,
	}, nil
}

func isNewerVersion(current, latest string) bool {
	c := cleanVersion(current)
	l := cleanVersion(latest)

	cParts := strings.Split(c, ".")
	lParts := strings.Split(l, ".")

	for i := 0; i < len(cParts) && i < len(lParts); i++ {
		cNum, _ := strconv.Atoi(cParts[i])
		lNum, _ := strconv.Atoi(lParts[i])
		if lNum > cNum {
			return true
		}
		if cNum > lNum {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

func cleanVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	return v
}

// DownloadAndInstallUpdate downloads the binary and installs it
func DownloadAndInstallUpdate(info *UpdateInfo) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "financas-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempExePath := filepath.Join(tempDir, info.AssetName)

	// Try to download using local gh CLI first (handles private repo automatically)
	downloadedViaGH := false
	_, errGH := exec.LookPath("gh")
	if errGH == nil {
		cmd := exec.Command("gh", "release", "download", info.Version, "--repo", "TiagoAlbuquerque/financas", "--pattern", info.AssetName, "--dir", tempDir, "--clobber")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			downloadedViaGH = true
		}
	}

	// Fallback to API direct download if CLI failed
	if !downloadedViaGH {
		err = downloadAssetAPI(info, tempExePath)
		if err != nil {
			return fmt.Errorf("falha ao baixar binário: %w", err)
		}
	}

	// Replace active binary
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)

	err = os.Rename(exePath, oldPath)
	if err != nil {
		return fmt.Errorf("falha ao renomear executável ativo: %w", err)
	}

	err = copyFile(tempExePath, exePath)
	if err != nil {
		// Rollback on failure
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("falha ao copiar novo executável: %w", err)
	}

	if runtime.GOOS != "windows" {
		err = os.Chmod(exePath, 0755)
		if err != nil {
			return fmt.Errorf("falha ao atribuir permissão de execução: %w", err)
		}
	}

	return nil
}

func downloadAssetAPI(info *UpdateInfo, dstPath string) error {
	var downloadURL string
	var reqHeaders = make(map[string]string)

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" && info.AssetID > 0 {
		// Private repo API download
		downloadURL = fmt.Sprintf("https://api.github.com/repos/TiagoAlbuquerque/financas/releases/assets/%d", info.AssetID)
		reqHeaders["Accept"] = "application/octet-stream"
		reqHeaders["Authorization"] = "Bearer " + token
	} else {
		// Public repo direct download
		downloadURL = info.DownloadURL
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "FinancasPersonalApp-Updater")
	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download status: %s", resp.Status)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// CleanupOldBinary removes any .old files left behind on update
func CleanupOldBinary() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		_ = os.Remove(oldPath)
	}
}

// TriggerRestart spawns the new binary and terminates the current process
func TriggerRestart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exePath)
	if runtime.GOOS != "windows" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
