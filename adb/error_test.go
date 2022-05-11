package adb

import "testing"

func TestParseError(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{"FAILED_DEXOPT", ErrDexopt},
		{"PARSE_FAILED_NOT_APK", ErrNotApk},
		{"FAILED_ABORTED", ErrAborted},
	}

	for _, c := range tests {
		got := parseError(c.in)
		if got != c.want {
			t.Fatalf("Parse error in %s - want %v, got %v", c.in, c.want, got)
		}
	}
}
