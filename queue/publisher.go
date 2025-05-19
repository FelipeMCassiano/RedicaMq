package queue

import (
	"io"
	"net/http"
	"sync"
)

func (qs *queueServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.PathValue("queue")
	body := http.MaxBytesReader(w, r.Body, 8192) // a magic number
	msg, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	qs.publish(queueName, msg)
	w.WriteHeader(http.StatusAccepted)
}

func (qs *queueServer) publish(queueName string, msg []byte) {
	qs.susbcriberMu.Lock()
	subscribers := qs.queue[queueName]
	qs.susbcriberMu.Unlock()

	if (len(subscribers)) == 0 {
		qs.messagesMu.Lock()

		if _, exists := qs.messages[queueName]; !exists {
			qs.messages[queueName] = make(chan []byte, qs.messagesBufferSize)
		}
		select {
		case qs.messages[queueName] <- msg:
		default:
			<-qs.messages[queueName] // drop the oldest
			qs.messages[queueName] <- msg
		}
		qs.messagesMu.Unlock()
		return
	}

	var wg sync.WaitGroup
	send(subscribers, msg, &wg)
	wg.Wait()
}

func send(subscribers []*subscriber, msg []byte, wg *sync.WaitGroup) {
	wg.Add(len(subscribers))
	for _, s := range subscribers {
		go func(s *subscriber) {
			defer wg.Done()
			select {
			case s.msgs <- msg:
			default:
				wg.Add(1)
				defer wg.Done()
				go s.closeSlow()
			}
		}(s)
	}
}
