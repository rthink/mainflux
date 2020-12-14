package service

import (
	"github.com/mainflux/mainflux/timeseries_service/model"
	Utils "github.com/mainflux/mainflux/timeseries_service/utils"
	uuid "github.com/satori/go.uuid"
	"log"
	"strings"
	"time"
)

func PumpRunningSeconds(publishers string, sensorNames string, startTime string, endTime string) model.TimeseriesData {
	publisherList := strings.Split(publishers, ",")
	sensorNameList := strings.Split(sensorNames, ",")
	publisherUUIDList := []uuid.UUID{}
	for i := range publisherList {
		UUID, _ := uuid.FromString(publisherList[i])
		publisherUUIDList = append(publisherUUIDList, UUID)
	}
	for i := 0; i < len(publisherUUIDList); i++ {
		println(publisherUUIDList[i].String())
	}
	for i := 0; i < len(sensorNameList); i++ {
		println(sensorNameList[i])
	}
	println(startTime)
	println(endTime)
	measurement, err := Utils.PumpRunningSeconds(publisherUUIDList, sensorNameList, startTime, endTime)
	if err != nil {
		log.Println(err)
		return model.TimeseriesData{}
	}
	var timeseriesData model.TimeseriesData
	var timeseriesDataArrays []model.TimeseriesDataArray //= []model.TimeseriesDataArray{}
	for j := range measurement {
		for i := range measurement[j].Messages {
			timeseriesDataPoint := model.TimeseriesDataPoint{}
			t := measurement[j].Messages[i].Time
			t2 := time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
			timeseriesDataPoint.SetTime(t2)
			if measurement[j].Messages[i].Value != nil {
				timeseriesDataPoint.SetValue(*(measurement[j].Messages[i].Value))
			}

			var timeseriesDataPointList []model.TimeseriesDataPoint
			timeseriesDataPointList = append(timeseriesDataPointList, timeseriesDataPoint)
			var timeseriesDataArray model.TimeseriesDataArray
			timeseriesDataArray.SetData(timeseriesDataPointList)
			timeseriesDataArray.SetName(measurement[j].Messages[i].Name)
			timeseriesDataArray.SetPublisher(measurement[j].Messages[i].Publisher)
			timeseriesDataArray.SetSubtopic(measurement[j].Messages[i].Subtopic)
			timeseriesDataArray.SetUnit(measurement[j].Messages[i].Unit)
			timeseriesDataArray.SetTotal(1)
			timeseriesDataArrays = append(timeseriesDataArrays, timeseriesDataArray)
		}
	}
	timeseriesData.SetMessages(timeseriesDataArrays)
	timeseriesData.String()
	return timeseriesData
}

func GetLastMeasurement(publishers string, sensorName string) model.TimeseriesData {
	publisherList := strings.Split(publishers, ",")
	publisherUUIDList := []uuid.UUID{}
	for i := range publisherList {
		UUID, _ := uuid.FromString(publisherList[i])
		publisherUUIDList = append(publisherUUIDList, UUID)
	}
	measurement, err := Utils.GetLastMeasurement(publisherUUIDList, sensorName)
	if err != nil {
		log.Println(err)
		return model.TimeseriesData{}
	}

	var timeseriesData model.TimeseriesData
	var timeseriesDataArrays []model.TimeseriesDataArray //= []model.TimeseriesDataArray{}
	for j := range measurement {
		for i := range measurement[j].Messages {
			timeseriesDataPoint := model.TimeseriesDataPoint{}
			t := measurement[j].Messages[i].Time
			t2 := time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
			timeseriesDataPoint.SetTime(t2)
			if measurement[j].Messages[i].Value != nil {
				timeseriesDataPoint.SetValue(*(measurement[j].Messages[i].Value))
			}

			var timeseriesDataPointList []model.TimeseriesDataPoint
			timeseriesDataPointList = append(timeseriesDataPointList, timeseriesDataPoint)
			var timeseriesDataArray model.TimeseriesDataArray
			timeseriesDataArray.SetData(timeseriesDataPointList)
			timeseriesDataArray.SetName(measurement[j].Messages[i].Name)
			timeseriesDataArray.SetPublisher(measurement[j].Messages[i].Publisher)
			timeseriesDataArray.SetSubtopic(measurement[j].Messages[i].Subtopic)
			timeseriesDataArray.SetUnit(measurement[j].Messages[i].Unit)
			timeseriesDataArray.SetTotal(1)
			timeseriesDataArrays = append(timeseriesDataArrays, timeseriesDataArray)
		}
	}
	timeseriesData.SetMessages(timeseriesDataArrays)
	timeseriesData.String()
	return timeseriesData
}

func GetTimeseriesByPublisher(publisher string, sensorName string, startTime string, endTime string, aggregationType string, interval string) model.TimeseriesData {
	publisherUUID, _ := uuid.FromString(publisher)
	measurement, err := Utils.GetTimeseriesByPublisher(publisherUUID, sensorName, startTime, endTime, aggregationType, interval)
	if err != nil {
		log.Println(err)
		return model.TimeseriesData{}
	}
	var timeseriesData model.TimeseriesData
	var timeseriesDataArrays []model.TimeseriesDataArray //= []model.TimeseriesDataArray{}
	var timeseriesDataArray model.TimeseriesDataArray
	var timeseriesDataPointList []model.TimeseriesDataPoint
	timeseriesDataArray.Publisher = publisher
	timeseriesDataArray.Name = sensorName
	if len(measurement.Messages) > 0 {
		timeseriesDataArray.Unit = measurement.Messages[0].Unit
	}
	timeseriesDataArray.Total = len(measurement.Messages)
	messages := measurement.Messages
	for i := range messages {
		timeseriesDataPoint := model.TimeseriesDataPoint{}
		t := messages[i].Time
		t2 := time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
		timeseriesDataPoint.SetTime(t2)
		if messages[i].Value != nil {
			timeseriesDataPoint.SetValue(*(messages[i].Value))
		}
		timeseriesDataPointList = append(timeseriesDataPointList, timeseriesDataPoint)
	}
	timeseriesDataArray.Data = timeseriesDataPointList
	timeseriesDataArrays = append(timeseriesDataArrays, timeseriesDataArray)
	timeseriesData.SetMessages(timeseriesDataArrays)
	timeseriesData.String()
	return timeseriesData
}