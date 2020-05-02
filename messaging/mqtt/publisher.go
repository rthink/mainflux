// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	broker "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/messaging"
)

var _ messaging.Publisher = (*Publisher)(nil)

type Publisher struct {
	conn *broker.Conn
}

// NewPublisher returns NATS message Publisher.
func NewPublisher(conn *broker.Conn) messaging.Publisher {
	return Publisher{
		conn: conn,
	}
}

func (pub Publisher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	if err := pub.conn.Publish(topic, data); err != nil {
		return err
	}

	return nil
}
