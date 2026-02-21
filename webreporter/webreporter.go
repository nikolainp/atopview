package webreporter

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/nikolainp/atopview/storage"
)

// WebReporter ...
type WebReporter interface {
	Start(ctx context.Context) error
}

///////////////////////////////////////////////////////////////////////////////

//go:embed static
var staticContent embed.FS

//go:embed templates
var templateContent embed.FS

///////////////////////////////////////////////////////////////////////////////

type webReporter struct {
	storage   *storage.Storage
	srv       http.Server
	templates *template.Template
	logger    *log.Logger

	title  string
	filter *dataFilter

	port int

	cancelChan chan bool
}

// NewWebReporter ...
func NewWebReporter(storage *storage.Storage) WebReporter {
	var err error

	obj := new(webReporter)

	obj.port = 8090

	obj.storage = storage
	obj.logger = log.New(os.Stdout, "http: ", log.LstdFlags)

	obj.templates, err = template.ParseFS(templateContent, "templates/*.html")
	checkErr(err)

	//	details := obj.getRootDetails()
	//obj.title = details.Title
	//obj.filter = getDataFilter(obj.templates.Lookup("dataFilter.html"))
	//obj.filter.setTime(details.FirstEventTime, details.LastEventTime)

	obj.srv = http.Server{
		Handler: obj.getHandlers(),
		Addr:    fmt.Sprintf(":%d", obj.port),
	}

	return obj
}

// Start ...
func (obj *webReporter) Start(ctx context.Context) error {
	log.Printf("start web-server, port: %d\n", obj.port)

	var wg sync.WaitGroup

	wg.Go(func() {
		<-ctx.Done()

		log.Println("web-server is shutting down...")

		webCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		obj.srv.SetKeepAlivesEnabled(false)
		if err := obj.srv.Shutdown(webCtx); err != nil {
			log.Fatalf("could not gracefully shutdown the web-server: %v\n", err)
		}
	})

	err := obj.srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	wg.Wait()
	log.Println("server stopped")
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *webReporter) getHandlers() http.Handler {

	sm := http.NewServeMux()

	sm.HandleFunc("/", obj.rootPage)
	// sm.HandleFunc("/processes", obj.processes)
	// sm.HandleFunc("/performance", obj.performance)
	// sm.HandleFunc("/performance/{id}", obj.performance)
	// sm.HandleFunc("/servercontexts", obj.servercontexts)
	// sm.HandleFunc("/servercontexts/{id}", obj.servercontexts)

	sm.HandleFunc("/datafilter", obj.filter.setContext)
	sm.HandleFunc("/data/{section}/{source...}", obj.dataSource)

	sm.Handle("/static/", http.FileServer(http.FS(staticContent)))
	sm.HandleFunc("/headers", obj.headers)

	//logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	return logging(obj.logger)(sm)
}

func (obj *webReporter) headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

var checkErr = func(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func(start time.Time) {
				logger.Printf("(%s) %s %s", time.Since(start), req.Method, req.URL.Path)
			}(time.Now())
			next.ServeHTTP(w, req)
		})
	}
}
