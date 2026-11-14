package utils

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestImportImageLimits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cover.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(MaxImportImageBytes + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err = ReadImportImage(path); err == nil {
		t.Fatal("oversized file accepted")
	}
	if err = os.WriteFile(path, []byte("not an image"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadImportImage(path); err == nil {
		t.Fatal("invalid image accepted")
	}
	// Original-mode imports must preserve the compressed image bytes.
	var b bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	jpeg.Encode(&b, img, nil)
	if err = os.WriteFile(path, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := ReadImportImage(path)
	if err != nil || !bytes.Equal(data, b.Bytes()) {
		t.Fatal("original image changed", err)
	}
}

func TestCropUsesCenteredSourceRectangle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 120, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			c := color.RGBA{255, 0, 0, 255}
			if x >= 40 && x < 80 {
				c = color.RGBA{0, 255, 0, 255}
			}
			img.Set(x, y, c)
		}
	}
	var b bytes.Buffer
	jpeg.Encode(&b, img, &jpeg.Options{Quality: 100})
	out, err := ResizeAndCropImage(b.Bytes(), 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	r, g, _, _ := decoded.At(10, 10).RGBA()
	if g < 50000 || r > 5000 {
		t.Fatalf("wrong crop: r=%d g=%d", r, g)
	}
	if _, err = ResizeAndCropImage(b.Bytes(), 10000, 10000); err == nil {
		t.Fatal("oversized target accepted")
	}
}

func TestImportImageRejectsHugePixelHeader(t *testing.T) {
	var b bytes.Buffer
	jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 40, 20)), nil)
	data := b.Bytes()
	sof := bytes.Index(data, []byte{0xff, 0xc0})
	if sof < 0 {
		t.Fatal("missing SOF")
	}
	binary.BigEndian.PutUint16(data[sof+5:sof+7], 5000)
	binary.BigEndian.PutUint16(data[sof+7:sof+9], 6000)
	if err := ValidateImagePixels(data); err == nil {
		t.Fatal("30 million pixel header accepted")
	}
}

func BenchmarkImportImageCrop(b *testing.B) {
	var encoded bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 6000, 4000))
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ResizeAndCropImage(encoded.Bytes(), 400, 600); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCropRejectsOverflowingDimensions(t *testing.T) {
	largestInt := int(^uint(0) >> 1)
	if _, err := ResizeAndCropImage(nil, largestInt, largestInt); err == nil || err.Error() != "目标封面超过 400 万像素" {
		t.Fatalf("overflow bypassed dimension guard: %v", err)
	}
}
