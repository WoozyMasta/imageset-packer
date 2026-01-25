// Package cli implements the command-line interface for imageset-packer.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/imageset-packer/internal/vars"
)

// Root defines global CLI flags.
type Root struct{}

// CmdVersion prints build metadata.
type CmdVersion struct{}

// Execute runs the version command.
func (c *CmdVersion) Execute(args []string) error {
	vars.Print()
	return nil
}

// Run parses arguments and executes the selected command.
func Run(args []string) error {
	var root Root

	parser := flags.NewParser(&root, flags.Default)
	parser.Name = filepath.Base(os.Args[0])

	prog := parser.Name
	if _, err := parser.AddCommand(
		"build",
		"Build projects from .imageset-packer.yaml",
		fmt.Sprintf(
			`Run multiple pack jobs from a config file.

Examples:
  %s build ./my-imageset-packer-config.yaml
  %s build --project ui --project icons`,
			prog, prog,
		),
		&CmdBuild{},
	); err != nil {
		return err
	}

	if _, err := parser.AddCommand(
		"pack",
		"Pack images into .imageset + .edds atlas",
		fmt.Sprintf(
			`Pack a directory of images into an EDDS atlas and imageset file.

Examples:
  %s pack ./icons -x 3 -g 2
  %s pack ./icons ./out --force --group-dirs
  %s pack ./icons -P mod/data/images -r bssf`,
			prog, prog, prog,
		),
		&CmdPack{},
	); err != nil {
		return err
	}

	if _, err := parser.AddCommand(
		"unpack",
		"Unpack .imageset + .edds into images",
		fmt.Sprintf(
			`Extract images from an imageset + edds pair.

Examples:
  %s unpack ui.imageset ui.edds
  %s unpack ui.imageset ui.edds --groups --out-format tga`,
			prog, prog,
		),
		&CmdUnpack{},
	); err != nil {
		return err
	}

	if _, err := parser.AddCommand(
		"convert",
		"Convert a single image file between formats",
		fmt.Sprintf(
			`Convert one image between supported formats.

Examples:
  %s convert icon.png icon.tga
  %s convert atlas.edds atlas.png`,
			prog, prog,
		),
		&CmdConvert{},
	); err != nil {
		return err
	}

	if _, err := parser.AddCommand(
		"version",
		"Print build metadata",
		fmt.Sprintf(
			`Show build information.

Examples:
  %s version`,
			prog,
		),
		&CmdVersion{},
	); err != nil {
		return err
	}

	_, err := parser.ParseArgs(args)

	if err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			return nil
		}
		return err
	}

	return nil
}
