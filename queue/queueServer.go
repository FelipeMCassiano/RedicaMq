package queue

import (
	"net/http"
	"sync"
	"time"

	"github.com/FelipeMCassiano/RedicaMq/config"
)

type (
	messagesBuffer struct {
		queues        map[string]*queueMessages
		expiredQueues map[string]bool

		mu         sync.RWMutex
		ttl        time.Duration
		janitor    *janitor
		priority   *minHeap
		bufferSize int
	}
	queueServer struct {
		subscriberMessageBuffer int

		serverMux http.ServeMux

		susbcriberMu sync.RWMutex

		messagesMu         sync.RWMutex
		messages           *messagesBuffer
		messagesBufferSize int

		queue map[string][]*subscriber
	}
)

const (
	bufferSize               = 10_000_000
	storedMessagesBufferSize = 10_000
	timeDuration             = 1 * time.Minute
	janitorCleaning          = timeDuration / 2
)

func NewQueueServer(cfg *config.Config) *queueServer {
	qs := &queueServer{
		subscriberMessageBuffer: cfg.Subscriber.BufferSize,
		queue:                   make(map[string][]*subscriber),
		messages:                newMessagesBuffer(cfg.Messages.TimeToLive, cfg.Messages.BufferSize),
	}
	qs.serverMux.HandleFunc("/subscribe/{queue}", qs.subscribeHandler)
	qs.serverMux.HandleFunc("POST /publish/{queue}", qs.publishHandler)

	return qs
}

func (qs *queueServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qs.serverMux.ServeHTTP(w, r)
}
