package discord

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// supportedImageMIME lists the image formats Discord accepts for image-data
// fields (icons, banners, avatars, emoji, stickers). Discord expects an RFC 2397
// data URI of the form data:image/<type>;base64,<data>
// (https://discord.com/developers/docs/reference#image-data).
var supportedImageMIME = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
}

// EncodeImageData turns raw image bytes into the data-URI string Discord expects.
// The MIME type is detected from the content; unsupported types are rejected so
// the caller fails fast rather than sending an upload Discord will refuse.
func EncodeImageData(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("image data is empty")
	}
	mime := http.DetectContentType(raw)
	// DetectContentType can append parameters (e.g. "; charset=..."); for images
	// it returns a bare type, but normalize defensively.
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if !supportedImageMIME[mime] {
		return "", fmt.Errorf("unsupported image type %q; Discord accepts PNG, JPEG, or GIF", mime)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

// IsImageDataURI reports whether s already looks like a Discord image data URI.
func IsImageDataURI(s string) bool {
	if !strings.HasPrefix(s, "data:image/") {
		return false
	}
	i := strings.Index(s, ";base64,")
	if i < 0 {
		return false
	}
	mime := s[len("data:"):i]
	return supportedImageMIME[mime]
}

// supportedAudioMIME lists the audio formats Discord accepts for soundboard
// sounds (https://discord.com/developers/docs/resources/soundboard).
var supportedAudioMIME = map[string]bool{
	"audio/mpeg": true,
	"audio/mp3":  true,
	"audio/ogg":  true,
}

// IsAudioDataURI reports whether s looks like a base64 audio data URI Discord
// accepts for soundboard sounds.
func IsAudioDataURI(s string) bool {
	if !strings.HasPrefix(s, "data:audio/") {
		return false
	}
	i := strings.Index(s, ";base64,")
	if i < 0 {
		return false
	}
	mime := s[len("data:"):i]
	return supportedAudioMIME[mime]
}
