package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	// gcs "github.com/hashicorp/go-getter/gcs/v2"
	// s3 "github.com/hashicorp/go-getter/s3/v2"
	getter "github.com/hashicorp/go-getter/v2"
)

func get_pkg(url, source_dir, pwd string) {
	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	req := &getter.Request{
		Src:     url,
		Dst:     source_dir,
		Pwd:     pwd,
		GetMode: getter.ModeDir,
	}
	req.ProgressListener = defaultProgressBar
	wg := sync.WaitGroup{}
	wg.Add(1)

	client := getter.DefaultClient

	// Disable symlinks for all client requests
	client.DisableSymlinks = true

	getters := getter.Getters
	// getters = append(getters, new(gcs.Getter))
	// getters = append(getters, new(s3.Getter))
	client.Getters = getters

	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		res, err := client.Get(ctx, req)
		if err != nil {
			errChan <- err
			return
		}
		log.Printf("[log] :: get_pkg.go | request destination ➡️  %s", res.Dst)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Printf("[log] :: get_pkg.go | signal ➡️  %v", sig)
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		log.Fatalf("[err] :: get_pkg.go | Error downloading: %s", err)
	}
}
