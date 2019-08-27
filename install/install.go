package install

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"
)

// NodeVersion version struct
type NodeVersion struct {
	Version string
	Date    string
	Files   []string
	Npm     string
	V8      string
	Uv      string
	Zlib    string
	Openssl string
	Modules string
	Lts     interface{}
}

func download(dest string, url string) error {
	client := grab.NewClient()
	req, _ := grab.NewRequest(dest, url)
	log.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	log.Printf("  %v\n", resp.HTTPResponse.Status)
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("  downloaded %v/%v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())
		case <-resp.Done:
			log.Printf("  downloaded %v/%v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())
			break Loop
		}
	}
	if err := resp.Err(); err != nil {
		log.Printf("Download failed: %v\n", err)
		return err
	}
	log.Printf("Download completed")
	return nil
}

// Mongo installs local mongo (STUB)
func Mongo() error {
	return nil
}

// Redis installs local redis (STUB)
func Redis() error {
	return nil
}

// Node installs local nodejs
func Node(version string) error {
	ver, err := GetNodeVersion(version)
	if err != nil {
		return err
	}

	file := getFileName(runtime.GOOS, runtime.GOARCH, ver)

	url := fmt.Sprintf("https://nodejs.org/dist/%s/%s", ver, file)
	download(file, url)
	log.Print(file)

	if runtime.GOOS == "windows" {
		extractZip("deps", file)
	} else {
		extractTarGz("deps", file)
	}
	os.Remove(file)
	name := file
	if runtime.GOOS == "windows" {
		name = strings.TrimSuffix(name, ".zip")
	} else {
		name = strings.TrimSuffix(name, ".tar.gz")
	}
	log.Print(name)
	err = os.Rename(fmt.Sprintf("deps/%s", name), "deps/node")
	if err != nil {
		return err
	}
	return nil
}

// Yarn installs YarnJS
func Yarn() error {
	type asset struct {
		URL  string `json:"browser_download_url"`
		Name string
	}
	type releases struct {
		Assets []asset
	}

	url := "https://api.github.com/repos/yarnpkg/yarn/releases/latest"

	yarnClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "screeps-launcher")

	res, getErr := yarnClient.Do(req)
	if getErr != nil {
		return getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return readErr
	}

	rel := releases{}
	err = json.Unmarshal(body, &rel)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var file string
	for _, asset := range rel.Assets {
		if strings.HasSuffix(asset.Name, ".tar.gz") {
			file = asset.Name
			download(file, asset.URL)
			break
		}
	}

	err = extractTarGz("deps", file)
	if err != nil {
		return err
	}
	os.Remove(file)
	name := strings.TrimSuffix(file, ".tar.gz")
	log.Print(name)
	err = os.Rename(fmt.Sprintf("deps/%s", name), "deps/yarn")
	if err != nil {
		return err
	}
	return nil
}

// WindowsBuildTools Installs Windows Build tools (Python, vc++ compiler, etc)
func WindowsBuildTools() error {
	cmd := exec.Command(NodePath, NpmPath, "--silent", "--no-audit", "install", "-g", "windows-build-tools")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func getFileName(os string, arch string, version string) string {
	ext := "tar.gz"
	switch os {
	case "windows":
		os = "win"
		ext = "zip"
	}
	switch arch {
	case "386":
		arch = "x86"
	case "amd64":
		arch = "x64"
	case "arm":
		arch = "armv6l"
	}
	return fmt.Sprintf("node-%s-%s-%s.%s", version, os, arch, ext)
}

// GetNodeVersion returns the version for the provided string
func GetNodeVersion(version string) (string, error) {
	url := "https://nodejs.org/dist/index.json"

	nodeClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "screeps-launcher")

	res, getErr := nodeClient.Do(req)
	if getErr != nil {
		return "", getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	versions := make([]NodeVersion, 0)
	err = json.Unmarshal(body, &versions)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	ver := version
	if version[0:1] != "v" {
		ver = getWantedVersion(version, versions)
	}
	if ver == "" {
		return "", fmt.Errorf("Could not find node version: %s", version)
	}
	return ver, nil
}

func getWantedVersion(version string, versions []NodeVersion) string {
	for _, ver := range versions {
		switch ver.Lts.(type) {
		case string:
			if ver.Lts.(string) == version {
				return ver.Version
			}
		}
	}
	return ""
}
