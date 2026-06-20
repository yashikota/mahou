package magick_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashikota/magick-go/internal/magick"
	"github.com/yashikota/magick-go/internal/runtimebundle"
)

func setup(t *testing.T) {
	t.Helper()
	bundle, err := runtimebundle.Ensure()
	if err != nil {
		t.Skipf("runtime bundle not available: %v", err)
	}
	policyDir, err := runtimebundle.ApplyPolicy(false)
	if err != nil {
		t.Fatalf("apply policy: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(policyDir) })
	runtimebundle.ConfigureEnvironment(bundle.Root, policyDir)
	if _, err := magick.Load(bundle.Root); err != nil {
		t.Fatalf("load magick: %v", err)
	}
}

func createTestPNG(t *testing.T, path string) {
	t.Helper()
	// Minimal valid 1x1 red PNG (67 bytes)
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // 8-bit RGB
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x36, 0x28, 0x19,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, // IEND chunk
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(path, png, 0o644); err != nil {
		t.Fatalf("write test png: %v", err)
	}
}

func TestIdentify(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	createTestPNG(t, input)

	info, err := magick.Identify(input)
	if err != nil {
		t.Fatalf("Identify: %v", err)
	}
	if info.Width != 1 || info.Height != 1 {
		t.Errorf("expected 1x1, got %dx%d", info.Width, info.Height)
	}
	if info.Format != "PNG" {
		t.Errorf("expected format PNG, got %s", info.Format)
	}
}

func TestConvertPNGToJPEG(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "output.jpg")
	createTestPNG(t, input)

	err := magick.Convert(input, output, magick.ConvertOptions{Quality: 85})
	if err != nil {
		t.Fatalf("Convert PNG->JPEG: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Error("output is not a valid JPEG (missing SOI marker)")
	}
}

func TestConvertPNGToWebP(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "output.webp")
	createTestPNG(t, input)

	err := magick.Convert(input, output, magick.ConvertOptions{})
	if err != nil {
		t.Fatalf("Convert PNG->WebP: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) < 4 || string(data[0:4]) != "RIFF" {
		t.Error("output is not a valid WebP (missing RIFF header)")
	}
}

func TestResize(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "resized.png")
	createTestPNG(t, input)

	err := magick.Resize(input, output, 1, magick.ConvertOptions{})
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}

	info, err := magick.Identify(output)
	if err != nil {
		t.Fatalf("Identify resized: %v", err)
	}
	if info.Width != 1 {
		t.Errorf("expected width 1, got %d", info.Width)
	}
}

func TestFormats(t *testing.T) {
	setup(t)
	formats := magick.Formats()
	if len(formats) == 0 {
		t.Fatal("Formats returned empty list")
	}
	required := []string{"JPEG", "PNG", "WEBP"}
	for _, f := range required {
		found := false
		for _, got := range formats {
			if got == f {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required format %s not found in supported formats", f)
		}
	}
}
