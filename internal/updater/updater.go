package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const ReleasesBase = "https://github.com/getkaze/helm/releases"

type CheckResult struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	UpdateURL string `json:"update_url"`
	Available bool   `json:"available"`
}

func Check(current string) (*CheckResult, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(ReleasesBase + "/latest")
	if err != nil {
		return nil, fmt.Errorf("fetch version: %w", err)
	}
	resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return nil, fmt.Errorf("fetch version: no redirect from /latest")
	}

	parts := strings.Split(loc, "/")
	latest := parts[len(parts)-1]

	return &CheckResult{
		Current:   current,
		Latest:    latest,
		UpdateURL: downloadURL(latest),
		Available: current != "dev" && IsNewer(latest, current),
	}, nil
}

func IsNewer(a, b string) bool {
	aParts, aOk := parseSemver(a)
	bParts, bOk := parseSemver(b)
	if !aOk || !bOk {
		return a != b
	}
	for i := 0; i < 3; i++ {
		if aParts[i] != bParts[i] {
			return aParts[i] > bParts[i]
		}
	}
	return false
}

func parseSemver(v string) ([3]int, bool) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var result [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		result[i] = n
	}
	return result, true
}

func Download(version string) (string, error) {
	url := downloadURL(version)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: status %d", resp.StatusCode)
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("find executable: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(exe), "helm-update-*")
	if err != nil {
		tmp, err = os.CreateTemp("", "helm-update-*")
		if err != nil {
			return "", fmt.Errorf("create temp: %w", err)
		}
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("write: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), 0755); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("chmod: %w", err)
	}

	return tmp.Name(), nil
}

func Replace(tmpPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err == nil {
		return nil
	}

	src, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("open update: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(exe, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create binary: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return fmt.Errorf("write binary: %w", err)
	}

	if err := dst.Close(); err != nil {
		return fmt.Errorf("close binary: %w", err)
	}

	os.Remove(tmpPath)
	return nil
}

func downloadURL(version string) string {
	return fmt.Sprintf("%s/download/%s/helm-%s-%s", ReleasesBase, version, runtime.GOOS, runtime.GOARCH)
}
