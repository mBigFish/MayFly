package file

import "testing"

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", ".", false},
		{".", ".", false},
		{"/etc/passwd", "/etc/passwd", false},
		{"a/b/../c", "a/c", false},
		{"../etc", "", true}, // 向上越界
		{"..", "", true},     // 向上越界
	}

	for _, c := range cases {
		got, err := normalizePath(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("normalizePath(%q) 应报错", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizePath(%q) 不应报错: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("normalizePath(%q) = %q，期望 %q", c.in, got, c.want)
		}
	}
}
