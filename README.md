# magick-go

`magickgo` is a standalone, pure-Go CLI that bundles a complete ImageMagick 7 runtime.
No system dependencies required — the binary extracts a self-contained
`runtime-<target>.tar.zst` into the user cache, configures the ImageMagick
environment, and loads `libMagickWand` dynamically through `purego` (no CGO).

## Supported Targets

| OS    | Architecture | Target          |
|-------|-------------|-----------------|
| Linux | amd64       | `linux-amd64`   |
| Linux | arm64       | `linux-arm64`   |
| macOS | arm64       | `darwin-arm64`  |

## Supported Image Formats

ImageMagick 7.1.2-8 Q16-HDRI with the following delegates enabled:

### Core Formats (all platforms)

| Format | Extension(s)         | Read | Write | Notes                      |
|--------|---------------------|------|-------|----------------------------|
| JPEG   | `.jpg`, `.jpeg`     | ✓    | ✓     | libjpeg-turbo              |
| PNG    | `.png`              | ✓    | ✓     | libpng                     |
| WebP   | `.webp`             | ✓    | ✓     | libwebp                    |
| TIFF   | `.tiff`, `.tif`     | ✓    | ✓     | libtiff                    |
| GIF    | `.gif`              | ✓    | ✓     | Built-in                   |
| BMP    | `.bmp`              | ✓    | ✓     | Built-in                   |
| HEIC   | `.heic`, `.heif`    | ✓    | ✓     | libheif (HEVC/AV1)         |
| AVIF   | `.avif`             | ✓    | ✓     | libheif + AOM              |
| JXL    | `.jxl`              | ✓    | ✓     | libjxl (JPEG XL)           |
| SVG    | `.svg`, `.svgz`     | ✓    | ✓     | librsvg                    |
| PDF    | `.pdf`              | ✓    | ✓     | Ghostscript delegate       |
| EXR    | `.exr`              | ✓    | ✓     | OpenEXR                    |
| PSD    | `.psd`, `.psb`      | ✓    | ✓     | Photoshop format           |
| JP2    | `.jp2`, `.j2k`      | ✓    | ✓     | OpenJPEG                   |

### Additional Formats

| Format   | Extension(s)     | Read | Write | Notes                    |
|----------|-----------------|------|-------|--------------------------|
| DPX      | `.dpx`          | ✓    | ✓     | Digital cinema            |
| TGA      | `.tga`          | ✓    | ✓     | Targa                    |
| PPM/PGM  | `.ppm`, `.pgm`  | ✓    | ✓     | Netpbm                   |
| PAM      | `.pam`          | ✓    | ✓     | Portable Arbitrary Map   |
| FITS     | `.fits`         | ✓    | ✓     | Astronomy                |
| ICO      | `.ico`          | ✓    | ✓     | Windows icon             |
| PCX      | `.pcx`          | ✓    | ✓     | PC Paintbrush            |
| XPM      | `.xpm`          | ✓    | ✓     | X11 pixmap               |
| FARBFELD | `.ff`           | ✓    | ✓     | suckless format          |
| QOI      | `.qoi`          | ✓    | ✓     | Quite OK Image           |
| HDR      | `.hdr`          | ✓    | ✓     | Radiance RGBE            |
| DJVU     | `.djvu`         | ✓    | —     | Linux only               |

### Delegates

| Platform | Delegates                                                                                              |
|----------|-------------------------------------------------------------------------------------------------------|
| Linux    | bzlib cairo djvu fftw fontconfig freetype heic jbig jng jp2 jpeg jxl lcms lqr lzma openexr pango png ps raqm raw rsvg tiff webp wmf xml zip zlib zstd |
| macOS    | bzlib cairo fontconfig freetype heic jng jp2 jpeg jxl lcms lzma openexr pango png ps rsvg tiff webp xml zlib zstd |

### Security Policy

PDF, PostScript (PS/EPS), MVG, and MSL formats are disabled by the default
security policy. URL/HTTP/HTTPS delegates are also blocked. Use
`--policy permissive` to enable all formats for trusted workflows.

## Commands

```sh
magickgo doctor --verbose     # Show runtime info, delegates, format support
magickgo formats              # List all supported formats
magickgo identify input.png   # Image metadata (dimensions, format, depth)
magickgo convert input.heic output.webp              # Format conversion
magickgo convert input.png output.jpg --quality 85   # With quality
magickgo convert input.jpg output.jpg --strip        # Strip metadata
magickgo resize input.jpg output.webp --width 1200   # Resize (aspect-ratio preserved)
```

### Options

| Flag             | Description                              |
|------------------|------------------------------------------|
| `--quality N`    | Output quality (1-100, format-dependent) |
| `--strip`        | Remove EXIF/metadata                     |
| `--auto-orient`  | Auto-rotate based on EXIF orientation    |
| `--format FMT`   | Override output format                   |
| `--width N`      | Target width for resize (Lanczos filter) |
| `--policy P`     | `safe` (default) or `permissive`         |
| `--json`         | JSON output (doctor, identify, formats)  |
| `--verbose`      | Verbose output (doctor)                  |

## Architecture

```
┌─────────────────────────────────────┐
│         magickgo binary             │
│  (CGO_ENABLED=0, pure Go)          │
├─────────────────────────────────────┤
│  cmd/magickgo/     CLI commands     │
│  internal/magick/  purego bindings  │
│  internal/runtimebundle/            │
│    ├── embed.go    //go:embed       │
│    ├── extract.go  zstd extraction  │
│    ├── env.go      LD_LIBRARY_PATH  │
│    └── policy.go   security policy  │
├─────────────────────────────────────┤
│  runtime-<target>.tar.zst           │
│  (embedded at build time)           │
│    ├── bin/magick                    │
│    ├── lib/libMagickWand-7.*.so     │
│    ├── lib/libMagickCore-7.*.so     │
│    ├── lib/*.so (all dependencies)  │
│    ├── lib/ImageMagick-*/modules/   │
│    └── etc/ImageMagick-*/           │
└─────────────────────────────────────┘
```

## Building

The repository does not commit runtime bundles. CI builds them from source:

```sh
# Linux (builds ImageMagick from source)
bash scripts/build-runtime-linux.sh linux-amd64 internal/runtimebundle/assets/runtime-linux-amd64.tar.zst

# macOS (builds ImageMagick from source with Homebrew dependencies)
bash scripts/build-runtime-darwin.sh darwin-arm64 internal/runtimebundle/assets/runtime-darwin-arm64.tar.zst

# Then build the Go binary
CGO_ENABLED=0 go build -o dist/magickgo ./cmd/magickgo
```

CI caches the runtime tarball keyed on script content hash. First build takes
5-10 minutes; subsequent builds with unchanged scripts complete in under 3 minutes.
