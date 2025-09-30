package spicyreload

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// sse handler

func sseHandler(bmap *broadcastMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chanId, eventChan := bmap.newChannel()
		defer bmap.dropChannel(chanId)
		rc := http.NewResponseController(w)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/event-stream")
		if r.ProtoMajor == 1 {
			w.Header().Set("Connection", "keep-alive")
		}
		rc.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case <-eventChan:
				jsonb, _ := json.Marshal(map[string]string{"message": "reload"})
				fmt.Fprintf(w, "data: %s\n\n", string(jsonb))
				rc.Flush()
			}
		}
	}
}

// broadcaster

type broadcastMap struct {
	mu         *sync.Mutex
	channelMap map[string]chan bool
}

func newBroadcastMap() *broadcastMap {
	return &broadcastMap{
		mu:         &sync.Mutex{},
		channelMap: make(map[string]chan bool),
	}
}

// returns channel ID and the reciever channel
func (bm *broadcastMap) newChannel() (string, chan bool) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	found := true
	var randStr string
	for found {
		randStr = randomId()
		_, found = bm.channelMap[randStr]
	}
	newChan := make(chan bool)
	bm.channelMap[randStr] = newChan
	return randStr, newChan
}

func randomId() string {
	buf := make([]byte, 12)
	rand.Read(buf)
	return fmt.Sprintf("%x", sha256.Sum256(buf))[:10]
}

func (bm *broadcastMap) dropChannel(chanId string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	c, found := bm.channelMap[chanId]
	if !found {
		return
	}
	delete(bm.channelMap, chanId)
	close(c)
	// drain events so we don't hang forever
	for range c {
		continue
	}
}

func (bm *broadcastMap) sendReloadSignal() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, sendChan := range bm.channelMap {
		select {
		case sendChan <- true:
		default:
		}
	}
}
