package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	getter "github.com/hashicorp/go-getter"
)

func main() {
	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	} else {
		log.Printf("pwd = %s", pwd)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &getter.Client{
		Ctx:  ctx,
		Src:  "github.com/easifem/md2src",
		Dst:  "/tmp/md2src",
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
		// Options: opts,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	c := make(chan os.Signal)
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
