package install

import "testing"

func TestGetFileName(t *testing.T) {
	tables := []struct {
		os      string
		arch    string
		version string
		ret     string
	}{
		{"windows", "386", "VER", "node-VER-win-x86.zip"},
		{"windows", "amd64", "VER", "node-VER-win-x64.zip"},
		{"linux", "amd64", "VER", "node-VER-linux-x64.tar.gz"},
		{"linux", "arm", "VER", "node-VER-linux-armv6l.tar.gz"},
		{"linux", "arm64", "VER", "node-VER-linux-arm64.tar.gz"},
	}
	for _, table := range tables {
		ret := getFileName(table.os, table.arch, table.version)
		if ret != table.ret {
			t.Errorf("getFileName(%s, %s, %v) was incorrect, got: %s, want: %s", table.os, table.arch, table.version, ret, table.ret)
		}
	}
}
