package event

import "github.com/mainflux/mainflux/graphql"

type AlarmEvent struct {
	Id               string
	TimeStamp        graphql.DateTime
	Description      string
	TargetAsset      string
	Status           graphql.EventStatusType
	ResolvedBy       string
	ResolveTimeStamp graphql.DateTime
	Memo             string
	AlarmType        graphql.AlarmType
	TriggeredBy      string
	TriggerTimeRange []graphql.DateTime
}

func AddAlarmEvent(params map[string]interface{}) {

}

func updateAlarmEvent(params map[string]interface{}) {

}

