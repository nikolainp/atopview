package config

import (
	"bytes"
	"flag"
	"path/filepath"
)

// PrintUsage ...
type PrintUsage struct {
	error
	Usage string
}

// PrintVersion ...
type PrintVersion struct {
	error
}

// Config ...
type Config struct {
	programName string

	PathUtilinty   string
	PathLog        string
	PathStorage    string
	ShowReportOnly bool
}

///////////////////////////////////////////////////////////////////////////////

// NewConfig ...
func NewConfig(args []string) (obj Config, err error) {
	var isPrintVersion bool

	obj.programName = args[0]
	fsOut := &bytes.Buffer{}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(fsOut)
	fs.StringVar(&obj.PathUtilinty, "x", "/usr/bin/atop", "path to the atop executable file")
	fs.BoolVar(&isPrintVersion, "v", false, "print version")
	fs.BoolVar(&obj.ShowReportOnly, "r", false, "just show report")

	if err = fs.Parse(args[1:]); err != nil {
		err = PrintUsage{Usage: fsOut.String()}
		return
	}

	if isPrintVersion {
		err = PrintVersion{}
		return
	}

	if fs.NArg() < 1 {
		err = PrintUsage{Usage: fsOut.String()}
		return
	}

	obj.PathLog, _ = filepath.Abs(fs.Arg(0))
	obj.PathStorage = obj.PathLog + ".report.sqlite"

	return
}

///////////////////////////////////////////////////////////////////////////////
