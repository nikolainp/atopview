package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nikolainp/atopview/config"
	"github.com/nikolainp/atopview/logparser"
	"github.com/nikolainp/atopview/logreader"
	"github.com/nikolainp/atopview/monitor"
	"github.com/nikolainp/atopview/storage"
)

var (
	version = "dev"
	//	commit  = "none"
	date = "unknown"
)

func init() {
}

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := withSignalNotify()
	defer cancel()

	conf, err := getConfig(os.Args)
	if err != nil {
		return
	}

	var storage *storage.Storage
	if !conf.ShowReportOnly {
		if storage, err = getNewStorage(); err != nil {
			return
		}

		monitor := monitor.NewMonitor()
		monitor.Start(ctx, "")

		dataTrasfer := make(chan []byte)
		wg.Go(func() {
			worker := logparser.NewLogParser()
			worker.WithMonitor(monitor)
			worker.ReadData(ctx, dataTrasfer)
		})
		wg.Go(func() {
			worker := logreader.NewLogReader(conf.PathUtilinty, conf.PathLog)
			if errText, err := worker.ReadData(ctx, dataTrasfer); err != nil {
				fmt.Fprintf(os.Stderr, "atop error: %v\n", err)
				fmt.Fprintf(os.Stderr, "%s", errText)
				cancel()
			}
			close(dataTrasfer)
		})

		wg.Wait()
		// monitor.Start("Save data: parts: %[1]d/%[2]d time: %[5]s")
		monitor.Stop()
		storage.FlushAll(conf.PathStorage)
	} else {
		if storage, err = getOldStorage(conf.PathStorage); err != nil {
			return
		}
	}

	// startWebServer(storage, cancelChan)
}

///////////////////////////////////////////////////////////////////////////////

func withSignalNotify() (context.Context, context.CancelFunc) {
	signChan := make(chan os.Signal, 10)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	ctxCancel, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case signal := <-signChan:
			// Run Cleanup
			fmt.Fprintf(os.Stderr, "\nCaptured %v, stopping and exiting...\n", signal)
			cancel()
		case <-ctxCancel.Done():
			return
		}
	}()

	return ctxCancel, cancel
}

func getConfig(args []string) (config.Config, error) {
	conf, err := config.NewConfig(args)

	if err != nil {
		switch err := err.(type) {
		case config.PrintVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case config.PrintUsage:
			fmt.Fprint(os.Stderr, err.Usage)
		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
	}

	return conf, err
}

///////////////////////////////////////////////////////////////////////////////

func getNewStorage() (*storage.Storage, error) {

	db, err := storage.CreateCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Storage error: %v\n", err)
	}

	return db, err
}

func getOldStorage(path string) (*storage.Storage, error) {

	db, err := storage.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Storage error: %v\n", err)
	}

	return db, err
}

// func startWebServer(storage *storage.Storage, isCancelChan chan bool) {
// 	reporter := webreporter.New(storage, isCancelChan)
// 	if err := reporter.Start(); err != nil {
// 		fmt.Fprintf(os.Stderr, "WebServer error: %v", err)
// 		cancelAndExit()
// 	}
// }
