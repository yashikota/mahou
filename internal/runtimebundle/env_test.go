package runtimebundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurePathPrependsProcessConfig(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "runtime")
	configDir := filepath.Join(string(filepath.Separator), "policy")

	got := configurePath(root, configDir)
	want := configDir + string(os.PathListSeparator) + filepath.Join(root, "etc", "ImageMagick-7")
	if got != want {
		t.Fatalf("configurePath() = %q, want %q", got, want)
	}
}
