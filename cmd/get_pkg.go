package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	getter "github.com/hashicorp/go-getter/v2"
)

// get package from web
func get_pkg(url, source_dir, pwd string) {
	// Get the pwd

	ctx, cancel := context.WithCancel(context.Background())

	client := &getter.Client{
		Ctx:  ctx,
		Src:  url,
		Dst:  source_dir,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		err := client.Get()
		if err != nil {
			errChan <- err
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Printf("signal %v", sig)
	case <-ctx.Done():
		wg.Wait()
		log.Printf("success!")
	case err := <-errChan:
		wg.Wait()
		log.Fatalf("Error downloading: %s", err)
	}
}
