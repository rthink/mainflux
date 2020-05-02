// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package forwarder

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/messaging"
	broker "github.com/nats-io/nats.go"
)

func forwarder(msg messaging.Message) error {
	return nil
} 