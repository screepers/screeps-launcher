package recovery

import "io/ioutil"

import "path/filepath"

import "os"

type fileData map[string][]byte

func (r *Recovery) filesBackup() (fileData, error) {
	data := fileData{}
	dirs := []string{
		"assets",
		r.config.LocalMods,
	}
	files := []string{
		"config.yml",
		"db.json",
		"market.yml",
	}
	dirs = append(dirs, r.config.Backup.Dirs...)
	files = append(files, r.config.Backup.Files...)
	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
	}
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if err == nil {
			data[file] = bytes
		}
	}
	return data, nil
}

func (r *Recovery) filesRestore(data fileData) error {
	for file, bytes := range data {
		dir := filepath.Dir(file)
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		ioutil.WriteFile(file, bytes, 0644)
	}
	return nil
}
