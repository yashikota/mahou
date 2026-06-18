package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yashikota/magick-go/internal/magick"
)

func runIdentify(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("identify", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("identify requires input path")
	}
	if _, err := initialize(opts); err != nil {
		return err
	}
	info, err := magick.Identify(fs.Arg(0))
	if err != nil {
		return err
	}
	if opts.json {
		return printJSON(info)
	}
	fmt.Fprintf(os.Stdout, "%s %s %dx%d depth=%d\n", info.Path, info.Format, info.Width, info.Height, info.Depth)
	return nil
}
