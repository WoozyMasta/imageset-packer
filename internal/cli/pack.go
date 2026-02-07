package cli

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/woozymasta/imageset-packer/internal/imageio"
	"github.com/woozymasta/imageset-packer/internal/imageset"
	"github.com/woozymasta/imageset-packer/internal/packer"
	"golang.org/x/image/draw"
)

// PackPackingFlags defines atlas packing parameters.
type PackPackingFlags struct {
	Rule          string  `short:"r" long:"rule" description:"Packing rule" default:"bl" choice:"bssf" choice:"blsf" choice:"baf" choice:"bl" choice:"cp" yaml:"rule"`
	OutputFormat  string  `short:"F" long:"out-format" description:"Output format for DDS/EDDS" choice:"bgra8" choice:"dxt1" choice:"dxt5" default:"bgra8" yaml:"out_format"`
	MinSize       int     `short:"m" long:"min-size" description:"Minimum texture size (power of 2)" default:"256" yaml:"min_size"`
	MaxSize       int     `short:"M" long:"max-size" description:"Maximum texture size (power of 2)" default:"4096" yaml:"max_size"`
	Gap           int     `short:"g" long:"gap" description:"Gap between images" default:"0" yaml:"gap"`
	Quality       int     `short:"q" long:"quality" description:"DXT1/DXT5 quality level 1..10, 0=optimal" default:"0" yaml:"quality"`
	Mipmaps       int     `short:"x" long:"mipmaps" description:"Mipmap levels for DDS/EDDS output, 0=full chain" default:"0" yaml:"mipmaps"`
	AspectPenalty float64 `short:"a" long:"aspect-penalty" description:"Aspect penalty for non-square textures" default:"0.25" yaml:"aspect_penalty"`
	PreferHeight  bool    `short:"p" long:"prefer-height" description:"Prefer height over width for aspect ratio" yaml:"prefer_height"`
	ForceSquare   bool    `short:"S" long:"force-square" description:"Force square texture" yaml:"force_square"`
	AllowRotate   bool    `short:"R" long:"rotate" description:"Allow 90-degree rotation for better packing" yaml:"rotate"`
}

// PackInputFlags defines input discovery and preprocessing options.
type PackInputFlags struct {
	GroupSeparator string   `short:"s" long:"group-separator" description:"Separator for group name in filename (e.g. '_' for 'Group_Image.png')" yaml:"group_separator"`
	AlphaKey       string   `long:"alpha-key" description:"Color key as RRGGBB (e.g. ff00ff) -> alpha=0 for bmp/tga/tiff by default" default:"ff00ff" yaml:"alpha_key"`
	InFormats      []string `short:"i" long:"in-format" description:"Allowed input formats: png,tga,tiff,bmp (repeatable). Default: png,tga,tiff,bmp" yaml:"in_format"`
	MaxInputSide   int      `short:"D" long:"max-input-side" description:"Downscale inputs so the longest side is at most N pixels (0=off)" default:"0" yaml:"max_input_side"`
	GroupDirs      bool     `short:"d" long:"group-dirs" description:"Treat subdirectories as groups" yaml:"group_dirs"`
	AlphaKeyOff    bool     `long:"alpha-key-off" description:"Disable color key transparency processing" yaml:"alpha_key_off"`
	AlphaKeyAll    bool     `long:"alpha-key-all" description:"Apply color key to all formats, including png" yaml:"alpha_key_all"`
}

// CmdPack packs images into a texture atlas and imageset definition.
type CmdPack struct {
	// betteralign:ignore

	Name  string `short:"n" long:"name" description:"ImageSet name (default: input directory name)" yaml:"name"`
	Force bool   `short:"f" long:"force" description:"Overwrite existing output files" yaml:"force"`
	Camel bool   `short:"c" long:"camel-case" description:"Use CamelCase names in imageset output (default: snake_case)" yaml:"camel_case"`
	Path  string `short:"P" long:"edds-path" description:"Prefix path for imageset texture reference (e.g. mod/data/images)" yaml:"edds_path"`
	Skip  bool   `short:"u" long:"skip-unchanged" description:"Skip writing when inputs are unchanged" yaml:"skip_unchanged"`

	Packing PackPackingFlags `group:"Packing" yaml:"packing"`
	Input   PackInputFlags   `group:"Input" yaml:"input"`

	Args struct {
		Input  string `positional-arg-name:"input" description:"Input directory with images" required:"yes" yaml:"input_dir"`
		Output string `positional-arg-name:"output" description:"Output directory (default: input directory)" yaml:"output_dir"`
	} `positional-args:"yes" required:"yes" yaml:"args"`
}

// imageFile represents a single image file.
type imageFile struct {
	image     image.Image
	path      string
	name      string
	groupName string
	width     int
	height    int
}

// Execute runs the pack command.
func (c *CmdPack) Execute(args []string) error {
	return runPack(c)
}

// runPack runs the pack command.
func runPack(opts *CmdPack) error {
	outputDir := opts.Args.Output
	if outputDir == "" {
		outputDir = opts.Args.Input
	}

	if opts.Packing.Mipmaps < 0 {
		return fmt.Errorf("mipmaps must be >= 0")
	}
	if err := imageio.ValidateQualityLevel(opts.Packing.Quality); err != nil {
		return fmt.Errorf("invalid --quality: %w", err)
	}
	outputFormat, err := imageio.ParseOutputFormat(opts.Packing.OutputFormat)
	if err != nil {
		return fmt.Errorf("invalid --output-format: %w", err)
	}

	name := opts.Name
	if name == "" {
		absInput, err := filepath.Abs(opts.Args.Input)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		name = filepath.Base(absInput)
	}

	imagesetPath := filepath.Join(outputDir, name+".imageset")
	eddsPath := filepath.Join(outputDir, name+".edds")

	allowed := normalizeFormats(opts.Input.InFormats)
	if len(allowed) == 0 {
		allowed = map[string]bool{"png": true, "tga": true, "tiff": true, "bmp": true}
	}

	alphaKeyRGB, err := imageio.ParseHexRGB(opts.Input.AlphaKey)
	if err != nil {
		return fmt.Errorf("invalid --alpha-key: %w", err)
	}

	var imageFiles []imageFile

	// Read input dir
	if opts.Input.GroupDirs {
		groups, err := readImageFilesFromDirs(opts.Args.Input, allowed)
		if err != nil {
			return fmt.Errorf("failed to read directories: %w", err)
		}

		// stable iteration
		groupNames := make([]string, 0, len(groups))
		for g := range groups {
			groupNames = append(groupNames, g)
		}
		sort.Strings(groupNames)

		for _, groupName := range groupNames {
			for _, file := range groups[groupName] {
				img, err := imageio.Read(file)
				if err != nil {
					return fmt.Errorf("failed to read image %q: %w", file, err)
				}

				img = applyColorKeyIfNeeded(img, file, opts, alphaKeyRGB)
				img, w, h := downscaleIfNeeded(img, opts.Input.MaxInputSide)

				baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
				imageFiles = append(imageFiles, imageFile{
					path:      file,
					name:      baseName,
					groupName: groupName,
					width:     w,
					height:    h,
					image:     img,
				})
			}
		}

		// root (no group)
		rootFiles, err := readImageFiles(opts.Args.Input, allowed)
		if err != nil {
			return fmt.Errorf("failed to read root directory: %w", err)
		}

		for _, file := range rootFiles {
			img, err := imageio.Read(file)
			if err != nil {
				return fmt.Errorf("failed to read image %q: %w", file, err)
			}

			img = applyColorKeyIfNeeded(img, file, opts, alphaKeyRGB)
			img, w, h := downscaleIfNeeded(img, opts.Input.MaxInputSide)

			baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			imageFiles = append(imageFiles, imageFile{
				path:      file,
				name:      baseName,
				groupName: "",
				width:     w,
				height:    h,
				image:     img,
			})
		}
	} else if opts.Input.GroupSeparator != "" {
		files, err := readImageFiles(opts.Args.Input, allowed)
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		for _, file := range files {
			img, err := imageio.Read(file)
			if err != nil {
				return fmt.Errorf("failed to read image %q: %w", file, err)
			}

			img = applyColorKeyIfNeeded(img, file, opts, alphaKeyRGB)
			img, w, h := downscaleIfNeeded(img, opts.Input.MaxInputSide)

			baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			groupName, imageName := splitGroupName(baseName, opts.Input.GroupSeparator)

			imageFiles = append(imageFiles, imageFile{
				path:      file,
				name:      imageName,
				groupName: groupName,
				width:     w,
				height:    h,
				image:     img,
			})
		}
	} else {
		files, err := readImageFiles(opts.Args.Input, allowed)
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		for _, file := range files {
			img, err := imageio.Read(file)
			if err != nil {
				return fmt.Errorf("failed to read image %q: %w", file, err)
			}

			img = applyColorKeyIfNeeded(img, file, opts, alphaKeyRGB)
			img, w, h := downscaleIfNeeded(img, opts.Input.MaxInputSide)

			baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			imageFiles = append(imageFiles, imageFile{
				path:      file,
				name:      baseName,
				groupName: "",
				width:     w,
				height:    h,
				image:     img,
			})
		}
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("no input images found in %q", opts.Args.Input)
	}

	// detect name collisions (global)
	seen := make(map[string]string, len(imageFiles))
	for _, f := range imageFiles {
		key := f.name
		if prev, ok := seen[key]; ok {
			return fmt.Errorf("duplicate image name %q (paths: %q and %q). rename or enable grouping separator/dirs", key, prev, f.path)
		}
		seen[key] = f.path
	}

	cachePath := filepath.Join(outputDir, name+".imagehash")
	var inputsHash uint64
	if opts.Skip {
		var err error
		inputsHash, err = computeInputsHash(opts, imageFiles)
		if err != nil {
			return err
		}
		if shouldSkipPack(cachePath, imagesetPath, eddsPath, inputsHash) {
			fmt.Printf("Inputs unchanged; skipping write for %s\n", imagesetPath)
			return nil
		}
	}

	if !opts.Force {
		if _, err := os.Stat(imagesetPath); err == nil {
			return fmt.Errorf("output file %q already exists (use --force)", imagesetPath)
		}
		if _, err := os.Stat(eddsPath); err == nil {
			return fmt.Errorf("output file %q already exists (use --force)", eddsPath)
		}
	}

	imageInfos := make([]packer.ImageInfo, 0, len(imageFiles))
	for _, imgFile := range imageFiles {
		imageInfos = append(imageInfos, packer.ImageInfo{
			Name:   imgFile.name,
			Width:  imgFile.width,
			Height: imgFile.height,
			Image:  imgFile.image,
		})
	}

	cfg := packer.Config{
		MinSize:       opts.Packing.MinSize,
		MaxSize:       opts.Packing.MaxSize,
		Gap:           opts.Packing.Gap,
		PreferHeight:  opts.Packing.PreferHeight,
		ForceSquare:   opts.Packing.ForceSquare,
		AllowRotate:   opts.Packing.AllowRotate,
		AspectPenalty: opts.Packing.AspectPenalty,
		Rule:          parseRule(opts.Packing.Rule),
	}

	result, err := packer.Pack(imageInfos, cfg)
	if err != nil {
		return fmt.Errorf("failed to pack images: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	placementMap := make(map[string]packer.Placement, len(result.Placements))
	for _, placement := range result.Placements {
		placementMap[placement.Name] = placement
	}

	imagesetData := &imageset.ImageSetClass{
		Name:    name,
		RefSize: [2]int{result.Width, result.Height},
		Textures: []imageset.ImageSetTextureClass{
			{
				Mpix: 1,
				Path: formatEddsRefPath(opts.Path, name),
			},
		},
	}

	groupsMap := make(map[string][]imageset.ImageSetDefClass)
	var rootImages []imageset.ImageSetDefClass

	for _, imgFile := range imageFiles {
		placement, ok := placementMap[imgFile.name]
		if !ok {
			return fmt.Errorf("placement not found for image %q", imgFile.name)
		}

		imgDef := imageset.ImageSetDefClass{
			Name: imgFile.name,
			Pos:  [2]int{placement.X, placement.Y},
			Size: [2]int{placement.Width, placement.Height},
		}

		if imgFile.groupName != "" {
			groupsMap[imgFile.groupName] = append(groupsMap[imgFile.groupName], imgDef)
		} else {
			rootImages = append(rootImages, imgDef)
		}
	}

	if len(groupsMap) > 0 {
		imagesetData.Groups = make([]imageset.ImageSetGroupClass, 0, len(groupsMap))
		groupNames := make([]string, 0, len(groupsMap))
		for groupName := range groupsMap {
			groupNames = append(groupNames, groupName)
		}
		sort.Strings(groupNames)

		for _, groupName := range groupNames {
			imagesetData.Groups = append(imagesetData.Groups, imageset.ImageSetGroupClass{
				Name:   groupName,
				Images: groupsMap[groupName],
			})
		}

		if len(rootImages) > 0 {
			imagesetData.Images = rootImages
		}
	} else {
		imagesetData.Images = rootImages
	}

	imagesetFile, err := os.Create(imagesetPath)
	if err != nil {
		return fmt.Errorf("failed to create imageset file: %w", err)
	}
	defer func() { _ = imagesetFile.Close() }()

	if err := imageset.Write(imagesetFile, imagesetData, opts.Camel); err != nil {
		return fmt.Errorf("failed to write imageset file: %w", err)
	}

	if err := imageio.WriteWithOptions(eddsPath, result.Image, &imageio.EncodeSettings{
		Format:  outputFormat,
		Quality: opts.Packing.Quality,
		Mipmaps: opts.Packing.Mipmaps,
	}); err != nil {
		return fmt.Errorf("failed to write EDDS file: %w", err)
	}

	if opts.Skip && inputsHash != 0 {
		if err := writeCacheHash(cachePath, inputsHash); err != nil {
			return err
		}
	}

	if name != "" {
		fmt.Printf("Packed %d images from %s as %s into %dx%d\n", len(imageInfos), opts.Args.Input, name, result.Width, result.Height)
	} else {
		fmt.Printf("Packed %d images from %s into %dx%d\n", len(imageInfos), opts.Args.Input, result.Width, result.Height)
	}
	fmt.Printf("Outputs: %s, %s\n", imagesetPath, eddsPath)

	return nil
}

// applyColorKeyIfNeeded applies the color key if needed.
func applyColorKeyIfNeeded(img image.Image, path string, opts *CmdPack, key imageio.RGB) image.Image {
	if opts.Input.AlphaKeyOff {
		return img
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if opts.Input.AlphaKeyAll || ext == "bmp" || ext == "tga" || ext == "tiff" {
		return imageio.ApplyColorKey(img, key)
	}

	return img
}

// downscaleIfNeeded downscales the image if needed.
func downscaleIfNeeded(img image.Image, maxSide int) (image.Image, int, int) {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	if maxSide <= 0 {
		return img, width, height
	}
	if width <= maxSide && height <= maxSide {
		return img, width, height
	}

	longSide := width
	if height > width {
		longSide = height
	}
	scale := float64(maxSide) / float64(longSide)

	newWidth := int(math.Round(float64(width) * scale))
	if newWidth < 1 {
		newWidth = 1
	}

	newHeight := int(math.Round(float64(height) * scale))
	if newHeight < 1 {
		newHeight = 1
	}

	scaled := img
	curW := width
	curH := height
	for curW > newWidth*2 || curH > newHeight*2 {
		stepW := max(newWidth, curW/2)
		stepH := max(newHeight, curH/2)
		scaled = scaleImage(scaled, stepW, stepH)
		curW = stepW
		curH = stepH
	}

	if curW != newWidth || curH != newHeight {
		scaled = scaleImage(scaled, newWidth, newHeight)
	}

	return scaled, newWidth, newHeight
}

// scaleImage scales the image using the CatmullRom algorithm.
func scaleImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

// normalizeFormats normalizes the input formats.
func normalizeFormats(in []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.TrimPrefix(s, ".")
		if s == "" {
			continue
		}
		m[s] = true
	}

	return m
}

// readImageFiles reads the image files from the directory.
func readImageFiles(dir string, allowed map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(e.Name()), "."))
		if allowed[ext] {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(out)
	return out, nil
}

// readImageFilesFromDirs reads the image files from the directories.
func readImageFilesFromDirs(rootDir string, allowed map[string]bool) (map[string][]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]string)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		groupDir := filepath.Join(rootDir, e.Name())
		files, err := readImageFiles(groupDir, allowed)
		if err != nil {
			return nil, fmt.Errorf("failed to read group directory %q: %w", groupDir, err)
		}

		if len(files) > 0 {
			groups[e.Name()] = files
		}
	}

	return groups, nil
}

// splitGroupName splits the group name from the filename.
func splitGroupName(filename, separator string) (groupName, imageName string) {
	idx := strings.Index(filename, separator)
	if idx == -1 {
		return "", filename
	}

	return filename[:idx], filename[idx+len(separator):]
}

// parseRule parses the packing rule.
func parseRule(s string) packer.Rule {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bssf":
		return packer.BestShortSideFit
	case "blsf":
		return packer.BestLongSideFit
	case "baf":
		return packer.BestAreaFit
	case "bl":
		return packer.BottomLeft
	case "cp":
		return packer.ContactPoint
	default:
		return packer.BestShortSideFit
	}
}

// formatEddsRefPath formats the EDDS reference path.
func formatEddsRefPath(prefix, name string) string {
	p := strings.TrimSpace(prefix)
	if p == "" {
		return fmt.Sprintf("%s.edds", name)
	}

	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.Trim(p, "/")
	if p == "" {
		return fmt.Sprintf("%s.edds", name)
	}

	return fmt.Sprintf("%s/%s.edds", p, name)
}
