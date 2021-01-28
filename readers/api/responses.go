// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"strconv"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/readers"
)

var _ mainflux.Response = (*pageRes)(nil)

type pageRes struct {
	Total    uint64            `json:"total"`
	Offset   uint64            `json:"offset"`
	Limit    uint64            `json:"limit"`
	Messages []readers.Message `json:"messages,omitempty"`
}
type res struct {
	Total    uint64            `json:"total"`
	Messages []readers.Message `json:"messages,omitempty"`
}

func (res pageRes) Headers() map[string]string {
	var ret map[string]string = make(map[string]string)
	ret["limit"] = strconv.FormatUint(res.Limit, 10)
	ret["offset"] = strconv.FormatUint(res.Offset, 10)
	ret["total"] = strconv.FormatUint(res.Total, 10)

	//jsonStr, err := json.MarshalIndent(res.Messages, "", "	")
	//if err != nil {
	//	ret["messages"] = "error"
	//} else {
	//	ret["messages"] = string(jsonStr)
	//}
	return ret
}

func (res pageRes) Code() int {
	return http.StatusOK
}

func (res pageRes) Empty() bool {
	return false
}

type errorRes struct {
	Err string `json:"error"`
}

func (res res) Headers() map[string]string {
	var ret map[string]string = make(map[string]string)
	ret["total"] = strconv.FormatUint(res.Total, 10)

	//jsonStr, err := json.MarshalIndent(res.Messages, "", "	")
	//if err != nil {
	//	ret["messages"] = "error"
	//} else {
	//	ret["messages"] = string(jsonStr)
	//}
	return ret
}

func (res res) Code() int {
	return http.StatusOK
}

func (res res) Empty() bool {
	return false
}
