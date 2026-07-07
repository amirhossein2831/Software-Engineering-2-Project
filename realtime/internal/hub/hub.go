package hub

import "sync"

type Client struct {
	send     chan []byte
	channels []string
}

func (c *Client) Send() <-chan []byte { return c.send }

type Hub struct {
	mu       sync.RWMutex
	subs     map[string]map[*Client]struct{}
	sendSize int
}

func New(sendSize int) *Hub {
	return &Hub{subs: make(map[string]map[*Client]struct{}), sendSize: sendSize}
}

func (h *Hub) Register(channels []string) *Client {
	c := &Client{send: make(chan []byte, h.sendSize), channels: channels}
	h.mu.Lock()
	for _, ch := range channels {
		if h.subs[ch] == nil {
			h.subs[ch] = make(map[*Client]struct{})
		}
		h.subs[ch][c] = struct{}{}
	}
	h.mu.Unlock()
	return c
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	for _, ch := range c.channels {
		if set := h.subs[ch]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.subs, ch)
			}
		}
	}
	h.mu.Unlock()
	close(c.send)
}

func (h *Hub) Broadcast(channel string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.subs[channel] {
		select {
		case c.send <- msg:
		default:
		}
	}
}
