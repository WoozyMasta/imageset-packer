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

	AlphaKey    string `long:"alpha-key" description:"Color key as RRGGBB -> alpha=0" default:""`
	Format      string `short:"F" long:"format" description:"Output format for DDS/EDDS" choice:"bgra8" choice:"dxt1" choice:"dxt5" default:"bgra8"`
	Quality     int    `short:"q" long:"quality" description:"DXT1/DXT5 quality level 1..10, 0=optimal" default:"0"`
	Mipmaps     int    `short:"x" long:"mipmaps" description:"Mipmap levels for DDS/EDDS output, 0=full chain" default:"0"`
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

	if c.Mipmaps < 0 {
		return fmt.Errorf("mipmaps must be >= 0")
	}
	if err := imageio.ValidateQualityLevel(c.Quality); err != nil {
		return fmt.Errorf("invalid --quality: %w", err)
	}

	outputFormat, err := imageio.ParseOutputFormat(c.Format)
	if err != nil {
		return fmt.Errorf("invalid --format: %w", err)
	}

	if ext != "dds" && ext != "edds" {
		if strings.TrimSpace(c.Format) != "" || c.Quality != 0 || c.Mipmaps != 0 {
			return fmt.Errorf("--format/--quality/--mipmaps are supported only for dds/edds output")
		}
		return imageio.Write(c.Args.Output, img)
	}
	if ext == "dds" && c.Mipmaps != 0 {
		return fmt.Errorf("--mipmaps is supported only for edds output")
	}

	return imageio.WriteWithOptions(c.Args.Output, img, &imageio.EncodeSettings{
		Format:  outputFormat,
		Quality: c.Quality,
		Mipmaps: c.Mipmaps,
	})
}
