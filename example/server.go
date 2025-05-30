package example

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hudsn/akimbo"
)

func ExampleServer(ctx context.Context, port int) {
	_, callingFile, _, _ := runtime.Caller(0)
	staticPath := filepath.Join(filepath.Dir(callingFile), "static")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	relativePath, err := filepath.Rel(cwd, staticPath)
	if err != nil {
		log.Fatal(err)
	}
	fileFn := http.FileServer(http.Dir(relativePath))

	reloadConfig := akimbo.Config{
		UrlPath:    "/akimboreload",
		Extensions: []string{"css", "js", "html"},
		Paths:      []string{"example/static"},
	}
	reloader, err := akimbo.NewReloader(reloadConfig)
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/", fileFn.ServeHTTP)
	router.HandleFunc(reloader.SSEHandlerPath(), reloader.SSEHandler(ctx))
	router.HandleFunc(reloader.ScriptHandlerPath(), reloader.ScriptHandler(true))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go server.ListenAndServe()
	slog.Info(fmt.Sprintf("example server listening on http://localhost:%d\nOpen it in your browser and edit files in the static folder to trigger the live reload.", port))

	<-ctx.Done()

	sdctx, cancel := context.WithCancel(context.Background())
	cancel()
	server.Shutdown(sdctx)
}
