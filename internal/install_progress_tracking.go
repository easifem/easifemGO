package internal

import (
	"io"
	"os"
	"sync"
)

// ProgressBar reports actual transferred bytes; protocols without a known
// content length keep the stage's spinner rather than inventing a percentage.
type ProgressBar struct{ stage *stageOutput }

func (p *ProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	stage := p.stage
	owned := stage == nil
	if owned {
		stage = newStageOutput(os.Stderr, "Download", animatedOutput(), quiet)
	}
	stage.mu.Lock()
	stage.bytes, stage.completed, stage.total = true, currentSize, totalSize
	stage.initialBytes = currentSize
	stage.mu.Unlock()
	return &progressReader{stream: stream, stage: stage, owned: owned}
}

type progressReader struct {
	stream   io.ReadCloser
	stage    *stageOutput
	owned    bool
	once     sync.Once
	closeErr error
	readErr  error
}

func (r *progressReader) Read(data []byte) (int, error) {
	n, err := r.stream.Read(data)
	r.stage.mu.Lock()
	r.stage.completed += int64(n)
	r.stage.mu.Unlock()
	if err != nil && err != io.EOF {
		r.readErr = err
	}
	return n, err
}

func (r *progressReader) Close() error {
	r.once.Do(func() {
		r.closeErr = r.stream.Close()
		if r.owned {
			err := r.readErr
			if err == nil {
				err = r.closeErr
			}
			r.stage.finish(err)
		}
	})
	return r.closeErr
}
