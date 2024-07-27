package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupErrorHandlerTest(t *testing.T) (string, func()) {
	path := "./error_pages"
	err := os.MkdirAll(path, os.ModePerm)
	assert.NoError(t, err)
	cleanup := func() {
		os.RemoveAll(path)
	}

	errorFileContent := "<html><body>Error Page</body></html>"
	err = os.WriteFile(fmt.Sprintf("%s/404.html", path), []byte(errorFileContent), 0644)
	assert.NoError(t, err)
	return path, cleanup
}

func TestErrorHandler_Panic(t *testing.T) {
	path, cleanup := setupErrorHandlerTest(t)
	defer cleanup()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	handler := errorHandler(path, "invalid/format")
	_ = handler
}

func TestErrorHandler_ErrorReadingMediaTypeExtension(t *testing.T) {
	path, cleanup := setupErrorHandlerTest(t)
	defer cleanup()

	logBuffer := new(strings.Builder)
	log.SetOutput(logBuffer)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(FormatHeader, "emptytype/")

	rr := httptest.NewRecorder()
	handler := errorHandler(path, "text/html")
	handler(rr, req)

	assert.Contains(t, logBuffer.String(), "unexpected error reading media type extension")
	assert.Equal(t, 404, rr.Code)
	assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
}

func TestErrorHandler_CouldntGetMediaTypeExtension(t *testing.T) {
	path, cleanup := setupErrorHandlerTest(t)
	defer cleanup()

	logBuffer := new(strings.Builder)
	log.SetOutput(logBuffer)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(FormatHeader, "application/x-unknown")

	rr := httptest.NewRecorder()
	handler := errorHandler(path, "text/html")
	handler(rr, req)

	assert.Contains(t, logBuffer.String(), "couldn't get media type extension. Using")
	assert.Equal(t, 404, rr.Code)
	assert.Equal(t, "application/x-unknown", rr.Header().Get("Content-Type"))
}

func TestErrorHandler_FileNotFound(t *testing.T) {
	handler := errorHandler("", "text/html")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(CodeHeader, "500")
	req.Header.Set(FormatHeader, "text/html")

	rr := httptest.NewRecorder()
	handler(rr, req)

	assert.Equal(t, 500, rr.Code)
	assert.Contains(t, rr.Body.String(), "404 page not found\n")
}

func TestErrorHandler_4xx(t *testing.T) {
	tempDir := t.TempDir()

	errorFilePath := fmt.Sprintf("%s/4xx.html", tempDir)
	err := os.WriteFile(errorFilePath, []byte("404 - Not Found"), 0666)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Code", "404")
	req.Header.Set("Format", "text/html")

	rr := httptest.NewRecorder()
	handler := errorHandler(tempDir, "text/html")
	handler(rr, req)

	assert.Equal(t, 404, rr.Code)
	assert.Contains(t, rr.Body.String(), "404 - Not Found")
}

func TestErrorHandler_Debug(t *testing.T) {
	path, cleanup := setupErrorHandlerTest(t)
	defer cleanup()

	t.Setenv("DEBUG", "1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Code", "404")
	req.Header.Set("X-Format", "text/html")
	req.Header.Set("Content-Type", "text/html")

	rr := httptest.NewRecorder()
	handler := errorHandler(path, "text/html")
	handler(rr, req)

	assert.Equal(t, 404, rr.Code)
	assert.Equal(t, "404", rr.Header().Get("X-Code"))
	assert.Equal(t, "text/html", rr.Header().Get("X-Format"))
	assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
}
