package dj

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSynthesizeWAVHeader(t *testing.T) {
	data := synthesize(sounds["doit"])
	if len(data) <= 44 {
		t.Fatalf("wav too small: %d", len(data))
	}
	if !bytes.Equal(data[0:4], []byte("RIFF")) {
		t.Fatalf("missing RIFF header")
	}
	if !bytes.Equal(data[8:12], []byte("WAVE")) {
		t.Fatalf("missing WAVE header")
	}
	if !bytes.Equal(data[36:40], []byte("data")) {
		t.Fatalf("missing data chunk")
	}
}

func TestRenderSound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wav")
	got, err := renderSound("success", path)
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("got %q, want %q", got, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() <= 44 {
		t.Fatalf("rendered wav too small: %d", info.Size())
	}
}

func TestUnknownSound(t *testing.T) {
	if _, err := renderSound("missing", filepath.Join(t.TempDir(), "x.wav")); err == nil {
		t.Fatal("expected error")
	}
}
