package asset

import (
	"encoding/json"
	"fmt"
	"github.com/mainflux/mainflux/graphql"
)

var (
	edgeDevices        []EdgeDevice
	// key:id
	idEdgeDeviceMap    map[string]EdgeDevice = make(map[string]EdgeDevice)
)


type edData struct {
	EdgeDevices   []EdgeDevice   `json:"edgeDevice"`
}

type EdgeDevice struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Read      []Measurement   `json:"read"`
}

type Measurement struct {
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	RangeLowerLimit    string   `json:"rangeLowerLimit"`
	RangeUpperLimit    string   `json:"rangeUpperLimit"`
	AlarmLowerLimit    string   `json:"alarmLowerLimit"`
	AlarmUpperLimit    string   `json:"alarmUpperLimit"`
	CriticalLowerLimit string   `json:"criticalLowerLimit"`
	CriticalUpperLimit string   `json:"criticalUpperLimit"`
	UnitMeasure        string   `json:"unitMeasure"`
	NameMeasure        string   `json:"nameMeasure"`

}

func initEdgeDevice() {
	result := graphql.Query("queryEdgeDevice")
	var edData edData
	e := json.Unmarshal(result, &edData)
	if e != nil {

	}
	edgeDevices = edData.EdgeDevices
	// TODO
	measurement := Measurement{"111", "test", "100", "100", "100", "100", "100", "100", "cm", "cm"}
	measurements := make([]Measurement, 1)
	measurements[0] = measurement
	edgeDevice := EdgeDevice{"123_dev", "test", measurements}
	edgeDevices := make([]EdgeDevice, 1)
	edgeDevices[0] = edgeDevice

	fmt.Println("edgeDevices len:", len(edgeDevices))
	for i := 0; i < len(edgeDevices); i++  {
		idEdgeDeviceMap[edgeDevices[i].Id] = edgeDevices[i]
	}
}


func GetEdgeDeviceById(id string) *EdgeDevice {
	if _, ok := idEdgeDeviceMap[id]; ok {
		edgeDevice := idEdgeDeviceMap[id]
		return &edgeDevice;
	}
	return nil
}