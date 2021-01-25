// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

var _ mainflux.Response = (*pageRes)(nil)

type pageRes struct {
	Total    uint64            `json:"total"`
	Offset   uint64            `json:"offset"`
	Limit    uint64            `json:"limit"`
	Messages []readers.Message `json:"messages,omitempty"`
}

func (res pageRes) Headers() map[string]string {
	var ret map[string]string = make(map[string]string)
	ret["limit"] = strconv.FormatUint(res.Limit, 10)
	ret["offset"] = strconv.FormatUint(res.Offset, 10)
	ret["total"] = strconv.FormatUint(res.Total, 10)

	str := ""
	for i := range res.Messages {
		str = str + "\n" + temp(res.Messages[i])
	}
	ret["messages"] = str
	return ret
}
func temp(msg interface{}) string {
	message, ok := msg.(senml.Message)
	if ok == false {
		//fmt.Println("temp func : error")
		return "error"
	}
	jsonstr, _ := json.MarshalIndent(message, "", "	")
	//fmt.Println("responses:message = " , message)
	fmt.Println("responses:jsonstr = " + string(jsonstr))
	return string(jsonstr)
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
