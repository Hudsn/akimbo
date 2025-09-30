package example

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/hudsn/spicyreload"
)

func noCacheHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		h.ServeHTTP(w, r)
	})
}
func ExampleServer(ctx context.Context, port int) {

	fileFn := http.FileServer(http.Dir("example/static"))

	reloadConfig := spicyreload.Config{
		UrlPath:    "/spicyreload",
		Extensions: []string{"css", "js", "html"},
		Paths:      []string{"example/static"},
	}
	reloader, err := spicyreload.NewReloader(reloadConfig)
	if err != nil {
		log.Fatal(err)
	}
	noCacheHandler(fileFn)
	router := http.NewServeMux()
	// router.HandleFunc("/", fileFn.ServeHTTP)
	router.HandleFunc("/", noCacheHandler(fileFn).ServeHTTP)
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
