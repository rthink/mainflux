// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/readers"
)

func listMessagesEndpoint(svc readers.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(listMessagesReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ReadAll(req.chanID, req.offset, req.limit, req.query)
		if err != nil {
			return nil, err
		}

		return pageRes{
			Total:    page.Total,
			Offset:   page.Offset,
			Limit:    page.Limit,
			Messages: page.Messages,
		}, nil
	}
}

//返回处理lastMeasurement请求的handler
func lastMeasurementEndpoint(svc readers.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(lastMeasurementReq)

		/*if err := req.validate(); err != nil {
			return nil, err
		}*/
		//调用服务
		page, err := svc.GetLastMeasurement(req.chanIDs, req.query)
		if err != nil {
			return nil, err
		}

		return pageRes{
			Total:    page.Total,
			Offset:   page.Offset,
			Limit:    page.Limit,
			Messages: page.Messages,
		}, nil
	}
}

//返回处理pumpRunningSeconds请求的handler
func pumpRunningSecondsEndpoint(svc readers.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(pumpRunningReq)
		//调用服务
		page, err := svc.PumpRunningSeconds(req.chanIDs, req.query)
		if err != nil {
			return nil, err
		}

		return pageRes{
			Total:    page.Total,
			Offset:   page.Offset,
			Limit:    page.Limit,
			Messages: page.Messages,
		}, nil
	}
}

//返回处理getMessageByPublisher请求的handler
func getMessageByPublisherEndpoint(svc readers.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(getMessageReq)
		//调用服务
		page, err := svc.GetMessageByPublisher(req.chanID, req.offset, req.limit, req.aggregationType, req.interval, req.query)
		if err != nil {
			return nil, err
		}

		return pageRes{
			Total:    page.Total,
			Offset:   page.Offset,
			Limit:    page.Limit,
			Messages: page.Messages,
		}, nil
	}
}
