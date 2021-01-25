// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"fmt"
	influxdata "github.com/influxdata/influxdb/client/v2"
	"log"
	"sync"

	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/influxdb"
)

var _ readers.MessageRepository = (*messageRepositoryMock)(nil)

type messageRepositoryMock struct {
	mutex    sync.Mutex
	messages map[string][]readers.Message
}

func (repo *messageRepositoryMock) GetLastMeasurement(chanIDs []string, query map[string]string) (readers.MessagesPage, error) {
	connInflux := ConnInflux()
	influx := influxdb.New(connInflux, "mainflux")
	return influx.GetLastMeasurement(chanIDs, query)
}

func (repo *messageRepositoryMock) PumpRunningSeconds(chanIDs []string, query map[string]string) (readers.MessagesPage, error) {
	connInflux := ConnInflux()
	influx := influxdb.New(connInflux, "mainflux")
	return influx.PumpRunningSeconds(chanIDs, query)
}

func (repo *messageRepositoryMock) GetMessageByPublisher(chanID string, offset, limit uint64, aggregationType string, interval string, query map[string]string) (readers.MessagesPage, error) {
	connInflux := ConnInflux()
	influx := influxdb.New(connInflux, "mainflux")
	return influx.GetMessageByPublisher(chanID, offset, limit, aggregationType, interval, query)
}

// NewMessageRepository returns mock implementation of message repository.
func NewMessageRepository(messages map[string][]readers.Message) readers.MessageRepository {
	return &messageRepositoryMock{
		mutex:    sync.Mutex{},
		messages: messages,
	}
}

func ConnInflux() influxdata.Client {
	cli, err := influxdata.NewHTTPClient(influxdata.HTTPConfig{
		Addr:     "http://118.31.19.149:8086", /*"http://127.0.0.1:8086"*/
		Username: "mainflux",                  /*"admin"*/
		Password: "mainflux",                  /*""*/
	})
	if err != nil {
		log.Fatal(err)
	}
	return cli
}

func (repo *messageRepositoryMock) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	end := offset + limit

	numOfMessages := uint64(len(repo.messages[chanID]))
	if offset < 0 || offset >= numOfMessages {
		return readers.MessagesPage{}, nil
	}

	if limit < 1 {
		return readers.MessagesPage{}, nil
	}

	if offset+limit > numOfMessages {
		end = numOfMessages
	}

	return readers.MessagesPage{
		Total:    numOfMessages,
		Limit:    limit,
		Offset:   offset,
		Messages: repo.messages[chanID][offset:end],
	}, nil
}
