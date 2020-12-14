package Utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
	uuid "github.com/satori/go.uuid"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	testDB = "mainflux" /*"test"*/
)

var (
	client influxdata.Client = ConnInflux()
)
var errReadMessages = errors.New("failed to read messages from influxdb database")

func PumpRunningSeconds(publishers []uuid.UUID, names []string, startTime string, endTime string) ([]readers.MessagesPage, error) {
	//List<UUID> publishers,List<String> names, String startTime, String endTime
	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	buffer.WriteString("INTEGRAL(value) as value ")
	buffer.WriteString("FROM ")
	buffer.WriteString("\"messages\" ")
	buffer.WriteString("WHERE ")
	buffer.WriteString("time >= '")
	buffer.WriteString(startTime)
	buffer.WriteString("' ")

	if endTime != "" {
		buffer.WriteString("AND time <= '")
		buffer.WriteString(endTime)
		buffer.WriteString("' ")
	}
	if len(publishers) > 0 {
		buffer.WriteString(" and ( ")
		for i := 0; i < len(publishers); i++ {
			s := publishers[i]
			buffer.WriteString("\"publisher\" = '")
			buffer.WriteString(s.String())
			buffer.WriteString("' ")
			if i != len(publishers)-1 {
				buffer.WriteString(" or ")
			}
		}
		buffer.WriteString(" ) ")
	}

	if len(names) > 0 {
		buffer.WriteString(" and ( ")
		for i := 0; i < len(names); i++ {
			s := names[i]
			buffer.WriteString("\"name\" = '")
			buffer.WriteString(s)
			buffer.WriteString("' ")
			if i != len(names)-1 {
				buffer.WriteString(" or ")
			}
		}
		buffer.WriteString(" ) ")
	}
	buffer.WriteString("group by \"publisher\" , \"name\"")
	log.Println(buffer.String())

	ret, err := QueryDB(client, buffer.String())
	if err != nil {
		log.Println(err)
		return []readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	fmt.Println(ret)
	rett := []readers.MessagesPage{}

	for i := range ret[0].Series {
		result := ret[0].Series[i]
		rets := []senml.Message{}
		for _, v := range result.Values {
			temp := parseMessage(result.Columns, v)
			parseMessageTags(result.Tags, &temp)
			rets = append(rets, temp)
		}
		rett = append(rett, readers.MessagesPage{0, 0, 0, rets})
	}

	return rett, nil
}

func GetLastMeasurement(publishers []uuid.UUID, sensorName string) ([]readers.MessagesPage, error) {
	//List<UUID> publishers, String sensorName
	//SELECT last("value") as "value", "unit" FROM messages WHERE publisher =~ /(2a9dfc0a-661c-442b-895e-7d61a56cc57c|c8100287-c493-4177-9f80-48c803ef63c8)/ AND "name" = 'pump_1' GROUP BY publisher,"name";
	var buffer bytes.Buffer
	buffer.WriteString("SELECT last(\"value\") as \"value\", \"unit\"")
	buffer.WriteString(" FROM ")
	buffer.WriteString(" messages where publisher =~ /(")
	for i := range publishers {
		buffer.WriteString(publishers[i].String())
		if i != len(publishers)-1 {
			buffer.WriteString("|")
		}
	}
	buffer.WriteString(")/ AND \"name\" = '")
	buffer.WriteString(sensorName)
	buffer.WriteString("' GROUP BY publisher")
	log.Println("cmd = " + buffer.String())

	ret, err := QueryDB(client, buffer.String())
	fmt.Println(ret)
	if err != nil {
		log.Println(err)
		return []readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	rett := []readers.MessagesPage{}
	for i := range ret[0].Series {
		result := ret[0].Series[i]
		rets := []senml.Message{}
		for _, v := range result.Values {
			temp := parseMessage(result.Columns, v)
			parseMessageTags(result.Tags, &temp)
			temp.Name = sensorName
			rets = append(rets, temp)
		}
		rett = append(rett, readers.MessagesPage{0, 0, 0, rets})
	}
	return rett, nil
}

func GetTimeseriesByPublisher(publisher uuid.UUID, sensorName string, startTime string, endTime string, aggregationType string, interval string) (readers.MessagesPage, error) {

	var buffer bytes.Buffer
	buffer.WriteString("SELECT ")
	if interval != "" && aggregationType != "" {
		buffer.WriteString(aggregationType + "(\"value\") as value, ")
		buffer.WriteString(" last(\"unit\") as unit ")
	} else {
		buffer.WriteString("\"value\", \"unit\" ")
	}

	buffer.WriteString("FROM ")
	buffer.WriteString("\"messages\" ")
	buffer.WriteString("WHERE ")
	buffer.WriteString("time > '")
	buffer.WriteString(startTime)
	buffer.WriteString("' ")

	if endTime != "" {
		buffer.WriteString("AND ")
		buffer.WriteString("time <= '")
		buffer.WriteString(endTime)
		buffer.WriteString("' ")
	}
	buffer.WriteString("AND ")
	buffer.WriteString("\"publisher\" = '")
	buffer.WriteString(publisher.String())
	buffer.WriteString("' ")

	if sensorName != "" {
		buffer.WriteString("AND ")
		buffer.WriteString("\"name\" = '")
		buffer.WriteString(sensorName)
		buffer.WriteString("' ")
	}
	buffer.WriteString("GROUP BY ")
	if interval != "" && aggregationType != "" {
		buffer.WriteString(" time(" + interval + "), ")
	}
	buffer.WriteString("\"name\"")
	log.Println(buffer.String())

	//reader := reader.New(client, testDB)
	ret, err := QueryDB(client, buffer.String())
	fmt.Println(ret)
	if err != nil {
		log.Println(err)
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	result := ret[0].Series[0]
	rets := []senml.Message{}
	for _, v := range result.Values {
		temp := parseMessage(result.Columns, v)
		parseMessageTags(result.Tags, &temp)
		temp.Publisher = publisher.String()
		rets = append(rets, temp)
	}
	return readers.MessagesPage{
		Total:    0,
		Offset:   0,
		Limit:    0,
		Messages: rets,
	}, nil
}

func fmtCondition(chanID string, query map[string]string) string {
	condition := fmt.Sprintf(`channel='%s'`, chanID)
	for name, value := range query {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher":
			condition = fmt.Sprintf(`%s AND %s='%s'`, condition, name,
				strings.Replace(value, "'", "\\'", -1))
		case
			"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND "%s"='%s'`, condition, name,
				strings.Replace(value, "\"", "\\\"", -1))
		}
	}
	return condition
}

// ParseMessage and parseValues are util methods. Since InfluxDB client returns
// results in form of rows and columns, this obscure message conversion is needed
// to return actual []broker.Message from the query result.
func parseValues(value interface{}, name string, msg *senml.Message) {
	if name == "sum" && value != nil {
		if valSum, ok := value.(json.Number); ok {
			sum, err := valSum.Float64()
			if err != nil {
				return
			}

			msg.Sum = &sum
		}
		return
	}

	if strings.HasSuffix(strings.ToLower(name), "value") {
		switch value.(type) {
		case bool:
			v := value.(bool)
			msg.BoolValue = &v
		case json.Number:
			num, err := value.(json.Number).Float64()
			if err != nil {
				return
			}
			msg.Value = &num
		case string:
			if strings.HasPrefix(name, "string") {
				v := value.(string)
				msg.StringValue = &v
				return
			}

			if strings.HasPrefix(name, "data") {
				v := value.(string)
				msg.DataValue = &v
			}
		}
	}
}
func parseMessageTags(tags map[string]string, message *senml.Message) {
	message.Publisher = tags["publisher"]
	message.Name = tags["name"]
}
func parseMessage(names []string, fields []interface{}) senml.Message {
	m := senml.Message{}
	v := reflect.ValueOf(&m).Elem()
	for i, name := range names {
		parseValues(fields[i], name, &m)
		msgField := v.FieldByName(strings.Title(name))
		if !msgField.IsValid() {
			continue
		}

		f := msgField.Interface()
		switch f.(type) {
		case string:
			if s, ok := fields[i].(string); ok {
				msgField.SetString(s)
			}
		case float64:
			if name == "time" {
				t, err := time.Parse(time.RFC3339Nano, fields[i].(string))
				if err != nil {
					continue
				}

				v := float64(t.UnixNano()) / float64(1e9)
				msgField.SetFloat(v)
				continue
			}

			val, _ := strconv.ParseFloat(fields[i].(string), 64)
			msgField.SetFloat(val)
		}
	}

	return m
}
