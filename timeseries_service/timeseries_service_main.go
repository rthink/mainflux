package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mainflux/mainflux/timeseries_service/service"
	"net/http"
)

func main() {
	router := gin.Default()

	router.POST("/timeseries/pumpRunningSeconds/:publishers", func(context *gin.Context) {
		publishers := context.Param("publishers")
		sensorNames := context.PostForm("sensorNames")
		startTime := context.PostForm("startTime")
		endTime := context.PostForm("endTime")
		timeseriesData := service.PumpRunningSeconds(publishers, sensorNames, startTime, endTime)
		jsonstr, _ := json.MarshalIndent(timeseriesData,"","	")
		println(string(jsonstr))
		context.String(http.StatusOK, string(jsonstr))
	})
	router.GET("/timeseries/last/:publishers", func(context *gin.Context) {
		publishers := context.Param("publishers")
		sensorName := context.Query("sensorName")
		timeseriesData := service.GetLastMeasurement(publishers, sensorName)
		jsonstr, _ := json.MarshalIndent(timeseriesData,"","	")
		println(string(jsonstr))
		context.String(http.StatusOK, string(jsonstr))
	})
	router.GET("/timeseries/all/:publisher", func(context *gin.Context) {
		publishers := context.Param("publisher")
		sensorName := context.Query("sensorName")
		startTime := context.Query("startTime")
		endTime := context.Query("endTime")
		aggregationType := context.Query("aggregationType")
		interval := context.Query("interval")
		timeseriesData := service.GetTimeseriesByPublisher(publishers, sensorName, startTime, endTime, aggregationType,interval)
		jsonstr, _ := json.MarshalIndent(timeseriesData,"","	")
		println(string(jsonstr))
		context.String(http.StatusOK, string(jsonstr))
	})
	router.Run(":8905")
}
