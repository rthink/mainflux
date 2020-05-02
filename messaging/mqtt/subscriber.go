// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"

	broker "github.com/eclipse/paho.mqtt.golang"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/messaging"
)

var (
	errAlreadySubscribed = errors.New("already subscribed to topic")
	errNotSubscribed     = errors.New("not subscribed")
	errEmptyTopic        = errors.New("empty topic")
)

var _ messaging.Subscriber = (*Subscriber)(nil)

type Subscriber struct {
	conn          *broker.Conn
	logger        log.Logger
	mu            sync.Mutex
	subscriptions map[string]*broker.Subscription
}

// NewPublisher returns NATS message Subscriber
func NewSubscriber(conn *broker.Conn, logger log.Logger) Subscriber {
	return Subscriber{
		conn:   conn,
		logger: logger,
	}
}

// Parameter queue specifies the queue for the Subscribe method.
// If queue is specified (is not an empty string), Subscribe method
// will execute NATS QueueSubscribe which is conceptually different
// from ordinary subscribe. For more information, please take a look
// here: https://docs.nats.io/developing-with-nats/receiving/queues.
// If the queue is empty, Subscribe will be used.
func (s *Subscriber) Subscribe(topic string, queue string, handler messaging.MessageHandler) error {
	if topic == "" {
		return errEmptyTopic
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}
	nh := s.mqttHandler(handler)
	if queue != "" {
		sub, err := s.conn.QueueSubscribe(topic, queue, nh)
		if err != nil {
			return err
		}
		s.subscriptions[topic] = sub
		return nil
	}
	sub, err := s.conn.Subscribe(topic, nh)
	if err != nil {
		return err
	}
	s.subscriptions[topic] = sub
	return nil
}

func (s *Subscriber) Unsubscribe(topic string) error {
	if topic == "" {
		return errEmptyTopic
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	topic = fmt.Sprintf("%s.%s", chansPrefix, topic)

	sub, ok := s.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}

	if err := sub.Unsubscribe(); err != nil {
		return err
	}

	delete(s.subscriptions, topic)
	return nil
}

func (s *Subscriber) mqttHandler(h messaging.MessageHandler) broker.MsgHandler {
	return func(m *broker.Msg) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h(msg); err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
	}
}
