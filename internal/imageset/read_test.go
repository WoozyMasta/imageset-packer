package imageset

import (
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tokens  []string
		want    int
		wantErr bool
	}{
		{name: "empty", tokens: nil, want: 0},
		{name: "numeric", tokens: []string{"3"}, want: 3},
		{name: "named-single", tokens: []string{"ISHorizontalTile"}, want: 1},
		{name: "named-combined-plus", tokens: []string{"ISHorizontalTile", "+", "ISVerticalTile"}, want: 3},
		{name: "named-combined-space", tokens: []string{"ISHorizontalTile", "ISVerticalTile"}, want: 3},
		{name: "invalid", tokens: []string{"UNKNOWN_FLAG"}, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseFlags(tc.tokens)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseFlags(%v) expected error", tc.tokens)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFlags(%v) unexpected error: %v", tc.tokens, err)
			}
			if got != tc.want {
				t.Fatalf("parseFlags(%v) = %d, want %d", tc.tokens, got, tc.want)
			}
		})
	}
}

func TestParseClassName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line string
		want string
	}{
		{line: "ImageSetGroupClass GroupOne {", want: "GroupOne"},
		{line: "ImageSetGroupClass {", want: ""},
		{line: "ImageSetGroupClass    GroupTwo{", want: "GroupTwo"},
		{line: "", want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.line, func(t *testing.T) {
			t.Parallel()
			if got := parseClassName(tc.line); got != tc.want {
				t.Fatalf("parseClassName(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestReadFileRefSizeError(t *testing.T) {
	t.Parallel()

	path := writeTmpImageSetFile(t, "ImageSetClass {\n\tRefSize bad 1\n}\n")
	_, err := ReadFile(path)
	if err == nil {
		t.Fatal("expected ReadFile error for invalid RefSize")
	}
	if !strings.Contains(err.Error(), "invalid RefSize") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadFileRootAndGroups(t *testing.T) {
	t.Parallel()

	content := `ImageSetClass {
	Name "ui"
	RefSize 256 256
	Images {
		ImageSetDefClass RootIcon {
			Name "root_icon"
			Pos 1 2
			Size 3 4
			Flags 0
		}
	}
	Groups {
		ImageSetGroupClass HUD {
			Name "HUD"
			Images {
				ImageSetDefClass GroupIcon {
					Name "group_icon"
					Pos 10 20
					Size 30 40
					Flags ISHorizontalTile
				}
			}
		}
	}
}`

	path := writeTmpImageSetFile(t, content)
	is, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if is.Name != "ui" {
		t.Fatalf("name = %q, want %q", is.Name, "ui")
	}
	if is.RefSize != [2]int{256, 256} {
		t.Fatalf("refsize = %v, want [256 256]", is.RefSize)
	}
	if len(is.Images) != 1 || is.Images[0].Name != "root_icon" {
		t.Fatalf("unexpected root images: %+v", is.Images)
	}
	if len(is.Groups) != 1 {
		t.Fatalf("groups len = %d, want 1", len(is.Groups))
	}
	if is.Groups[0].Name != "HUD" {
		t.Fatalf("group name = %q, want HUD", is.Groups[0].Name)
	}
	if len(is.Groups[0].Images) != 1 || is.Groups[0].Images[0].Name != "group_icon" {
		t.Fatalf("unexpected group images: %+v", is.Groups[0].Images)
	}
	if is.Groups[0].Images[0].Flags != 1 {
		t.Fatalf("group image flags = %d, want 1", is.Groups[0].Images[0].Flags)
	}
}

func writeTmpImageSetFile(t *testing.T, content string) string {
	t.Helper()

	p := t.TempDir() + "/tmp.imageset"
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write tmp imageset: %v", err)
	}

	return p
}
