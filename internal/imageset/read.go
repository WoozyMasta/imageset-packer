package imageset

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadFile parses an imageset file from disk.
func ReadFile(path string) (*ImageSetClass, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	is := &ImageSetClass{}
	sc := bufio.NewScanner(f)

	var (
		inImages      bool
		inGroups      bool
		inGroupImages bool

		curGroup *ImageSetGroupClass
		curDef   *ImageSetDefClass
		inDef    bool
	)

	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		// section opens
		switch {
		case curGroup != nil && strings.HasPrefix(line, "Images") && strings.HasSuffix(line, "{"):
			inGroupImages = true
			continue
		case strings.HasPrefix(line, "Textures"):
			// ignore for now
			continue
		case strings.HasPrefix(line, "Images"):
			if strings.HasSuffix(line, "{") {
				inImages = true
				continue
			}
		case strings.HasPrefix(line, "Groups"):
			if strings.HasSuffix(line, "{") {
				inGroups = true
				continue
			}
		case strings.HasPrefix(line, "ImageSetGroupClass"):
			// ImageSetGroupClass <id> {
			curGroup = &ImageSetGroupClass{}
			if name := parseClassName(line); name != "" {
				curGroup.Name = name
			}
			inGroupImages = false
			continue
		case strings.HasPrefix(line, "ImageSetDefClass"):
			curDef = &ImageSetDefClass{}
			inDef = true
			continue
		}

		// block close
		if line == "}" {
			if inDef && curDef != nil {
				// finalize def -> to root or group images
				if curGroup != nil && inGroupImages {
					curGroup.Images = append(curGroup.Images, *curDef)
				} else {
					is.Images = append(is.Images, *curDef)
				}
				curDef = nil
				inDef = false
				continue
			}

			// close sections
			if inGroupImages {
				inGroupImages = false
				continue
			}
			if curGroup != nil && inGroups {
				is.Groups = append(is.Groups, *curGroup)
				curGroup = nil
				continue
			}
			if inImages {
				inImages = false
				continue
			}
			if inGroups {
				inGroups = false
				continue
			}
			continue
		}

		// key-value lines
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "Name":
			val := strings.TrimSpace(strings.Trim(line[len("Name"):], " \t"))
			val = strings.Trim(val, "\"")
			if inDef && curDef != nil {
				curDef.Name = val
			} else if curGroup != nil {
				curGroup.Name = val
			} else {
				is.Name = val
			}

		case "RefSize":
			if len(fields) < 3 {
				return nil, fmt.Errorf("line %d: invalid RefSize", lineNo)
			}
			w, err1 := strconv.Atoi(fields[1])
			h, err2 := strconv.Atoi(fields[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("line %d: invalid RefSize values", lineNo)
			}
			is.RefSize = [2]int{w, h}

		case "Pos":
			if curDef == nil || len(fields) < 3 {
				return nil, fmt.Errorf("line %d: invalid Pos", lineNo)
			}
			x, err1 := strconv.Atoi(fields[1])
			y, err2 := strconv.Atoi(fields[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("line %d: invalid Pos values", lineNo)
			}
			curDef.Pos = [2]int{x, y}

		case "Size":
			if curDef == nil || len(fields) < 3 {
				return nil, fmt.Errorf("line %d: invalid Size", lineNo)
			}
			w, err1 := strconv.Atoi(fields[1])
			h, err2 := strconv.Atoi(fields[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("line %d: invalid Size values", lineNo)
			}
			curDef.Size = [2]int{w, h}

		case "Flags":
			if curDef == nil || len(fields) < 2 {
				return nil, fmt.Errorf("line %d: invalid Flags", lineNo)
			}

			v, err := parseFlags(fields[1:])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			curDef.Flags = v
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	return is, nil
}

// parseFlags parses the flags from the tokens.
func parseFlags(tokens []string) (int, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	// numeric fast-path
	if len(tokens) == 1 {
		if v, err := strconv.Atoi(tokens[0]); err == nil {
			return v, nil
		}
	}

	// text flags
	m := map[string]int{
		"ISHorizontalTile": 1,
		"ISVerticalTile":   2,
	}

	flags := 0
	// Supports "ISHorizontalTile + ISVerticalTile" or "ISHorizontalTile ISVerticalTile".
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		t = strings.Trim(t, "|+")
		if t == "" {
			continue
		}
		if v, ok := m[t]; ok {
			flags |= v
			continue
		}
		// Allow numeric tokens like "0".
		if v, err := strconv.Atoi(t); err == nil {
			flags |= v
			continue
		}
		return 0, fmt.Errorf("invalid Flags value %q", t)
	}

	return flags, nil
}

// parseClassName parses the class name from the line.
func parseClassName(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}
	name := strings.Trim(fields[1], "{")
	if name == "" && len(fields) >= 3 {
		name = strings.Trim(fields[2], "{")
	}

	return name
}
