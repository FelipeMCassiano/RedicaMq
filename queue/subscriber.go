package queue

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}

func (qs *queueServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := qs.subscribe(w, r)
	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		return
	}
}

func (qs *queueServer) subscribe(w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	queueName := r.PathValue("queue")

	s := &subscriber{
		msgs: make(chan []byte, qs.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "conection too slow to keep up with messages")
			}
		},
	}

	qs.addSubscriber(queueName, s)
	defer qs.deleteSubscriver(queueName, s)

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}

	c = c2
	mu.Unlock()
	defer c.CloseNow()
	ctx := c.CloseRead(context.Background())
	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (qs *queueServer) addSubscriber(queueName string, s *subscriber) {
	qs.messagesMu.Lock()
	defer qs.messagesMu.Unlock()
	if queue, ok := qs.messages.queues[queueName]; ok {
		var wg sync.WaitGroup
		defer wg.Wait()
		for len(queue.messages) > 0 {
			if msg, ok := qs.messages.Pop(queueName); ok {
				send([]*subscriber{s}, msg, &wg)
			} else {
				break
			}
		}
		if len(queue.messages) == 0 {
			qs.messages.DeleteQueue(queueName)
		}
		wg.Wait()
	}

	qs.susbcriberMu.Lock()
	qs.queue[queueName] = append(qs.queue[queueName], s)
	qs.susbcriberMu.Unlock()
}

func (qs *queueServer) deleteSubscriver(queueName string, s *subscriber) {
	qs.susbcriberMu.Lock()
	qs.queue[queueName] = deleteFromSubscriberSlice(qs.queue[queueName], s)
	qs.susbcriberMu.Unlock()
}

func deleteFromSubscriberSlice(subscribers []*subscriber, s *subscriber) []*subscriber {
	i := 0
	for _, sub := range subscribers {
		if sub != s {
			subscribers[i] = sub
			i++
		}
	}

	for j := i; j < len(subscribers); j++ {
		subscribers[j] = nil
	}

	return subscribers[:i]
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}
