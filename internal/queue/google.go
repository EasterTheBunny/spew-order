package queue

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

type PubSub interface {
	Publish(context.Context, string, []byte) (string, error)
}

func NewGooglePubSub(projectID string) PubSub {

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}

	return &GooglePubSub{
		client: client}
}

type GooglePubSub struct {
	client *pubsub.Client
}

func (g *GooglePubSub) Publish(ctx context.Context, topic string, data []byte) (id string, err error) {
	m := &pubsub.Message{
		Data: data,
	}

	return g.client.Topic(topic).Publish(ctx, m).Get(ctx)
}

func NewMockPubSub() *MockPubSub {
	return &MockPubSub{
		subscribers: make(map[string][]chan domain.OrderMessage)}
}

type MockPubSub struct {
	subscribers map[string][]chan domain.OrderMessage
}

func (m *MockPubSub) Publish(ctx context.Context, topic string, data []byte) (id string, err error) {

	t, ok := m.subscribers[topic]
	if ok {
		rand.Seed(time.Now().UnixNano())
		id = m.randSeq(10)

		for _, sub := range t {
			go func(s chan domain.OrderMessage) {
				s <- domain.OrderMessage{Data: data}
			}(sub)
		}

		return
	}

	err = errors.New("topic not found")
	return
}

func (m *MockPubSub) Subscribe(topic string, c chan domain.OrderMessage) {
	m.subscribers[topic] = append(m.subscribers[topic], c)
}

func (m MockPubSub) randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
