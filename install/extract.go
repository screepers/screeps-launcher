package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"fmt"
	"path/filepath"
	"strings"
)

func extractTarGz(dest string, src string) error {
	r, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer r.Close()
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()
	
	tr := tar.NewReader(gzr)
	
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		target := filepath.Join(dest, header.Name)
		dir := filepath.Dir(target)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.Mkdir(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			err := os.Symlink(header.Linkname, target)
			if err != nil {
				return err
			}
		}
	}
}

func extractZip(dest string, src string) error {
	zipfile, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	for _, f := range zipfile.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest) + string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", target)
		}
		dir := filepath.Dir(target)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		if f.FileInfo().IsDir() {
			if _, err := os.Stat(target); err != nil {
				if err := os.Mkdir(target, 0755); err != nil {
					return err
				}
			}
		} else {
			w, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			r, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, r);
			w.Close()
			r.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
