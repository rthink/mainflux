package graphql


// 报警状态
type EventStatusType string
const (
	Active          EventStatusType = "Active"
	Cancelled       EventStatusType = "Cancelled"
	Stale           EventStatusType = "Stale"
	Resolved        EventStatusType = "Resolved"
	Ignored         EventStatusType = "Ignored"
)

// 报警类型
type AlarmType string
const (
	BelowAlarmLowerLimit               AlarmType = "BelowAlarmLowerLimit"
	AboveAlarmUpperLimit               AlarmType = "AboveAlarmUpperLimit"
	BelowCriticalLowerLimit            AlarmType = "BelowCriticalLowerLimit"
	AboveCriticalUpperLimit            AlarmType = "AboveCriticalUpperLimit"
	AboveRiseRateUpperLimit            AlarmType = "AboveRiseRateUpperLimit"
	ScheduledTransmissionOvertime      AlarmType = "ScheduledTransmissionOvertime"
)

// 报警类型
type TableName string
const (
	AlaramEvent               TableName = "alaramEvent"
)

type DateTime struct {
	year      int
	month     int
	day       int
	hour      int
	minute    int
	second    int
}

func createDateTime(dateTime string) {

}