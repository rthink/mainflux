package model

import "fmt"

type TimeseriesData struct {
	Messages []TimeseriesDataArray
}

type TimeseriesDataArray struct {
	Publisher string
	Name      string
	Subtopic  string
	Unit      string
	Total     int
	Data      []TimeseriesDataPoint
}
type TimeseriesDataPoint struct {
	Value float64
	Time  string
}
func (t TimeseriesData) String() {
	for i := range t.Messages {
		t.Messages[i].String()
	}
}
func (t TimeseriesDataArray) String() {
	fmt.Println(t.Publisher + " " + t.Name + " " + t.Subtopic + " " + t.Unit + " ", t.Total)
	for i := range t.Data {
		t.Data[i].String()
	}
}
func (t TimeseriesDataPoint) String() {
	println(t.Value, " " + t.Time)
}


func (t *TimeseriesData) SetMessages(messages []TimeseriesDataArray) {
	t.Messages = messages
}

func (t *TimeseriesDataArray) SetPublisher(publisher string) {
	t.Publisher = publisher
}


func (t *TimeseriesDataArray) SetName(name string) {
	t.Name = name
}

func (t *TimeseriesDataArray) SetSubtopic(subtopic string) {
	t.Subtopic = subtopic
}


func (t *TimeseriesDataArray) SetTotal(total int) {
	t.Total = total
}

func (t *TimeseriesDataArray) SetUnit(unit string) {
	t.Unit = unit
}


func (t *TimeseriesDataArray) SetData(data []TimeseriesDataPoint) {
	t.Data = data
}


func (t *TimeseriesDataPoint) SetValue(value float64) {
	t.Value = value
}


func (t *TimeseriesDataPoint) SetTime(time string) {
	t.Time = time
}
