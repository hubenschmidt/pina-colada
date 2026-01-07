package sse

import (
	"fmt"
	"log"
	"sync"
)

// PubSub manages SSE subscriptions with topic-based routing
type PubSub struct {
	mu   sync.RWMutex
	subs map[string][]chan Event
}

// global instance
var defaultPubSub = &PubSub{subs: make(map[string][]chan Event)}

// Subscribe creates a channel for receiving events on a topic
func Subscribe(topic string) <-chan Event {
	return defaultPubSub.Subscribe(topic)
}

// Unsubscribe removes a channel from a topic
func Unsubscribe(topic string, ch <-chan Event) {
	defaultPubSub.Unsubscribe(topic, ch)
}

// Publish sends an event to all subscribers of a topic
func Publish(topic string, event Event) {
	defaultPubSub.Publish(topic, event)
}

// CrawlerTopic returns the topic name for a crawler config
func CrawlerTopic(configID int64) string {
	return fmt.Sprintf("crawler:%d", configID)
}

// Subscribe creates a channel for receiving events on a topic
func (ps *PubSub) Subscribe(topic string) <-chan Event {
	ch := make(chan Event, 10)
	ps.mu.Lock()
	ps.subs[topic] = append(ps.subs[topic], ch)
	ps.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel from a topic
func (ps *PubSub) Unsubscribe(topic string, ch <-chan Event) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs := ps.subs[topic]
	for i, sub := range subs {
		if sub == ch {
			ps.subs[topic] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}

	if len(ps.subs[topic]) == 0 {
		delete(ps.subs, topic)
	}
}

// Publish sends an event to all subscribers of a topic
func (ps *PubSub) Publish(topic string, event Event) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subCount := len(ps.subs[topic])
	log.Printf("SSE: publishing %s to topic %s (%d subscribers)", event.Type, topic, subCount)

	for _, ch := range ps.subs[topic] {
		select {
		case ch <- event:
		default:
			log.Printf("SSE: dropped event (buffer full) for topic %s", topic)
		}
	}
}
