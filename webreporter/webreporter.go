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

	title, version string
	filter         *dataFilter
	mainMenu       *navigation
	counters       map[int]string

	port int
}

// NewWebReporter ...
func NewWebReporter(storage *storage.Storage) WebReporter {
	var err error

	obj := new(webReporter)

	obj.port = 8090

	obj.storage = storage
	obj.logger = log.New(os.Stdout, "http: ", log.LstdFlags)

	obj.templates, err = template.
		New("main").
		Delims("/*{{", "}}*/").
		ParseFS(templateContent, "templates/*.html")
	checkErr(err)

	details := obj.getRootDetails()
	obj.title = details.Title
	obj.version = details.Version

	obj.filter = newDataFilter(obj.templates.Lookup("dataFilter.html"), details.FirstEventTime, details.LastEventTime)
	obj.restoreDataFilter()

	obj.mainMenu = newNavigation(obj.templates.Lookup("mainmenu.html"), []webAnchor{
		{"/display", "data display"},
		{"/counters", "system counters"},
		{"/information", "system information"},
		{"/processes", "processes counters"},
	})

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
		obj.stopServer()
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

func (obj *webReporter) stopServer() {
	log.Println("web-server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	obj.srv.SetKeepAlivesEnabled(false)
	if err := obj.srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not gracefully shutdown the web-server: %v\n", err)
	}
}

func (obj *webReporter) getHandlers() http.Handler {

	sm := http.NewServeMux()

	sm.HandleFunc("/", obj.rootPage)
	sm.HandleFunc("/display", obj.dataDisplayPage)
	sm.HandleFunc("/counters", obj.countersPage)
	sm.HandleFunc("/information", obj.informationPage)
	// sm.HandleFunc("/performance/{id}", obj.performance)
	// sm.HandleFunc("/servercontexts", obj.servercontexts)
	// sm.HandleFunc("/servercontexts/{id}", obj.servercontexts)

	sm.HandleFunc("/datafilter", obj.dataFilter)
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
