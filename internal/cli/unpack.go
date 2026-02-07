package cli

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/edds"
	"github.com/woozymasta/imageset-packer/internal/imageio"
	"github.com/woozymasta/imageset-packer/internal/imageset"
)

// CmdUnpack extracts images from an imageset/edds pair.
type CmdUnpack struct {
	Args struct {
		ImageSetPath string `positional-arg-name:"imageset" description:"Path to .imageset" required:"yes"`
		EDDSPath     string `positional-arg-name:"edds" description:"Path to .edds" required:"yes"`
	} `positional-args:"yes" required:"yes"`

	OutFormat  string `short:"o" long:"out-format" description:"Output format: png,tga,tiff,bmp,dds (default: png)" default:"png"`
	OutputDir  string `short:"O" long:"output-dir" description:"Output directory (default: current dir)"`
	Overwrite  bool   `short:"f" long:"force" description:"Overwrite existing files"`
	KeepGroups bool   `short:"g" long:"groups" description:"Write groups into subdirectories"`
	Dedup      bool   `short:"d" long:"deduplicate" description:"Drop duplicate entries with identical Pos/Size"`
}

// Execute runs the unpack command.
func (c *CmdUnpack) Execute(args []string) error {
	return runUnpack(c)
}

func runUnpack(opts *CmdUnpack) error {
	is, err := imageset.ReadFile(opts.Args.ImageSetPath)
	if err != nil {
		return fmt.Errorf("read imageset: %w", err)
	}

	atlas, err := edds.Read(opts.Args.EDDSPath)
	if err != nil {
		return fmt.Errorf("read edds: %w", err)
	}

	// autoscale by RefSize (imageset) vs real atlas size (edds)
	refW := is.RefSize[0]
	refH := is.RefSize[1]

	b := atlas.Bounds()
	atlasW := b.Dx()
	atlasH := b.Dy()

	sx, sy := 1, 1
	if refW > 0 && refH > 0 {
		if atlasW%refW == 0 {
			sx = atlasW / refW
		}
		if atlasH%refH == 0 {
			sy = atlasH / refH
		}
	}
	if sx < 1 {
		sx = 1
	}
	if sy < 1 {
		sy = 1
	}

	outDir := opts.OutputDir
	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0750); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	format := strings.ToLower(strings.TrimPrefix(opts.OutFormat, "."))
	if format == "" {
		format = "png"
	}

	// root images
	rootImages := is.Images
	if opts.Dedup {
		rootImages = deduplicateDefs(rootImages)
	}
	if len(rootImages) > 0 {
		for _, def := range rootImages {
			if err := writeOne(atlas, def, sx, sy, outDir, "", format, opts.Overwrite); err != nil {
				return err
			}
		}
	}

	// groups
	for _, g := range is.Groups {
		groupImages := g.Images
		if opts.Dedup {
			groupImages = deduplicateDefs(groupImages)
		}
		groupDir := ""
		if opts.KeepGroups {
			groupDir = sanitizeName(g.Name)
		}
		for _, def := range groupImages {
			if err := writeOne(atlas, def, sx, sy, outDir, groupDir, format, opts.Overwrite); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeOne writes a single image to the output directory.
func writeOne(atlas image.Image, def imageset.ImageSetDefClass, sx, sy int, baseDir, groupDir, format string, overwrite bool) error {
	sub, err := crop(atlas, def.Pos[0]*sx, def.Pos[1]*sy, def.Size[0]*sx, def.Size[1]*sy)
	if err != nil {
		return fmt.Errorf("crop %q: %w", def.Name, err)
	}

	dir := baseDir
	if groupDir != "" {
		dir = filepath.Join(baseDir, groupDir)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("mkdir group dir: %w", err)
		}
	}

	outPath := filepath.Join(dir, def.Name+"."+format)
	if !overwrite {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("output file %q exists (use --force)", outPath)
		}
	}

	// DDS output expects full image
	if err := imageio.Write(outPath, sub); err != nil {
		return fmt.Errorf("write %q: %w", outPath, err)
	}

	return nil
}

// crop crops the image to the given rectangle.
func crop(src image.Image, x, y, w, h int) (*image.RGBA, error) {
	b := src.Bounds()

	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid crop size: %dx%d", w, h)
	}

	// Bounds include b.Min (not always 0,0).
	x0 := b.Min.X + x
	y0 := b.Min.Y + y
	x1 := x0 + w
	y1 := y0 + h

	if x0 < b.Min.X || y0 < b.Min.Y || x1 > b.Max.X || y1 > b.Max.Y {
		return nil, fmt.Errorf("crop out of bounds: rect=[%d,%d..%d,%d] src=[%d,%d..%d,%d]",
			x0, y0, x1, y1, b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), src, image.Point{X: x0, Y: y0}, draw.Src)

	return dst, nil
}

// sanitizeName sanitizes the name of the group.
func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "..", ".")
	if s == "" {
		return "group"
	}

	return s
}

// deduplicateDefs deduplicates the image definitions.
func deduplicateDefs(defs []imageset.ImageSetDefClass) []imageset.ImageSetDefClass {
	if len(defs) <= 1 {
		return defs
	}

	seen := make(map[[4]int]struct{}, len(defs))
	out := make([]imageset.ImageSetDefClass, 0, len(defs))
	for _, def := range defs {
		key := [4]int{def.Pos[0], def.Pos[1], def.Size[0], def.Size[1]}
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, def)
	}

	return out
}
