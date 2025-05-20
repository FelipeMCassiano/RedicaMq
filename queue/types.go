package queue

import (
	"net/http"
	"sync"
	"time"
)

type minHeapItem struct {
	MessageID string

	Expiration time.Time
	QueueName  string
}
type minHeap struct {
	items []minHeapItem
	mu    sync.RWMutex
}

type messageTTl struct {
	ID         string
	data       []byte
	expiration time.Time
	Deleted    bool
}

type queueMessages struct {
	messages []*messageTTl
	mu       sync.RWMutex
}
type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}
type janitor struct {
	Interval time.Duration
	stop     chan bool
	messages *messagesBuffer
}

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
