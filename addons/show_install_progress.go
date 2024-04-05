package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	getter "github.com/hashicorp/go-getter"
)

func WithProgress(pl getter.ProgressTracker) func(*getter.Client) error {
	return func(c *getter.Client) error {
		c.ProgressListener = pl
		return nil
	}
}

type MockProgressTracking struct {
	sync.Mutex
	downloaded map[string]int
}

func (p *MockProgressTracking) TrackProgress(src string,
	currentSize, totalSize int64, stream io.ReadCloser,
) (body io.ReadCloser) {
	p.Lock()
	defer p.Unlock()

	if p.downloaded == nil {
		p.downloaded = map[string]int{}
	}

	v, _ := p.downloaded[src]
	p.downloaded[src] = v + 1
	return stream
}

func TestGet_progress() {
	p := &MockProgressTracking{}
	dst := "/tmp/base"
	url := "https://github.com/easifem/base.git"
	defer os.RemoveAll(filepath.Dir(dst))
	if err := getter.GetFile(dst, url, WithProgress(p)); err != nil {
		log.Fatalf("[INTERNAL ERROR] :: show_install_progress.go | download failed: %v", err)
	}
	if p.downloaded["file"] != 1 {
		log.Println("Expected a file download")
	}
	if p.downloaded["otherfile"] != 1 {
		log.Println("Expected a otherfile download")
	}
}
