package magick_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
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
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
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
	if info.Width != 4 || info.Height != 4 {
		t.Errorf("expected 4x4, got %dx%d", info.Width, info.Height)
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

	err := magick.Resize(input, output, 2, magick.ConvertOptions{})
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}

	info, err := magick.Identify(output)
	if err != nil {
		t.Fatalf("Identify resized: %v", err)
	}
	if info.Width != 2 {
		t.Errorf("expected width 2, got %d", info.Width)
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
