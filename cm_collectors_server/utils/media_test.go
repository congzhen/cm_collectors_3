package utils

import "testing"

func TestIsClearlyNonVideoSource(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{src: `D:\media\cover.JPG`, want: true},
		{src: "https://example.com/poster.png?size=large", want: true},
		{src: "https://example.com/info.HTML#details", want: true},
		{src: "info.HTML#details", want: false},
		{src: "folder.jpg#name/movie.mp4", want: false},
		{src: "movie.mp4", want: false},
		{src: "movie.mkv", want: false},
		{src: "uncommon.container", want: false},
		{src: "https://example.com/live", want: false},
	}
	for _, test := range tests {
		if got := IsClearlyNonVideoSource(test.src); got != test.want {
			t.Errorf("IsClearlyNonVideoSource(%q) = %v, want %v", test.src, got, test.want)
		}
	}
}
