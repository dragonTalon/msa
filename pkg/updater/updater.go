// Package updater 提供 MSA 自更新能力
package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
)

const (
	// GitHub 仓库信息
	githubOwner = "dragonTalon"
	githubRepo  = "msa"

	// GitHub API 地址
	githubAPIURL = "https://api.github.com"
	// GitHub Releases 下载地址
	githubDownloadURL = "https://github.com"
)

// ReleaseInfo 包含 GitHub Release 信息
type ReleaseInfo struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

// Asset 包含 Release Asset 信息
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateResult 更新结果
type UpdateResult struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
	Message        string
}

// Updater 负责处理自更新逻辑
type Updater struct {
	currentVersion string
	currentCommit  string
	apiClient      *http.Client  // 用于 API 请求
	downloadClient *http.Client  // 用于文件下载
}

// New 创建新的 Updater 实例
func New(version, commit string) *Updater {
	return &Updater{
		currentVersion: version,
		currentCommit:  commit,
		apiClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		downloadClient: &http.Client{
			Timeout: 10 * time.Minute, // 下载大文件，设置较长的超时时间
		},
	}
}

// CheckLatestVersion 调用 GitHub Releases API 获取最新版本信息
// 注意：使用 /releases 而不是 /releases/latest 以获取包括预发布版本在内的所有版本
func (u *Updater) CheckLatestVersion() (*ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", githubAPIURL, githubOwner, githubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.apiClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误状态码: %d", resp.StatusCode)
	}

	var releases []ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("未找到任何 release")
	}

	// 返回第一个（最新的）release
	return &releases[0], nil
}

// CompareVersions 使用 semver 比较当前版本与最新版本
// 返回值: 1 表示 latest > current, 0 表示相等, -1 表示 latest < current
func (u *Updater) CompareVersions(latestTag string) (int, error) {
	currentVer, err := semver.NewVersion(u.getCurrentVersion())
	if err != nil {
		// 如果当前版本是 "dev"，认为需要更新
		if u.currentVersion == "dev" {
			return 1, nil
		}
		return 0, fmt.Errorf("解析当前版本失败: %w", err)
	}

	latestVer, err := semver.NewVersion(strings.TrimPrefix(latestTag, "v"))
	if err != nil {
		return 0, fmt.Errorf("解析最新版本失败: %w", err)
	}

	return latestVer.Compare(currentVer), nil
}

// buildDownloadURL 根据 OS/arch 构建下载 URL
func (u *Updater) buildDownloadURL(release *ReleaseInfo) (string, error) {
	assetName := u.getAssetName(release.TagName)

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("未找到匹配的资源文件: %s", assetName)
}

// downloadFile 下载文件到临时目录，返回文件路径
func (u *Updater) downloadFile(url string) (string, error) {
	log.Infof("正在下载: %s", url)

	resp, err := u.downloadClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载返回错误状态码: %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpDir, err := os.MkdirTemp("", "msa-update-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	tmpFile := filepath.Join(tmpDir, filepath.Base(url))
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer out.Close()

	// 显示下载进度
	total := resp.ContentLength
	written := int64(0)

	_, err = io.Copy(out, io.TeeReader(resp.Body, &progressWriter{
		total:   total,
		written: &written,
	}))
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Println() // 换行
	log.Info("下载完成")
	return tmpFile, nil
}

// progressWriter 用于显示下载进度
type progressWriter struct {
	total   int64
	written *int64
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	*p.written += int64(n)
	if p.total > 0 {
		percent := float64(*p.written) / float64(p.total) * 100
		fmt.Printf("\r下载进度: %.1f%% (%d/%d bytes)", percent, *p.written, p.total)
	} else {
		fmt.Printf("\r已下载: %d bytes", *p.written)
	}
	return n, nil
}

// verifyChecksum 校验下载文件的 SHA256
func (u *Updater) verifyChecksum(filePath string, release *ReleaseInfo) error {
	// 下载 checksums.txt
	var checksumURL string
	var targetAssetName string

	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			checksumURL = asset.BrowserDownloadURL
		}
		if asset.Name == u.getAssetName(release.TagName) {
			targetAssetName = asset.Name
		}
	}

	if checksumURL == "" {
		return fmt.Errorf("未找到 checksums.txt")
	}

	resp, err := u.downloadClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("下载 checksums.txt 失败: %w", err)
	}
	defer resp.Body.Close()

	checksums, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 checksums.txt 失败: %w", err)
	}

	// 解析 checksums.txt 查找目标文件的校验和
	var expectedChecksum string
	lines := strings.Split(string(checksums), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == targetAssetName {
			expectedChecksum = parts[0]
			break
		}
	}

	if expectedChecksum == "" {
		return fmt.Errorf("未在 checksums.txt 中找到 %s 的校验和", targetAssetName)
	}

	// 计算下载文件的校验和
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("计算校验和失败: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("校验和不匹配: 期望 %s, 实际 %s", expectedChecksum, actualChecksum)
	}

	log.Info("校验和验证通过")
	return nil
}

// replaceBinary 备份、替换、清理二进制文件
func (u *Updater) replaceBinary(newBinaryPath string) error {
	currentBinary, err := getBinaryPath()
	if err != nil {
		return fmt.Errorf("获取当前二进制路径失败: %w", err)
	}

	// Windows 不支持自动替换
	if runtime.GOOS == "windows" {
		fmt.Printf("\n下载完成！请手动替换:\n")
		fmt.Printf("1. 关闭当前 MSA 程序\n")
		fmt.Printf("2. 将 %s 解压\n", newBinaryPath)
		fmt.Printf("3. 将 msa.exe 复制到 %s\n", currentBinary)
		return nil
	}

	// 备份当前二进制
	backupPath := currentBinary + ".bak"
	log.Infof("备份当前二进制到: %s", backupPath)
	if err := os.Rename(currentBinary, backupPath); err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 替换二进制
	log.Info("正在替换二进制文件...")
	if err := os.Rename(newBinaryPath, currentBinary); err != nil {
		// 尝试恢复备份
		log.Errorf("替换失败，尝试恢复备份: %v", err)
		if restoreErr := os.Rename(backupPath, currentBinary); restoreErr != nil {
			log.Errorf("恢复备份也失败: %v", restoreErr)
		}
		return fmt.Errorf("替换二进制失败: %w", err)
	}

	// 清理备份
	if err := os.Remove(backupPath); err != nil {
		log.Warnf("删除备份文件失败: %v", err)
	}

	return nil
}

// extractBinary 从压缩包中提取二进制文件
func (u *Updater) extractBinary(archivePath string) (string, error) {
	tmpDir := filepath.Dir(archivePath)

	if strings.HasSuffix(archivePath, ".zip") {
		return u.extractZip(archivePath, tmpDir)
	}
	return u.extractTarGz(archivePath, tmpDir)
}

func (u *Updater) extractZip(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("打开 zip 文件失败: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "msa") && !strings.Contains(f.Name, "/") {
			destPath := filepath.Join(destDir, "msa")
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", fmt.Errorf("创建输出文件失败: %w", err)
			}
			defer out.Close()

			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("打开 zip 内容失败: %w", err)
			}
			defer rc.Close()

			if _, err := io.Copy(out, rc); err != nil {
				return "", fmt.Errorf("解压失败: %w", err)
			}
			return destPath, nil
		}
	}
	return "", fmt.Errorf("未在压缩包中找到 msa 二进制文件")
}

func (u *Updater) extractTarGz(tarGzPath, destDir string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", fmt.Errorf("打开 tar.gz 文件失败: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("创建 gzip reader 失败: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("读取 tar 失败: %w", err)
		}

		if strings.HasPrefix(hdr.Name, "msa") && hdr.Typeflag == tar.TypeReg {
			destPath := filepath.Join(destDir, "msa")
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return "", fmt.Errorf("创建输出文件失败: %w", err)
			}
			defer out.Close()

			if _, err := io.Copy(out, tr); err != nil {
				return "", fmt.Errorf("解压失败: %w", err)
			}
			return destPath, nil
		}
	}
	return "", fmt.Errorf("未在压缩包中找到 msa 二进制文件")
}

// Update 执行完整的更新流程
func (u *Updater) Update(checkOnly bool) (*UpdateResult, error) {
	result := &UpdateResult{
		CurrentVersion: u.currentVersion,
	}

	// 1. 检查最新版本
	log.Info("正在检查最新版本...")
	release, err := u.CheckLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("检查最新版本失败: %w", err)
	}

	result.LatestVersion = release.TagName

	// 2. 比较版本
	cmp, err := u.CompareVersions(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("比较版本失败: %w", err)
	}

	if cmp <= 0 {
		result.Message = "当前已是最新版本"
		return result, nil
	}

	result.Message = fmt.Sprintf("发现新版本: %s", release.TagName)

	// 如果只是检查，到这里就返回
	if checkOnly {
		return result, nil
	}

	// 3. 构建下载 URL
	downloadURL, err := u.buildDownloadURL(release)
	if err != nil {
		return nil, fmt.Errorf("构建下载 URL 失败: %w", err)
	}

	// 4. 下载文件
	archivePath, err := u.downloadFile(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	defer os.RemoveAll(filepath.Dir(archivePath))

	// 5. 校验文件
	if err := u.verifyChecksum(archivePath, release); err != nil {
		return nil, fmt.Errorf("校验失败: %w", err)
	}

	// 6. 解压
	binaryPath, err := u.extractBinary(archivePath)
	if err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	// 7. 替换二进制
	if err := u.replaceBinary(binaryPath); err != nil {
		return nil, fmt.Errorf("替换失败: %w", err)
	}

	result.Updated = true
	result.Message = fmt.Sprintf("更新成功！%s → %s", u.currentVersion, release.TagName)

	return result, nil
}

// getCurrentVersion 获取当前版本（去掉 v 前缀）
func (u *Updater) getCurrentVersion() string {
	return strings.TrimPrefix(u.currentVersion, "v")
}

// getAssetName 根据系统信息构建资源文件名
func (u *Updater) getAssetName(version string) string {
	// 文件名格式: msa_{version}_{os}_{arch}.{ext}
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	// version 需要去掉 v 前缀
	cleanVersion := strings.TrimPrefix(version, "v")

	return fmt.Sprintf("msa_%s_%s_%s.%s", cleanVersion, runtime.GOOS, runtime.GOARCH, ext)
}

// getBinaryPath 获取当前可执行文件路径
func getBinaryPath() (string, error) {
	return os.Executable()
}
