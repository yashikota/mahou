package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yashikota/magick-go/internal/magick"
)

func runFormats(args []string) error {
	var opts commonOptions
	fs := flag.NewFlagSet("formats", flag.ContinueOnError)
	addCommonFlags(fs, &opts)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if _, err := initialize(opts); err != nil {
		return err
	}
	formats := magick.Formats()
	if opts.json {
		return printJSON(formats)
	}
	for _, f := range formats {
		fmt.Fprintln(os.Stdout, f)
	}
	return nil
}
