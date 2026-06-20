package magick_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yashikota/magick-go/internal/magick"
	"github.com/yashikota/magick-go/internal/runtimebundle"
)

var testBundle *runtimebundle.Bundle

func setup(t *testing.T) {
	t.Helper()
	if testBundle == nil {
		bundle, err := runtimebundle.Ensure()
		if err != nil {
			t.Skipf("runtime bundle not available: %v", err)
		}
		policyDir, err := runtimebundle.ApplyPolicy(true)
		if err != nil {
			t.Fatalf("apply policy: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(policyDir) })
		runtimebundle.ConfigureEnvironment(bundle.Root, policyDir)
		if _, err := magick.Load(bundle.Root); err != nil {
			t.Fatalf("load magick: %v", err)
		}
		testBundle = bundle
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

func TestConvertFormats(t *testing.T) {
	setup(t)

	formats := []struct {
		name   string
		ext    string
		magic  []byte
		linux  bool
		darwin bool
	}{
		{"JPEG", "jpg", []byte{0xFF, 0xD8}, true, true},
		{"PNG", "png", []byte{0x89, 0x50, 0x4E, 0x47}, true, true},
		{"WebP", "webp", []byte("RIFF"), true, true},
		{"TIFF", "tiff", nil, true, true},
		{"GIF", "gif", []byte("GIF8"), true, true},
		{"BMP", "bmp", []byte("BM"), true, true},
		{"HEIC", "heic", nil, true, true},
		{"AVIF", "avif", nil, true, true},
		{"JXL", "jxl", nil, true, true},
		{"PDF", "pdf", []byte("%PDF"), true, true},
		{"EXR", "exr", []byte{0x76, 0x2F, 0x31, 0x01}, true, true},
		{"PSD", "psd", []byte("8BPS"), true, true},
		{"TGA", "tga", nil, true, true},
		{"PPM", "ppm", []byte("P6"), true, true},
		{"PAM", "pam", []byte("P7"), true, true},
		{"SVG", "svg", nil, true, true},
		{"DPX", "dpx", nil, true, true},
		{"FITS", "fits", nil, true, true},
		{"DJVU", "djvu", nil, true, false},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			if runtime.GOOS == "linux" && !f.linux {
				t.Skipf("%s not supported on linux", f.name)
			}
			if runtime.GOOS == "darwin" && !f.darwin {
				t.Skipf("%s not supported on darwin", f.name)
			}

			dir := t.TempDir()
			input := filepath.Join(dir, "input.png")
			output := filepath.Join(dir, "output."+f.ext)
			createTestPNG(t, input)

			err := magick.Convert(input, output, magick.ConvertOptions{})
			if err != nil {
				t.Fatalf("Convert PNG->%s: %v", f.name, err)
			}

			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			if len(data) == 0 {
				t.Fatalf("output file is empty")
			}
			if f.magic != nil && len(data) >= len(f.magic) {
				if !bytes.HasPrefix(data, f.magic) {
					t.Errorf("output magic mismatch: got %x, want %x", data[:len(f.magic)], f.magic)
				}
			}

			info, err := magick.Identify(output)
			if err != nil {
				t.Fatalf("Identify output: %v", err)
			}
			if info.Width == 0 || info.Height == 0 {
				t.Errorf("output has zero dimensions: %dx%d", info.Width, info.Height)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	setup(t)

	formats := []string{"jpg", "webp", "tiff", "png", "bmp", "gif"}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "input.png")
			mid := filepath.Join(dir, "mid."+ext)
			output := filepath.Join(dir, "output.png")
			createTestPNG(t, input)

			if err := magick.Convert(input, mid, magick.ConvertOptions{}); err != nil {
				t.Fatalf("Convert PNG->%s: %v", ext, err)
			}
			if err := magick.Convert(mid, output, magick.ConvertOptions{}); err != nil {
				t.Fatalf("Convert %s->PNG: %v", ext, err)
			}

			info, err := magick.Identify(output)
			if err != nil {
				t.Fatalf("Identify: %v", err)
			}
			if info.Width != 4 || info.Height != 4 {
				t.Errorf("round-trip lost dimensions: got %dx%d", info.Width, info.Height)
			}
		})
	}
}

func TestFormats(t *testing.T) {
	setup(t)
	formats := magick.Formats()
	if len(formats) == 0 {
		t.Fatal("Formats returned empty list")
	}
	required := []string{"JPEG", "PNG", "WEBP", "TIFF", "HEIC", "JXL", "GIF", "BMP", "SVG", "PDF", "AVIF"}
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
