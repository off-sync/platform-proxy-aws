package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setUp(ctx context.Context, t *testing.T) *SubscriptionManager {
	sm := NewSubscriptionManager(ctx)
	assert.NotNil(t, sm)

	return sm
}

func TestNewSubscriptionManagerWithNilContext(t *testing.T) {
	setUp(nil, t)
}

func TestNewSubscriptionManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	setUp(ctx, t)

	cancel()
}

func TestSubscribe(t *testing.T) {
	sm := setUp(nil, t)

	sub := sm.Subscribe()
	assert.NotNil(t, sub)
}

func TestPublish(t *testing.T) {
	sm := setUp(nil, t)

	sub := sm.Subscribe()
	assert.NotNil(t, sub)

	go func() {
		select {
		case e := <-sub:
			t.Log(e)
		}
	}()

	event := struct{}{}

	sm.Publish(event)
}

func TestCancellation(t *testing.T) {
	sm := setUp(nil, t)

	sub := sm.Subscribe()
	assert.NotNil(t, sub)

	event := struct{}{}

	go sm.Publish(event)
	go sm.Publish(event)
	go sm.Publish(event)
	sm.Publish(event)
}
