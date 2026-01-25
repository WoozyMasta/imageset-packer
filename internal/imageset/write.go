// Package imageset provides structures and functions for working with imageset files.
package imageset

import (
	"fmt"
	"io"
	"strings"
)

// Write writes an ImageSetClass to the writer in imageset text format.
func Write(w io.Writer, is *ImageSetClass, useCamelCase bool) error {
	return writeImageSetClass(w, is, 0, useCamelCase)
}

// writeImageSetClass writes ImageSetClass with indentation.
func writeImageSetClass(w io.Writer, is *ImageSetClass, indent int, useCamelCase bool) error {
	indentStr := strings.Repeat("\t", indent)

	// ImageSetClass {
	if _, err := fmt.Fprintf(w, "%sImageSetClass {\n", indentStr); err != nil {
		return err
	}

	// Name
	className := NormalizeName(is.Name, useCamelCase)
	if _, err := fmt.Fprintf(w, "%s\tName %q\n", indentStr, className); err != nil {
		return err
	}

	// RefSize
	if _, err := fmt.Fprintf(w, "%s\tRefSize %d %d\n", indentStr, is.RefSize[0], is.RefSize[1]); err != nil {
		return err
	}

	// Textures
	if len(is.Textures) > 0 {
		if _, err := fmt.Fprintf(w, "%s\tTextures {\n", indentStr); err != nil {
			return err
		}
		for _, tex := range is.Textures {
			if err := writeTexture(w, &tex, indent+2); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s\t}\n", indentStr); err != nil {
			return err
		}
	}

	// Images
	if len(is.Images) > 0 {
		if _, err := fmt.Fprintf(w, "%s\tImages {\n", indentStr); err != nil {
			return err
		}
		for _, img := range is.Images {
			if err := writeImageDef(w, &img, indent+2, useCamelCase); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s\t}\n", indentStr); err != nil {
			return err
		}
	}

	// Groups
	if len(is.Groups) > 0 {
		if _, err := fmt.Fprintf(w, "%s\tGroups {\n", indentStr); err != nil {
			return err
		}
		for _, group := range is.Groups {
			if err := writeGroup(w, &group, indent+2, useCamelCase); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s\t}\n", indentStr); err != nil {
			return err
		}
	}

	// }
	if _, err := fmt.Fprintf(w, "%s}\n", indentStr); err != nil {
		return err
	}

	return nil
}

// writeTexture writes ImageSetTextureClass.
func writeTexture(w io.Writer, tex *ImageSetTextureClass, indent int) error {
	indentStr := strings.Repeat("\t", indent)
	if _, err := fmt.Fprintf(w, "%sImageSetTextureClass {\n", indentStr); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\tmpix %d\n", indentStr, tex.Mpix); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\tpath %q\n", indentStr, tex.Path); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s}\n", indentStr); err != nil {
		return err
	}

	return nil
}

// writeImageDef writes ImageSetDefClass.
func writeImageDef(w io.Writer, img *ImageSetDefClass, indent int, useCamelCase bool) error {
	indentStr := strings.Repeat("\t", indent)

	// Use the name as a class identifier when available.
	className := NormalizeName(img.Name, useCamelCase)
	if className == "" {
		className = "default"
	}
	if _, err := fmt.Fprintf(w, "%sImageSetDefClass %s {\n", indentStr, className); err != nil {
		return err
	}

	name := NormalizeName(img.Name, useCamelCase)
	if _, err := fmt.Fprintf(w, "%s\tName %q\n", indentStr, name); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s\tPos %d %d\n", indentStr, img.Pos[0], img.Pos[1]); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\tSize %d %d\n", indentStr, img.Size[0], img.Size[1]); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\tFlags %d\n", indentStr, img.Flags); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s}\n", indentStr); err != nil {
		return err
	}

	return nil
}

// writeGroup writes ImageSetGroupClass.
func writeGroup(w io.Writer, group *ImageSetGroupClass, indent int, useCamelCase bool) error {
	indentStr := strings.Repeat("\t", indent)

	className := NormalizeName(group.Name, useCamelCase)
	if className == "" {
		className = "default"
	}
	if _, err := fmt.Fprintf(w, "%sImageSetGroupClass %s {\n", indentStr, className); err != nil {
		return err
	}

	name := NormalizeName(group.Name, useCamelCase)
	if _, err := fmt.Fprintf(w, "%s\tName %q\n", indentStr, name); err != nil {
		return err
	}

	if len(group.Images) > 0 {
		if _, err := fmt.Fprintf(w, "%s\tImages {\n", indentStr); err != nil {
			return err
		}
		for _, img := range group.Images {
			if err := writeImageDef(w, &img, indent+2, useCamelCase); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s\t}\n", indentStr); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s}\n", indentStr); err != nil {
		return err
	}

	return nil
}
