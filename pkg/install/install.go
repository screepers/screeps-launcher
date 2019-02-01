package install

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"encoding/json"	
	"io/ioutil"
	"net/http"
	"time"
	"strings"
	"github.com/cavaliercoder/grab"
)

type NodeVersion struct {
	Version string
	Date string
	Files []string
	Npm string
	V8 string
	Uv string
	Zlib string
	Openssl string
	Modules string
	Lts interface{}
}

func download(dest string, url string) error {
	client := grab.NewClient()
	req, _ := grab.NewRequest(dest, url)
	log.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	log.Printf("  %v\n", resp.HTTPResponse.Status)
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	Loop:
	for {
		select {
		case <-t.C:
			log.Printf("  downloaded %v/%v bytes (%.2f%%)\n", 
				resp.BytesComplete(),
				resp.Size,
				100 * resp.Progress())
		case <-resp.Done:
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

func InstallMongo() error {
	return nil
}
func InstallRedis() error {
	return nil
}

func InstallNode(version string) error {
	url := "https://nodejs.org/dist/index.json"

	nodeClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "screeps-launcher")

	res, getErr := nodeClient.Do(req)
	if getErr != nil {
		return getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return readErr
	}

	versions := make([]NodeVersion, 0)
	err = json.Unmarshal(body, &versions)
	if err != nil {
		fmt.Println(err)
		return err
	}

	ver := version
	if version[0:1] != "v" {
		ver = getWantedVersion(version, versions)
	}
	if ver == "" {
		return fmt.Errorf("Could not find node version: %s", version)
	}

	file := getFileName(runtime.GOOS, runtime.GOARCH, ver)

	url = fmt.Sprintf("https://nodejs.org/dist/%s/%s", ver, file)	

	download(fmt.Sprintf("./%s", file), url)

	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	ExtractTarGz("deps", f)
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