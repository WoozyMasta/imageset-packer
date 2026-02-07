package imageset

import "testing"

func TestNormalizeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		camel  bool
		expect string
	}{
		{name: "snake-basic", input: "My Icon Name", camel: false, expect: "my_icon_name"},
		{name: "snake-symbols", input: "UI@@Button##OK", camel: false, expect: "ui_button_ok"},
		{name: "camel-basic", input: "my icon name", camel: true, expect: "MyIconName"},
		{name: "camel-mixed", input: "HP_Bar-42", camel: true, expect: "HpBar42"},
		{name: "empty", input: "___", camel: false, expect: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeName(tc.input, tc.camel)
			if got != tc.expect {
				t.Fatalf("NormalizeName(%q, camel=%v) = %q, want %q", tc.input, tc.camel, got, tc.expect)
			}
		})
	}
}
