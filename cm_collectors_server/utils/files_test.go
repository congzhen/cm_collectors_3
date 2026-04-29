package utils

import (
	"reflect"
	"testing"
)

func TestSortFilesByOrderFileNameAscNatural(t *testing.T) {
	files := []string{
		"D:/video/show/10.mp4",
		"D:/video/show/02.mp4",
		"D:/video/show/1.mp4",
		"D:/video/show/第11集.mp4",
		"D:/video/show/第3集.mp4",
	}

	actual := SortFilesByOrder(files, FileNameAsc)
	expected := []string{
		"D:/video/show/1.mp4",
		"D:/video/show/02.mp4",
		"D:/video/show/第3集.mp4",
		"D:/video/show/10.mp4",
		"D:/video/show/第11集.mp4",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("SortFilesByOrder() = %#v, want %#v", actual, expected)
	}
}
