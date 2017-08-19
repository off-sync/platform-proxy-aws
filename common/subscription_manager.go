package common

import (
	"context"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type (
	eventType interface{}

	// SubscriptionChan models a channel through which events can be received.
	SubscriptionChan <-chan eventType
	subscriptions    map[string]chan eventType
)

// Defaults.
const (
	DefaultPublishTimeout = 1 * time.Second
)

// SubscriptionManager manages a set of subscriptions and allows events to be
// published to them.
type SubscriptionManager struct {
	subscriptionsLock *sync.RWMutex
	subscriptions     subscriptions

	publishTimeout time.Duration

	ctx           context.Context
	cancellations chan string
}

// NewSubscriptionManager creates a new SubscriptionManager.
func NewSubscriptionManager(ctx context.Context) *SubscriptionManager {
	if ctx == nil {
		ctx = context.Background()
	}

	sm := &SubscriptionManager{
		subscriptionsLock: &sync.RWMutex{},
		subscriptions:     make(subscriptions),
		publishTimeout:    DefaultPublishTimeout,
		ctx:               ctx,
		cancellations:     make(chan string, 10),
	}

	go sm.processCancellations()

	return sm
}

func (sm *SubscriptionManager) processCancellations() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case id := <-sm.cancellations:
			sm.subscriptionsLock.Lock()
			defer sm.subscriptionsLock.Unlock()

			if ch, found := sm.subscriptions[id]; found {
				close(ch)
				delete(sm.subscriptions, id)
			}

			break
		}
	}
}

// Subscribe creates a new subscription and returns the channel through which
// events will be sent.
func (sm *SubscriptionManager) Subscribe() SubscriptionChan {
	cs := make(chan eventType)
	id := uuid.NewV4().String()

	sm.subscriptionsLock.Lock()
	defer sm.subscriptionsLock.Unlock()
	sm.subscriptions[id] = cs

	return cs
}

// Publish publishes an event to all subscribers. Every subscriber has the
// configured publishTimeout duration to process the event. If it does not
// respond in time, the subscription is cancelled by closing the channel.
func (sm *SubscriptionManager) Publish(event interface{}) {
	sm.subscriptionsLock.RLock()
	defer sm.subscriptionsLock.RUnlock()

	wg := &sync.WaitGroup{}
	wg.Add(len(sm.subscriptions))

	for id := range sm.subscriptions {
		go sm.sendEvent(wg, id, event)
	}

	wg.Wait()
}

func (sm *SubscriptionManager) sendEvent(wg *sync.WaitGroup, id string, event eventType) {
	defer wg.Done()

	timeout := time.NewTimer(sm.publishTimeout)
	defer timeout.Stop()

	cs := sm.subscriptions[id]

	select {
	case cs <- event:
		// event was sent
		return
	case <-timeout.C:
		// timeout occurred
		sm.cancellations <- id
		return
	}
}
