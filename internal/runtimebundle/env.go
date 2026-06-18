package runtimebundle

import (
	"os"
	"path/filepath"
	"runtime"
)

func ConfigureEnvironment(root string) {
	coderPath, filterPath := modulePaths(root)
	setenv("MAGICK_HOME", root)
	setenv("MAGICK_CONFIGURE_PATH", filepath.Join(root, "etc", "ImageMagick-7"))
	setenv("MAGICK_CODER_MODULE_PATH", coderPath)
	setenv("MAGICK_FILTER_MODULE_PATH", filterPath)
	prependPath("PATH", filepath.Join(root, "bin"))
	if runtime.GOOS == "darwin" {
		prependPath("DYLD_LIBRARY_PATH", filepath.Join(root, "lib"))
	} else {
		prependPath("LD_LIBRARY_PATH", filepath.Join(root, "lib"))
	}
}

func setenv(k, v string) {
	_ = os.Setenv(k, v)
}

func prependPath(k, v string) {
	if cur := os.Getenv(k); cur != "" {
		_ = os.Setenv(k, v+string(os.PathListSeparator)+cur)
		return
	}
	_ = os.Setenv(k, v)
}

func Environment(root string) map[string]string {
	coderPath, filterPath := modulePaths(root)
	return map[string]string{
		"MAGICK_HOME":               root,
		"MAGICK_CONFIGURE_PATH":     filepath.Join(root, "etc", "ImageMagick-7"),
		"MAGICK_CODER_MODULE_PATH":  coderPath,
		"MAGICK_FILTER_MODULE_PATH": filterPath,
	}
}

func modulePaths(root string) (string, string) {
	coder := firstGlob(filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "coders"))
	filter := firstGlob(filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "filters"))
	if coder == "" {
		coder = filepath.Join(root, "lib", "ImageMagick-*", "modules-Q16HDRI", "coders")
	}
	if filter == "" {
		filter = filepath.Join(root, "lib", "ImageMagick-*", "modules-Q16HDRI", "filters")
	}
	return coder, filter
}

func firstGlob(pattern string) string {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}
