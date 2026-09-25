package exifgps

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"testing"
	"time"
)

func tinyJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 180, G: 80, B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSetGPSRoundTrip(t *testing.T) {
	src := tinyJPEG(t)
	if HasGPS(src) {
		t.Fatal("fresh JPEG should not have GPS")
	}

	when := time.Date(2024, 6, 1, 12, 30, 15, 0, time.UTC)
	pos := Position{Lat: 60.1699, Lon: 24.9384, Alt: 12.5, HasAlt: true, Time: when}
	out, err := SetGPS(src, pos)
	if err != nil {
		t.Fatal(err)
	}
	if !HasGPS(out) {
		t.Fatal("tagged JPEG missing GPS")
	}
	got, ok := ReadGPS(out)
	if !ok {
		t.Fatal("ReadGPS failed")
	}
	if math.Abs(got.Lat-pos.Lat) > 1e-5 || math.Abs(got.Lon-pos.Lon) > 1e-5 {
		t.Fatalf("coords = %v, want %v", got, pos)
	}
	if !got.HasAlt || math.Abs(got.Alt-pos.Alt) > 0.01 {
		t.Fatalf("alt = %v has=%t, want %v", got.Alt, got.HasAlt, pos.Alt)
	}

	// Second write should replace GPS, not fail, and keep the image decodable.
	pos2 := Position{Lat: -33.9249, Lon: 18.4241, Alt: -2, HasAlt: true, Time: when}
	out2, err := SetGPS(out, pos2)
	if err != nil {
		t.Fatal(err)
	}
	got, ok = ReadGPS(out2)
	if !ok {
		t.Fatal("ReadGPS after rewrite failed")
	}
	if math.Abs(got.Lat-pos2.Lat) > 1e-5 || math.Abs(got.Lon-pos2.Lon) > 1e-5 {
		t.Fatalf("rewritten coords = %v, want %v", got, pos2)
	}
	if _, err := jpeg.Decode(bytes.NewReader(out2)); err != nil {
		t.Fatalf("rewritten JPEG does not decode: %v", err)
	}
}

func TestSetGPSRejectsInvalid(t *testing.T) {
	_, err := SetGPS(tinyJPEG(t), Position{Lat: 100, Lon: 0})
	if err == nil {
		t.Fatal("expected error for invalid latitude")
	}
}
