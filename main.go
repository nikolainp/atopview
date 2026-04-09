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
	"github.com/nikolainp/atopview/webreporter"
)

var (
	version = "dev"
	//	commit  = "none"
	date = "unknown"
)

func init() {
}

func main() {
	ctx, cancel := withSignalNotify()
	defer cancel()

	conf, err := getConfig(os.Args)
	if err != nil {
		return
	}

	storage, err := getStorage(conf.PathStorage)
	if err != nil {
		return
	}

	if !conf.ShowReportOnly {
		var wg sync.WaitGroup

		monitor := monitor.NewMonitor()
		monitor.Start(ctx, "Parse: %[6]s - %[5]s time: %[7]s")
		transfer := logreader.NewDataTransfer(1024)

		wg.Go(func() {
			worker := logparser.NewLogParser(storage)
			worker.WithMonitor(monitor)
			worker.WithDetails(conf.PathLog, version)
			if err := worker.ReadData(ctx, transfer); err != nil {
				fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
				cancel()
			}
		})
		wg.Go(func() {
			worker := logreader.NewLogReader(conf.PathUtility, conf.PathLog)
			worker.WithMonitor(monitor)
			if err := worker.ReadData(ctx, transfer); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				cancel()
			}
			transfer.Close()
		})

		wg.Wait()
		// monitor.Start("Save data: parts: %[1]d/%[2]d time: %[5]s")

		if ctx.Err() != nil {
			return
		}

		monitor.WriteEvent("Start post processing\n")
		storage.FinishLoad()
		monitor.WriteEvent("Finish loading\n")
		monitor.Stop()
	}

	startWebServer(ctx, storage)
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

func getStorage(path string) (*storage.Storage, error) {

	db, err := storage.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Storage error: %v\n", err)
	}

	return db, err
}

func startWebServer(ctx context.Context, storage *storage.Storage) {
	reporter := webreporter.NewWebReporter(storage)
	if err := reporter.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "WebServer error: %v", err)
	}
}
