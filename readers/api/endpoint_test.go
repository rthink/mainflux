// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestGetLastMeasurement(t *testing.T) {
	svc := newService()
	tc := mocks.NewThingsService()
	ts := newServer(svc, tc)
	defer ts.Close()
	//fmt.Println(ts.URL)
	//for ;true; {
	//
	//}

	cases := []struct {
		url   string
		token string
	}{
		{
			url:   fmt.Sprintf("%s/messages/last/e17d05dc-f022-40b1-90b5-e7cb7de50ab0,4a68d620-edb4-4882-82fc-f5c0fccc0c31", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/last/8723d86d-85d5-4eba-9da2-3d9855772bf0,e17d05dc-f022-40b1-90b5-e7cb7de50ab0", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/last/4a68d620-edb4-4882-82fc-f5c0fccc0c31", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/last/83374d41-64b2-47a7-8981-e8ccd09c9159", ts.URL),
			token: token,
		},
	}
	for i := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    cases[i].url,
			token:  cases[i].token,
		}
		log.Println("request:")
		log.Println("url : " + cases[i].url)
		log.Println("token : " + cases[i].token)

		res, err := http.NewRequest(req.method, req.url, nil)
		q := res.URL.Query() // Query解析RawQuery并返回相应的值。也就是解析？后面的参数
		q.Add("name", "pump_1")
		res.URL.RawQuery = q.Encode() //重新赋值给URL的RawQuery字段
		//res.Header.Add("content-type","application/x-www-form-urlencoded")
		if err != nil {

		}
		if req.token != "" {
			res.Header.Set("Authorization", req.token)
		}
		do, err := req.client.Do(res)

		log.Println("rep:")
		log.Println("Status : ", do.Status)
		msg := do.Header.Get("messages")
		log.Println("msg = " + msg)
		temp := make([]byte, 10240)
		//temp := []byte{}//这种声明定义  切片的大小为0 不能用于read读取数据
		do.Body.Read(temp)

		log.Println("temp :" + string(temp))
		do.Body.Close()
	}
}

func TestPumpRunningSeconds(t *testing.T) {
	svc := newService()
	tc := mocks.NewThingsService()
	ts := newServer(svc, tc)
	defer ts.Close()
	cases := []struct {
		url   string
		token string
	}{
		{
			url:   fmt.Sprintf("%s/messages/pumpRunningSeconds/e17d05dc-f022-40b1-90b5-e7cb7de50ab0,4a68d620-edb4-4882-82fc-f5c0fccc0c31", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/pumpRunningSeconds/8723d86d-85d5-4eba-9da2-3d9855772bf0,e17d05dc-f022-40b1-90b5-e7cb7de50ab0", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/pumpRunningSeconds/4a68d620-edb4-4882-82fc-f5c0fccc0c31", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/pumpRunningSeconds/83374d41-64b2-47a7-8981-e8ccd09c9159,e17d05dc-f022-40b1-90b5-e7cb7de50ab0,8723d86d-85d5-4eba-9da2-3d9855772bf0", ts.URL),
			token: token,
		},
	}

	for i := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodPost,
			url:    cases[i].url,
			token:  cases[i].token,
		}
		log.Println("request:")
		log.Println("url : " + cases[i].url)
		log.Println("token : " + cases[i].token)
		m := url.Values{
			"from": {
				"2021-01-21T04:33:18Z",
			},
			"to": {
				"2021-01-21T05:38:18Z",
			},
			"name": {
				"pump_1,pump_2",
			},
		}
		res, err := http.NewRequest(req.method, req.url, strings.NewReader(m.Encode()))
		res.Header.Add("content-type", "application/x-www-form-urlencoded")
		if err != nil {
			//return nil, err
		}
		if req.token != "" {
			res.Header.Set("Authorization", req.token)
		}
		do, err := req.client.Do(res)

		//todo
		log.Println("rep:")
		log.Println("Status : ", do.Status)
		msg := do.Header.Get("messages")
		log.Println("msg = " + msg)
		temp := make([]byte, 10240)
		//temp := []byte{}//这种声明定义  切片的大小为0 不能用于read读取数据
		do.Body.Read(temp)

		log.Println("temp :" + string(temp))
		do.Body.Close()
	}
}

func TestGetMessageByPublisher(t *testing.T) {
	svc := newService()
	tc := mocks.NewThingsService()
	ts := newServer(svc, tc)
	defer ts.Close()
	cases := []struct {
		url   string
		token string
	}{
		{
			url:   fmt.Sprintf("%s/messages/list/ba22f57d-642e-4b82-9718-5e3b68809ac0", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/list/8723d86d-85d5-4eba-9da2-3d9855772bf0", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/list/4a68d620-edb4-4882-82fc-f5c0fccc0c31", ts.URL),
			token: token,
		},
		{
			url:   fmt.Sprintf("%s/messages/list/83374d41-64b2-47a7-8981-e8ccd09c9159", ts.URL),
			token: token,
		},
	}

	for i := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    cases[i].url,
			token:  cases[i].token,
		}
		//todo
		fmt.Println("request---------")
		fmt.Println("url : " + cases[i].url)
		fmt.Println("token : " + cases[i].token)
		res, err := http.NewRequest(req.method, req.url, nil)

		q := res.URL.Query()
		q.Add("name", "pump_1")
		q.Add("from", "2021-01-21T04:33:18Z")
		q.Add("to", "2021-01-21T05:38:18Z")
		q.Add("aggregationType", "mean")
		q.Add("interval", "15m")
		q.Add("limit", "8")
		q.Add("offset", "0")
		res.URL.RawQuery = q.Encode()
		//res.Header.Add("content-type","application/x-www-form-urlencoded")
		if err != nil {
			//return nil, err
		}
		if req.token != "" {
			res.Header.Set("Authorization", req.token)
		}
		do, err := req.client.Do(res)
		//todo
		fmt.Println("rep---------")
		fmt.Println("StatusCode : ", do.StatusCode)
		fmt.Println("Status : ", do.Status)
		msg := do.Header.Get("messages")
		fmt.Println("msg = " + msg)
		//fmt.Println(do.)
		temp := make([]byte, 10240)
		//temp := []byte{}//这种声明定义  切片的大小为0 不能用于read读取数据
		do.Body.Read(temp)

		log.Println("temp :" + string(temp))
		do.Body.Close()
	}
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
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
		},
		"read page with negative offset": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=-1&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with negative limit": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=-10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with zero limit": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=0", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with non-integer offset": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=abc&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with non-integer limit": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=abc", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with invalid channel id": {
			url:    fmt.Sprintf("%s/channels//messages?offset=0&limit=10", ts.URL),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with invalid token": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  invalid,
			status: http.StatusForbidden,
		},
		"read page with multiple offset": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&offset=1&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with multiple limit": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=20&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		"read page with empty token": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  "",
			status: http.StatusForbidden,
		},
		"read page with default offset": {
			url:    fmt.Sprintf("%s/channels/%s/messages?limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
		},
		"read page with default limit": {
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
		},
		"read page with value": {
			url:    fmt.Sprintf("%s/channels/%s/messages?v=%f", ts.URL, chanID, v),
			token:  token,
			status: http.StatusOK,
		},
		"read page with boolean value": {
			url:    fmt.Sprintf("%s/channels/%s/messages?vb=%t", ts.URL, chanID, vb),
			token:  token,
			status: http.StatusOK,
		},
		"read page with string value": {
			url:    fmt.Sprintf("%s/channels/%s/messages?vs=%s", ts.URL, chanID, vd),
			token:  token,
			status: http.StatusOK,
		},
		"read page with data value": {
			url:    fmt.Sprintf("%s/channels/%s/messages?vd=%s", ts.URL, chanID, vd),
			token:  token,
			status: http.StatusOK,
		},
		"read page with from": {
			url:    fmt.Sprintf("%s/channels/%s/messages?from=1608651539.673909", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
		},
		"read page with to": {
			url:    fmt.Sprintf("%s/channels/%s/messages?to=1508651539.673909", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
		},
	}

	for desc, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    tc.url,
			token:  tc.token,
		}
		res, err := req.make()
		log.Println("resp:")
		log.Println(res.Header.Get("messages"))
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected %d got %d", desc, tc.status, res.StatusCode))
	}
}
