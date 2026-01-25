package imageset

// ImageSetClass is the root structure of an imageset file.
//
//revive:disable-next-line:exported // Keep DayZ naming for compatibility.
type ImageSetClass struct {
	Name     string                 // Name of the image set.
	Textures []ImageSetTextureClass // Textures in the image set.
	Images   []ImageSetDefClass     // Images in the image set.
	Groups   []ImageSetGroupClass   // Groups in the image set.
	RefSize  [2]int                 // width, height // Reference size of the image set.
}

// ImageSetTextureClass represents a texture reference.
//
//revive:disable-next-line:exported // Keep DayZ naming for compatibility.
type ImageSetTextureClass struct {
	Path string // Path to the texture.
	Mpix int    // Mpix number of pixels per meter.
}

// ImageSetDefClass represents an image definition.
//
//revive:disable-next-line:exported // Keep DayZ naming for compatibility.
type ImageSetDefClass struct {
	Name  string // Name of the image.
	Pos   [2]int // x, y position of the image.
	Size  [2]int // width, height of the image.
	Flags int    // Flags of the image ISHorizontalTile ISVerticalTile or 0.
}

// ImageSetGroupClass represents a group of images.
//
//revive:disable-next-line:exported // Keep DayZ naming for compatibility.
type ImageSetGroupClass struct {
	Name   string             // Name of the group.
	Images []ImageSetDefClass // Images in the group.
}
