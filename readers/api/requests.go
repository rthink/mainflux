// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	chanID string
	offset uint64
	limit  uint64
	query  map[string]string
}

//定义处理lastMeasurement请求的请求体
type lastMeasurementReq struct {
	chanIDs []string
	query   map[string]string
}

//定义处理pumpRunningSeconds请求的请求体
type pumpRunningReq struct {
	chanIDs []string
	query   map[string]string
}

//定义处理getMessageByChannal请求的请求体
type getMessageReq struct {
	chanID          string
	offset          uint64
	limit           uint64
	aggregationType string
	interval        string
	query           map[string]string
}

func (req listMessagesReq) validate() error {
	if req.limit < 1 {
		return errInvalidRequest
	}

	return nil
}
