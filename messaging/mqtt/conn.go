// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	broker "github.com/nats-io/nats.go"
)

func Connect(url string) (*broker.Conn, error) {
	conn, err := broker.Connect(url)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Close(conn *broker.Conn) {
	conn.Close()
}
