package akimbo

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed inject.js
var injectBytes []byte

func replacePathInScript(ssePath string) []byte {
	return []byte(strings.ReplaceAll(string(injectBytes), "~~urlpath~~", ssePath))
}

func scriptHandler(ssePath string) http.HandlerFunc {
	scriptBytes := replacePathInScript(ssePath)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(scriptBytes)
	}
}
