package queue

import (
	"net/http"
	"sync"
	"time"
)

type (
	messagesBuffer struct {
		queues        map[string]*queueMessages
		expiredQueues map[string]bool

		mu       sync.RWMutex
		ttl      time.Duration
		janitor  *janitor
		priority *minHeap
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

func NewQueueServer() *queueServer {
	qs := &queueServer{
		subscriberMessageBuffer: bufferSize,
		queue:                   make(map[string][]*subscriber),
		messages:                newMessagesBuffer(timeDuration, janitorCleaning),
		messagesBufferSize:      storedMessagesBufferSize,
	}
	qs.serverMux.HandleFunc("/subscribe/{queue}", qs.subscribeHandler)
	qs.serverMux.HandleFunc("POST /publish/{queue}", qs.publishHandler)

	return qs
}

func (qs *queueServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qs.serverMux.ServeHTTP(w, r)
}
