package imageset

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteSmoke(t *testing.T) {
	t.Parallel()

	is := &ImageSetClass{
		Name:    "my ui",
		RefSize: [2]int{512, 256},
		Textures: []ImageSetTextureClass{
			{Mpix: 1, Path: "mod/data/ui.edds"},
		},
		Images: []ImageSetDefClass{
			{Name: "root_icon", Pos: [2]int{1, 2}, Size: [2]int{3, 4}},
		},
		Groups: []ImageSetGroupClass{
			{
				Name: "main hud",
				Images: []ImageSetDefClass{
					{Name: "group_icon", Pos: [2]int{5, 6}, Size: [2]int{7, 8}},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Write(&buf, is, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	out := buf.String()
	contains := []string{
		"ImageSetClass {",
		`Name "my_ui"`,
		"RefSize 512 256",
		"Textures {",
		`path "mod/data/ui.edds"`,
		"Images {",
		`Name "root_icon"`,
		"Groups {",
		"ImageSetGroupClass main_hud {",
		`Name "main_hud"`,
	}
	for _, s := range contains {
		if !strings.Contains(out, s) {
			t.Fatalf("output does not contain %q\n%s", s, out)
		}
	}
}
