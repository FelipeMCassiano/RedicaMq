package queue

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

type queueServer struct {
	subscriberMessageBuffer int

	publishLimiter *rate.Limiter

	serverMux http.ServeMux

	susbcriberMu sync.Mutex

	queue map[string][]*subscriber
}

func NewQueueServe() *queueServer {
	qs := &queueServer{
		subscriberMessageBuffer: 16,
		queue:                   make(map[string][]*subscriber),
		publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}

	qs.serverMux.HandleFunc("/subscribe/{queue}", qs.subscribeHandler)
	qs.serverMux.HandleFunc("POST /publish/{queue}", qs.publishHandler)

	return qs
}

func (qs *queueServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qs.serverMux.ServeHTTP(w, r)
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
