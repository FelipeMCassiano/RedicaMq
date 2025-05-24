package queue

import (
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

func newMinHeap() *minHeap {
	return &minHeap{
		items: make([]minHeapItem, 0),
	}
}

func (h *minHeap) Insert(item minHeapItem) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.items = append(h.items, item)

	h.bubbleUp(len(h.items) - 1)
}

func (h *minHeap) Peek() (minHeapItem, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.items) == 0 {
		return minHeapItem{}, false
	}

	return h.items[0], true
}

func (h *minHeap) ExtractMin() (minHeapItem, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.items) == 0 {
		return minHeapItem{}, false
	}

	min := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items = h.items[:last]
	h.bubbleDown(0)

	return min, false
}

func (h *minHeap) RemoveByID(queueName, messageID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, item := range h.items {
		if item.QueueName == queueName && item.MessageID == messageID {
			last := len(h.items) - 1
			h.items[i] = h.items[last]
			h.items = h.items[:last]
			if i < len(h.items) {
				h.bubbleUp(i)
				h.bubbleDown(i)
			}
			return true
		}
	}
	return false
}

func (h *minHeap) bubbleUp(index int) {
	for {
		parent := (index - 1) / 2
		if index == 0 || h.items[parent].Expiration.Before(h.items[index].Expiration) {
			break
		}

		h.items[parent], h.items[index] = h.items[index], h.items[parent]
		index = parent
	}
}

func (h *minHeap) bubbleDown(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		smallest := index

		if left < len(h.items) && h.items[left].Expiration.Before(h.items[smallest].Expiration) {
			smallest = left
		}

		if right < len(h.items) && h.items[right].Expiration.Before(h.items[smallest].Expiration) {
			smallest = right
		}

		if smallest == index {
			break
		}

		h.items[index], h.items[smallest] = h.items[smallest], h.items[index]
		index = smallest
	}
}
