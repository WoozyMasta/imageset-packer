package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/woozymasta/imageset-packer/internal/imageio"
)

// CmdConvert converts a single image between supported formats.
type CmdConvert struct {
	Args struct {
		Input  string `positional-arg-name:"input" description:"Input file: png,tga,tiff,bmp,dds,edds" required:"yes"`
		Output string `positional-arg-name:"output" description:"Output file: png,tga,tiff,bmp,dds,edds" required:"yes"`
	} `positional-args:"yes" required:"yes"`

	AlphaKey    string `long:"alpha-key" description:"Color key as RRGGBB -> alpha=0 (optional)" default:""`
	AlphaKeyOff bool   `long:"alpha-key-off" description:"Disable color key processing"`
}

// Execute runs the convert command.
func (c *CmdConvert) Execute(args []string) error {
	img, err := imageio.Read(c.Args.Input)
	if err != nil {
		return err
	}

	if !c.AlphaKeyOff && c.AlphaKey != "" {
		rgb, err := imageio.ParseHexRGB(c.AlphaKey)
		if err != nil {
			return fmt.Errorf("invalid --alpha-key: %w", err)
		}
		img = imageio.ApplyColorKey(img, rgb)
	}

	// Optional sanity: output ext known
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(c.Args.Output), "."))
	if ext == "" {
		return fmt.Errorf("output has no extension: %q", c.Args.Output)
	}

	return imageio.Write(c.Args.Output, img)
}
