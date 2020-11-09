package event

import (
	"github.com/mainflux/mainflux/graphql"
)

var (
	createAlarmEventForAsset string = "CreateAlarmEventForAsset"
	addEventUpdate           string = "AddEventUpdate"
)

type alarmEvent struct {
	ManagedUserID    string
	AssetID          string
	MeasurementID    string
	Description      string
	Memo             string
	AlarmType        graphql.AlarmType
	TriggerTimeRange []string
}

func AddAlarmEvent(uploadTime, assetID, measurementID, description, memo string, alarmType graphql.AlarmType) {
	paramsMap := make(map[string]interface{})
	paramsMap["assetID"] = assetID
	paramsMap["measurementID"] = measurementID
	paramsMap["description"] = description
	paramsMap["memo"] = memo
	paramsMap["alarmType"] = alarmType
	paramsMap["triggerTimeRange"] = []string{uploadTime, uploadTime}
	graphql.Mutation(createAlarmEventForAsset, paramsMap)
}

func UpdateAlarmEvent(id, uploadTime string) {
	paramsMap := make(map[string]interface{})
	paramsMap["id"] = id
	// TODO
	paramsMap["triggerTimeRange"]= []string{uploadTime, uploadTime}
	graphql.Mutation(addEventUpdate, paramsMap)
}

