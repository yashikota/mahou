package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yashikota/mahou/runtimebundle"
)

func main() {
	bundle, err := runtimebundle.Ensure()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(filepath.Join(bundle.Root, "lib"))
}
