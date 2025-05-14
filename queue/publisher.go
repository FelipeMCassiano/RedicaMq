package queue

import (
	"context"
	"io"
	"net/http"
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
	defer qs.susbcriberMu.Unlock()

	qs.publishLimiter.Wait(context.Background())
	for _, s := range qs.queue[queueName] {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}
