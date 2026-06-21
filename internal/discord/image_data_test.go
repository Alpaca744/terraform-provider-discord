package discord

import (
	"strings"
	"testing"
)

// pngHeader is a minimal valid PNG signature followed by an IHDR chunk start,
// enough for http.DetectContentType to classify it as image/png.
var pngHeader = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}

var gifHeader = []byte("GIF89a")

func TestEncodeImageData(t *testing.T) {
	got, err := EncodeImageData(pngHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Errorf("got %q, want image/png data URI", got)
	}

	if _, err := EncodeImageData(gifHeader); err != nil {
		t.Errorf("gif should be supported: %v", err)
	}

	if _, err := EncodeImageData(nil); err == nil {
		t.Error("empty input should error")
	}

	if _, err := EncodeImageData([]byte("just plain text, not an image")); err == nil {
		t.Error("non-image should error")
	}
}

func TestIsImageDataURI(t *testing.T) {
	if !IsImageDataURI("data:image/png;base64,AAAA") {
		t.Error("valid png data URI should pass")
	}
	if IsImageDataURI("data:image/webp;base64,AAAA") {
		t.Error("unsupported mime should fail")
	}
	if IsImageDataURI("https://example.com/x.png") {
		t.Error("plain URL should fail")
	}
}
