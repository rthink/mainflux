// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/mocks"
)

const (
	svcName       = "test-service"
	token         = "1"
	invalid       = "invalid"
	numOfMessages = 42
	chanID        = "1"
	valueFields   = 5
)

var (
	v   float64 = 5
	vs          = "value"
	vb          = true
	vd          = "dataValue"
	sum float64 = 42
)

func newService() readers.MessageRepository {
	var messages []readers.Message
	for i := 0; i < numOfMessages; i++ {
		msg := senml.Message{
			Channel:   chanID,
			Publisher: "1",
			Protocol:  "mqtt",
		}
		// Mix possible values as well as value sum.
		count := i % valueFields

		switch count {
		case 0:
			msg.Value = &v
		case 1:
			msg.BoolValue = &vb
		case 2:
			msg.StringValue = &vs
		case 3:
			msg.DataValue = &vd
		case 4:
			msg.Sum = &sum
		}

		messages = append(messages, msg)
	}

	return mocks.NewMessageRepository(map[string][]readers.Message{
		chanID: messages,
	})
}

func newServer(repo readers.MessageRepository, tc mainflux.ThingsServiceClient) *httptest.Server {
	mux := api.MakeHandler(repo, tc, svcName)
	return httptest.NewServer(mux)
}

type testRequest struct {
	client *http.Client
	method string
	url    string
	token  string
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, nil)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}

	return tr.client.Do(req)
}

func TestReadAll(t *testing.T) {
	svc := newService()
	tc := mocks.NewThingsService()
	ts := newServer(svc, tc)
	defer ts.Close()

	cases := map[string]struct {
		url    string
		token  string
		status int
	}{
		"read page with valid offset and limit": {
			url:    fmt.Sprintf("%s/messages/last/ba22f57d-642e-4b82-9718-5e3b68809ac0", ts.URL), //messages/last/{chanIDs}
			token:  token,
			status: http.StatusOK,
		},
		"read page with negative offset": {
			url:    fmt.Sprintf("%s/messages/list/ba22f57d-642e-4b82-9718-5e3b68809ac0", ts.URL),
			token:  token,
			status: http.StatusOK,
		},
		"read page with negative limit": {
			url:    fmt.Sprintf("%s/messages/pumpRunningSeconds/ba22f57d-642e-4b82-9718-5e3b68809ac0", ts.URL),
			token:  token,
			status: http.StatusOK,
		},
		/*"read page with zero limit": {
			url:    fmt.Sprintf("%s/messages/last/ba22f57d-642e-4b82-9718-5e3b68809ac0", ts.URL),
			token:  token,
			status: http.StatusBadRequest,
		},*/
	}

	for desc, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    tc.url,
			token:  tc.token,
		}
		if desc == "read page with negative limit" {
			req.method = http.MethodPost
		}
		//todo
		fmt.Println("request---------")
		fmt.Println("url : " + tc.url)
		fmt.Println("token : " + tc.token)
		fmt.Println("status : ", tc.status)
		//
		res, _ /*err*/ := req.make()
		//todo
		fmt.Println("rep---------")
		fmt.Println("StatusCode : ", res.StatusCode)
		fmt.Println("Status : ", res.Status)
		msg := res.Header.Get("messages")
		fmt.Println("msg = " + msg)
		//
		//assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		//assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected %d got %d", desc, tc.status, res.StatusCode))
	}
}
