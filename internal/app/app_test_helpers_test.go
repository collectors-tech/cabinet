package app

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/update"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	return newTestAppWithConfig(t, cfg)
}

func newTestAppWithConfig(t *testing.T, cfg config.Config) *App {
	t.Helper()
	_ = os.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = a.close()
	})
	return a
}

func doRequest(t *testing.T, a *App, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(rr, req.WithContext(context.Background()))
	return rr
}

func buildMultipartPhoto(t *testing.T, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write image data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return &b, w.FormDataContentType()
}

func sampleJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 80, A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return b.Bytes()
}
