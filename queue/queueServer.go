package queue

import (
	"net/http"
	"sync"
)

const (
	bufferSize               = 10_000_000
	storedMessagesBufferSize = 10_000
)

type queueServer struct {
	subscriberMessageBuffer int

	serverMux http.ServeMux

	susbcriberMu sync.RWMutex

	messagesMu         sync.RWMutex
	messages           map[string]chan []byte
	messagesBufferSize int

	queue map[string][]*subscriber
}

func NewQueueServe() *queueServer {
	qs := &queueServer{
		subscriberMessageBuffer: bufferSize,
		queue:                   make(map[string][]*subscriber),
		messages:                make(map[string]chan []byte),
		messagesBufferSize:      storedMessagesBufferSize,
	}
	qs.serverMux.HandleFunc("/subscribe/{queue}", qs.subscribeHandler)
	qs.serverMux.HandleFunc("POST /publish/{queue}", qs.publishHandler)

	return qs
}

func (qs *queueServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qs.serverMux.ServeHTTP(w, r)
}
