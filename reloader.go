package akimbo

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	UrlPath    string
	Extensions []string
	Paths      []string
}

type reloader struct {
	urlPath       string
	extensionList []string
	watcher       *fsnotify.Watcher
	broadcastMap  *broadcastMap
}

func (r reloader) SSEHandler(ctx context.Context) http.HandlerFunc {
	go r.run(ctx)
	return sseHandler(r.broadcastMap)
}

func (r reloader) SSEHandlerPath() string {
	return r.urlPath
}

// should print script tag determines whether a copy-pasteable snippet
// will get printed to console on app startup so that the developer can add it to their main template manually
func (r reloader) ScriptHandler(shouldPrintScriptTag bool) http.HandlerFunc {
	if shouldPrintScriptTag {
		fmt.Println("================")
		fmt.Println("Copy paste this script tag into your root html to enable hot reload:")
		fmt.Println(r.ScriptTag())
		fmt.Println("================")
		fmt.Println()
	}
	return scriptHandler(r.SSEHandlerPath())
}

func (r reloader) ScriptHandlerPath() string {
	return r.urlPath + "_script"
}

func (r reloader) ScriptTag() string {
	return fmt.Sprintf(`<script defer src="%s"></script>`, r.ScriptHandlerPath())
}

func (r *reloader) run(ctx context.Context) {
	wg := &sync.WaitGroup{}
	shouldSendReloadSignal := false
	wg.Add(1)
	go func(goWg *sync.WaitGroup) {
		defer goWg.Done()

		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					return
				}
				if strings.ToLower(event.Op.String()) == "write" {

					for _, entry := range r.extensionList {
						if strings.HasSuffix(strings.ToLower(event.Name), strings.ToLower(entry)) {
							shouldSendReloadSignal = true
							break
						}
					}
					if len(r.extensionList) == 0 {
						shouldSendReloadSignal = true
					}
				}
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}
				slog.Error("watcher generated error", "err", err.Error())

			case <-ctx.Done():
				return
			}
		}
	}(wg)

	// rate limit so that multiple files changing at once don't overwhelm
	wg.Add(1)
	go func(goWg *sync.WaitGroup) {
		defer goWg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				if shouldSendReloadSignal {
					r.broadcastMap.sendReloadSignal()
					shouldSendReloadSignal = false
				}
			case <-ctx.Done():
				return
			}
		}
	}(wg)
	wg.Wait()
}

func NewReloader(config Config) (*reloader, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	urlPath := cleanUrlPath(config.UrlPath)

	extensions := cleanExtensions(config.Extensions)

	paths := cleanPaths(config.Paths)
	addFoldersToWatch(watcher, paths)

	broadcastMap := newBroadcastMap()
	return &reloader{
		urlPath:       urlPath,
		watcher:       watcher,
		broadcastMap:  broadcastMap,
		extensionList: extensions,
	}, nil
}

func cleanPaths(pathListRaw []string) []string {
	paths := []string{}
	if len(pathListRaw) > 0 {
		for _, entry := range pathListRaw {
			entry = strings.TrimSpace(entry)
			paths = append(paths, entry)
		}
	}
	if len(pathListRaw) == 0 {
		paths = []string{"."}
	}
	return paths
}

func cleanExtensions(extListRaw []string) []string {
	extensions := []string{}
	if len(extListRaw) > 0 {
		for _, entry := range extListRaw {
			entry = strings.TrimSpace(entry)
			if !strings.HasPrefix(entry, ".") {
				entry = "." + entry
			}
			extensions = append(extensions, entry)
		}
	}
	return extensions
}

func cleanUrlPath(urlPathRaw string) string {
	if !strings.HasPrefix(urlPathRaw, "/") {
		urlPathRaw = "/" + urlPathRaw
	}
	return urlPathRaw
}
