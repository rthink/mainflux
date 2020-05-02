// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/messaging"
	opentracing "github.com/opentracing/opentracing-go"
)

func handler() {
	// TODO: implement handler
}

func Relay(from messaging.Subscriber, to messaging.Publisher, logger logger.Logger, tracer opentracing.Tracer) {
	// TODO: install MQTT handler on subsctibe to get the data and write it in the Publisher
}
