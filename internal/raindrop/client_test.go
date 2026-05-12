package raindrop

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestClientSendsBearerTokenAndDecodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/user" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	var raw json.RawMessage
	if err := client.Do(context.Background(), http.MethodGet, "user", nil, nil, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"result":true}` {
		t.Fatalf("raw = %s", raw)
	}
}

func TestClientReportsAPIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"result":false,"errorMessage":"Unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	err := client.Do(context.Background(), http.MethodGet, "user", nil, nil, nil)
	if err == nil || err.Error() != "raindrop api status 401: Unauthorized" {
		t.Fatalf("err = %v", err)
	}
}

func TestClientMultipartUploadsFileAndFields(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "upload-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatal(err)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello" {
			t.Fatalf("file data = %q", data)
		}
		if got := r.FormValue("collectionId"); got != "123" {
			t.Fatalf("collectionId = %q", got)
		}
		_, _ = w.Write([]byte(`{"result":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	raw, err := client.Multipart(context.Background(), http.MethodPut, "raindrop/file", nil, "file", tmp.Name(), map[string]string{"collectionId": "123"})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"result":true}` {
		t.Fatalf("raw = %s", raw)
	}
}
