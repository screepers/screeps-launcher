package install

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// tarEntry describes a single member of a test archive.
type tarEntry struct {
	header tar.Header
	body   string
}

// writeTarGz builds a .tar.gz with the given members and returns its path.
func writeTarGz(t *testing.T, name string, entries []tarEntry) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)
	for _, entry := range entries {
		header := entry.header
		if entry.body != "" {
			header.Size = int64(len(entry.body))
		}
		if err := tw.WriteHeader(&header); err != nil {
			t.Fatalf("write header %s: %v", header.Name, err)
		}
		if entry.body != "" {
			if _, err := tw.Write([]byte(entry.body)); err != nil {
				t.Fatalf("write body %s: %v", header.Name, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}
	return path
}

// newDest creates a destination directory and returns it.
func newDest(t *testing.T) string {
	t.Helper()

	dest := filepath.Join(t.TempDir(), "dest")
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatalf("create dest: %v", err)
	}
	return dest
}

// symlinksSupported reports whether the platform lets this test create a symlink.
func symlinksSupported(t *testing.T) bool {
	t.Helper()

	target := filepath.Join(t.TempDir(), "target")
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(target, link); err != nil {
		return false
	}
	os.Remove(link)
	return true
}

func TestExtractTarGzRejectsEscapingMembers(t *testing.T) {
	regular := func(name string) tarEntry {
		return tarEntry{tar.Header{Name: name, Mode: 0644, Typeflag: tar.TypeReg}, "pwned"}
	}

	tables := []struct {
		name    string
		entries []tarEntry
		escaped string
	}{
		{
			name:    "member above the destination",
			entries: []tarEntry{regular("kept.txt"), regular("../pwned.txt")},
			escaped: "../pwned.txt",
		},
		{
			name:    "traversal hidden behind a directory name",
			entries: []tarEntry{regular("kept.txt"), regular("sub/../../pwned.txt")},
			escaped: "../pwned.txt",
		},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			dest := newDest(t)
			archive := writeTarGz(t, "evil.tar.gz", table.entries)

			if err := extractTarGz(dest, archive); err == nil {
				t.Errorf("extractTarGz accepted an archive writing outside the destination directory")
			}

			escaped := filepath.Join(dest, table.escaped)
			if _, err := os.Lstat(escaped); err == nil {
				t.Errorf("member escaped the destination directory: %s was created", escaped)
			}
		})
	}
}

func TestExtractTarGzRejectsEscapingSymlink(t *testing.T) {
	if !symlinksSupported(t) {
		t.Skip("platform does not allow creating symlinks")
	}

	dest := newDest(t)
	entries := []tarEntry{
		{tar.Header{Name: "evil", Typeflag: tar.TypeSymlink, Linkname: filepath.Join("..", "..")}, ""},
		{tar.Header{Name: "evil/pwned.txt", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"},
	}
	archive := writeTarGz(t, "symlink.tar.gz", entries)

	if err := extractTarGz(dest, archive); err == nil {
		t.Errorf("extractTarGz accepted a symlink pointing outside the destination directory")
	}

	escaped := filepath.Join(dest, "..", "pwned.txt")
	if _, err := os.Lstat(escaped); err == nil {
		t.Errorf("member written through an escaping symlink: %s was created", escaped)
	}
}

func TestExtractTarGzExtractsContainedArchive(t *testing.T) {
	entries := []tarEntry{
		{tar.Header{Name: "sub", Mode: 0755, Typeflag: tar.TypeDir}, ""},
		{tar.Header{Name: "kept.txt", Mode: 0644, Typeflag: tar.TypeReg}, "kept"},
		{tar.Header{Name: "sub/nested.txt", Mode: 0644, Typeflag: tar.TypeReg}, "nested"},
	}

	dest := newDest(t)
	archive := writeTarGz(t, "good.tar.gz", entries)

	if err := extractTarGz(dest, archive); err != nil {
		t.Fatalf("extractTarGz rejected a contained archive: %v", err)
	}

	for _, name := range []string{"kept.txt", "sub/nested.txt"} {
		path := filepath.Join(dest, name)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("expected %s to be extracted: %v", path, err)
			continue
		}
		if len(body) == 0 {
			t.Errorf("expected %s to have content", path)
		}
	}
}

// TestWithin keeps the separator guard honest: "dir" must not be treated as a
// prefix of "dir-other".
func TestWithin(t *testing.T) {
	dest := filepath.Join("root", "dest")

	tables := []struct {
		path string
		ok   bool
	}{
		{dest, true},
		{filepath.Join(dest, "kept.txt"), true},
		{filepath.Join(dest, "sub", "kept.txt"), true},
		{filepath.Dir(dest), false},
		{filepath.Join(filepath.Dir(dest), "pwned.txt"), false},
		{dest + "ination", false},
		{filepath.Join(dest+"ination", "pwned.txt"), false},
	}

	for _, table := range tables {
		if got := within(dest, table.path); got != table.ok {
			t.Errorf("within(%s, %s) = %v, want %v", dest, table.path, got, table.ok)
		}
	}
}

func TestWithinLink(t *testing.T) {
	dest := t.TempDir()
	sub := filepath.Join(dest, "sub")
	parent := filepath.Dir(dest)
	absoluteOutside := filepath.Join(parent, "pwned.txt")

	tables := []struct {
		name     string
		dir      string
		linkname string
		ok       bool
	}{
		{"link to a sibling file", dest, "kept.txt", true},
		{"link to a file above the link location", sub, filepath.Join("..", "kept.txt"), true},
		{"link to a directory above the link location", sub, "..", true},
		{"link to the parent of the destination", dest, filepath.Join("..", ".."), false},
		{"link escaping through nested parents", sub, filepath.Join("..", "..", "pwned.txt"), false},
		{"absolute link outside the destination", dest, absoluteOutside, false},
		{"absolute link inside the destination", dest, filepath.Join(dest, "kept.txt"), true},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			if got := withinLink(dest, table.dir, table.linkname); got != table.ok {
				t.Errorf("withinLink(%s, %s, %s) = %v, want %v", dest, table.dir, table.linkname, got, table.ok)
			}
		})
	}
}
