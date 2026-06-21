package conns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

// MultipartFile describes a single file part in a multipart/form-data upload.
type MultipartFile struct {
	FieldName   string
	FileName    string
	ContentType string
	Content     []byte
}

// DoMultipart performs a multipart/form-data request (used by endpoints that
// take file uploads, such as creating guild stickers). Only 429 responses are
// retried, since multipart uploads are non-idempotent POSTs.
func (c *Client) DoMultipart(ctx context.Context, operation, method, path string, fields map[string]string, file MultipartFile, opts RequestOptions) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		body, contentType, err := buildMultipart(fields, file)
		if err != nil {
			return &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.cfg.APIBaseURL+path, body)
		if err != nil {
			return &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
		}
		c.setHeaders(req, opts)
		req.Header.Set("Content-Type", contentType)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if opts.Out != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, opts.Out); err != nil {
					return &APIError{Operation: operation, Method: method, Route: path,
						Status: resp.StatusCode, Message: fmt.Sprintf("decoding response: %v", err)}
				}
			}
			return nil
		}

		var de discordError
		_ = json.Unmarshal(respBody, &de)
		apiErr := &APIError{Operation: operation, Method: method, Route: path,
			Status: resp.StatusCode, Code: de.Code, Message: de.Message}
		lastErr = apiErr

		// Only 429 is safe to retry for a non-idempotent upload.
		if resp.StatusCode != http.StatusTooManyRequests || attempt == maxRetries {
			return apiErr
		}
		wait := retryAfter(resp.Header, de.RetryAfter)
		if wait <= 0 {
			wait = time.Duration(attempt+1) * 250 * time.Millisecond
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return lastErr
}

// buildMultipart assembles a multipart/form-data body from string fields and an
// optional file part, returning the body reader and its Content-Type header.
func buildMultipart(fields map[string]string, file MultipartFile) (io.Reader, string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, "", fmt.Errorf("writing form field %q: %w", k, err)
		}
	}

	if file.FieldName != "" {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, file.FieldName, file.FileName))
		if file.ContentType != "" {
			h.Set("Content-Type", file.ContentType)
		}
		part, err := mw.CreatePart(h)
		if err != nil {
			return nil, "", fmt.Errorf("creating file part: %w", err)
		}
		if _, err := part.Write(file.Content); err != nil {
			return nil, "", fmt.Errorf("writing file content: %w", err)
		}
	}

	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("finalizing multipart body: %w", err)
	}
	return &buf, mw.FormDataContentType(), nil
}
