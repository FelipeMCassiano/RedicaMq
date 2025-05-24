package queue

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

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
type janitor struct {
	Interval time.Duration
	stop     chan bool
	messages *messagesBuffer
}

func newMessagesBuffer(ttl, janitorTime time.Duration) *messagesBuffer {
	m := &messagesBuffer{
		queues:        make(map[string]*queueMessages),
		expiredQueues: make(map[string]bool),
		ttl:           ttl,
		priority:      newMinHeap(),
	}

	m.janitor = newJanitor(m, janitorTime)
	go m.janitor.RunJanitor()

	return m
}

func newQueueMessage() *queueMessages {
	return &queueMessages{
		messages: make([]*messageTTl, 0),
	}
}

func (q *queueMessages) AddMessage(data []byte, expiration time.Time) *messageTTl {
	q.mu.Lock()
	id := generateID()
	defer q.mu.Unlock()
	msg := &messageTTl{
		ID:         id,
		data:       data,
		expiration: expiration,
		Deleted:    false,
	}
	q.messages = append(q.messages, msg)

	return msg
}

func (q *queueMessages) Pop() (*messageTTl, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, msg := range q.messages {
		if !msg.Deleted {
			msg.Deleted = true
			return msg, true
		}
	}
	return nil, false
}

func (q *queueMessages) GetMessage(id string) (*messageTTl, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, msg := range q.messages {
		if msg.ID == id && !msg.Deleted {
			return msg, true
		}
	}

	return nil, false
}

func (q *queueMessages) MarkDeleted(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, msg := range q.messages {
		if msg.ID == id {
			msg.Deleted = true
			return true
		}
	}
	return false
}

func (q *queueMessages) Compact() {
	q.mu.Lock()
	defer q.mu.Unlock()

	filtered := make([]*messageTTl, 0, len(q.messages))
	for _, msg := range q.messages {
		if !msg.Deleted {
			filtered = append(filtered, msg)
		}
	}

	q.messages = filtered
}

func (m *messagesBuffer) Push(queueName string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[queueName]
	if !exists {
		queue = newQueueMessage()
		m.queues[queueName] = queue
	}

	expiration := time.Now().Add(m.ttl)

	msg := queue.AddMessage(data, expiration)

	m.priority.Insert(minHeapItem{
		Expiration: expiration,
		QueueName:  queueName,
		MessageID:  msg.ID,
	})
}

func (m *messagesBuffer) Pop(queueName string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	queue, exists := m.queues[queueName]
	if !exists || len(queue.messages) == 0 {
		return nil, false
	}
	msg, ok := queue.Pop()
	if !ok {
		return nil, false
	}

	m.expiredQueues[queueName] = true

	m.priority.RemoveByID(queueName, msg.ID)

	return msg.data, true
}

func (m *messagesBuffer) DeleteQueue(queueName string) {
	m.mu.Lock()
	delete(m.queues, queueName)
	m.mu.Unlock()
}

func (m *messagesBuffer) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for {
		item, ok := m.priority.Peek()
		if !ok {
			break
		}

		if now.After(item.Expiration) {
			m.priority.ExtractMin()

			if queue, exists := m.queues[item.QueueName]; exists {
				m.expiredQueues[item.QueueName] = true
				queue.MarkDeleted(item.MessageID)
			}
		} else {
			break
		}
	}

	for qName := range m.expiredQueues {
		if q, ok := m.queues[qName]; ok {
			q.Compact()
		}
	}
	m.expiredQueues = make(map[string]bool)
}

func (m *messagesBuffer) Stop() {
	m.janitor.Stop()
}

func newJanitor(m *messagesBuffer, interval time.Duration) *janitor {
	return &janitor{
		Interval: interval,
		stop:     make(chan bool),
		messages: m,
	}
}

func (j *janitor) RunJanitor() {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			j.messages.cleanupExpired()
		case <-j.stop:
			return
		}
	}
}

func (j *janitor) Stop() {
	j.stop <- true
}

func generateID() string {
	id, _ := uuid.NewRandom()
	return id.String()
}
