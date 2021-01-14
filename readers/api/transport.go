// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	contentType = "application/json"
	defLimit    = 10
	defOffset   = 0
	format      = "format"
	defFormat   = "messages"
)

var (
	errInvalidRequest     = errors.New("received invalid request")
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	auth                  mainflux.ThingsServiceClient
	queryFields           = []string{"format", "subtopic", "publisher", "protocol", "name", "v", "vs", "vb", "vd", "from", "to"}
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc readers.MessageRepository, tc mainflux.ThingsServiceClient, svcName string) http.Handler {
	auth = tc

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
	mux := bone.New()
	mux.Get("/channels/:chanID/messages", kithttp.NewServer(
		listMessagesEndpoint(svc),
		decodeList,
		encodeResponse,
		opts...,
	))

	//向路由器注册路径和处理器handler
	mux.Get("/messages/last/:chanIDs", kithttp.NewServer(
		lastMeasurementEndpoint(svc), //注册自己定义的处理器handler
		decodeLast,                   //提供解析request的方法，该方法解析原生态request 生成另一种request, 生成的request会被lastMeasurementEndpoint函数生成的handler使用到
		encodeResponse,               //将lastMeasurementEndpoint函数生成的handler所返回的response解析成http认识的response
		opts...,
	))
	mux.Post("/messages/pumpRunningSeconds/:chanIDs", kithttp.NewServer(
		pumpRunningSecondsEndpoint(svc),
		decodepumpRunning,
		encodeResponse,
		opts...,
	))
	mux.Get("/messages/list/:chanID", kithttp.NewServer(
		getMessageByPublisherEndpoint(svc),
		decodeMessageByChannal,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", mainflux.Version(svcName))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeMessageByChannal(_ context.Context, r *http.Request) (interface{}, error) {
	chanID := bone.GetValue(r, "chanID")
	if chanID == "" {
		return nil, errInvalidRequest
	}

	if err := authorize(r, chanID); err != nil {
		return nil, err
	}

	offset, err := getQuery(r, "offset", defOffset)
	if err != nil {
		return nil, err
	}

	limit, err := getQuery(r, "limit", defLimit)
	if err != nil {
		return nil, err
	}

	query := map[string]string{}
	if value := bone.GetQuery(r, "aggregationType"); len(value) == 1 {
		query["aggregationType"] = value[0]
	}
	if value := bone.GetQuery(r, "interval"); len(value) == 1 {
		query["interval"] = value[0]
	}

	for _, name := range queryFields {
		if value := bone.GetQuery(r, name); len(value) == 1 {
			query[name] = value[0]
		}
	}
	if query[format] == "" {
		query[format] = defFormat
	}

	req := listMessagesReq{
		chanID: chanID,
		offset: offset,
		limit:  limit,
		query:  query,
	}

	return req, nil
}

func decodepumpRunning(_ context.Context, r *http.Request) (interface{}, error) {
	chanIDs := bone.GetValue(r, "chanIDs")
	if chanIDs == "" {
		return nil, errInvalidRequest
	}
	chanIDList := strings.Split(chanIDs, ",")
	for i := range chanIDList {
		if err := authorize(r, chanIDList[i]); err != nil {
			return nil, err
		}
	}

	query := map[string]string{}
	for _, name := range queryFields {
		if value := bone.GetQuery(r, name); len(value) == 1 {
			query[name] = value[0]
		}
	}
	if query[format] == "" {
		query[format] = defFormat
	}

	req := pumpRunningReq{
		chanIDs: chanIDList,
		query:   query,
	}

	return req, nil
}
func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	chanID := bone.GetValue(r, "chanID")
	if chanID == "" {
		return nil, errInvalidRequest
	}

	if err := authorize(r, chanID); err != nil {
		return nil, err
	}

	offset, err := getQuery(r, "offset", defOffset)
	if err != nil {
		return nil, err
	}

	limit, err := getQuery(r, "limit", defLimit)
	if err != nil {
		return nil, err
	}

	query := map[string]string{}
	for _, name := range queryFields {
		if value := bone.GetQuery(r, name); len(value) == 1 {
			query[name] = value[0]
		}
	}
	if query[format] == "" {
		query[format] = defFormat
	}

	req := listMessagesReq{
		chanID: chanID,
		offset: offset,
		limit:  limit,
		query:  query,
	}

	return req, nil
}

func decodeLast(_ context.Context, r *http.Request) (interface{}, error) {
	chanIDs := bone.GetValue(r, "chanIDs")
	if chanIDs == "" {
		return nil, errInvalidRequest
	}
	chanIDList := strings.Split(chanIDs, ",")
	for i := range chanIDList {
		if err := authorize(r, chanIDList[i]); err != nil {
			return nil, err
		}
	}

	query := map[string]string{}
	for _, name := range queryFields {
		if value := bone.GetQuery(r, name); len(value) == 1 {
			query[name] = value[0]
		}
	}
	if query[format] == "" {
		query[format] = defFormat
	}

	req := lastMeasurementReq{
		chanIDs: chanIDList,
		query:   query,
	}

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

//func encodeLastResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
//	w.Header().Set("Content-Type", contentType)
//
//	if ar, ok := response.(mainflux.Response); ok {
//		for k, v := range ar.Headers() {
//			w.Header().Set(k, v)
//		}
//
//		w.WriteHeader(ar.Code())
//
//		if ar.Empty() {
//			return nil
//		}
//	}
//
//	return json.NewEncoder(w).Encode(response)
//}
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case errors.Contains(err, nil):
	case errors.Contains(err, errInvalidRequest):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errUnauthorizedAccess):
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	errorVal, ok := err.(errors.Error)
	if ok {
		w.Header().Set("Content-Type", contentType)
		if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func authorize(r *http.Request, chanID string) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return errUnauthorizedAccess
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := auth.CanAccessByKey(ctx, &mainflux.AccessByKeyReq{Token: token, ChanID: chanID})
	if err != nil {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.PermissionDenied {
			return errUnauthorizedAccess
		}
		return err
	}

	return nil
}

func getQuery(req *http.Request, name string, fallback uint64) (uint64, error) {
	vals := bone.GetQuery(req, name)
	if len(vals) == 0 {
		return fallback, nil
	}

	if len(vals) > 1 {
		return 0, errInvalidRequest
	}

	val, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		return 0, errInvalidRequest
	}

	return uint64(val), nil
}
