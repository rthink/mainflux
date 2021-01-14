package influxdb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers"

	influxdata "github.com/influxdata/influxdb/client/v2"
	jsont "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
)

const (
	countCol = "count_protocol"
	format   = "format"
	// Measurement for SenML messages
	defMeasurement = "messages"
)

var errReadMessages = errors.New("failed to read messages from influxdb database")

var _ readers.MessageRepository = (*influxRepository)(nil)

type influxRepository struct {
	database string
	client   influxdata.Client
}

// New returns new InfluxDB reader.
func New(client influxdata.Client, database string) readers.MessageRepository {
	return &influxRepository{
		database,
		client,
	}
}

/**
desc：获取指定chanID和name的设备的最新值
param:
	chanIDs: 要查询的chanID集合
	query: 查询条件，如name
*/
func (repo *influxRepository) GetLastMeasurement(chanIDs []string, query map[string]string) (readers.MessagesPage, error) {
	measurement, ok := query[format]
	if !ok {
		measurement = defMeasurement
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	//将chanIDs集合和query查询条件放进sql语句中
	condition := fmtConditionByChanIDs(chanIDs, query)

	cmd := fmt.Sprintf(`SELECT last("value") as "value", "unit" FROM %s WHERE %s GROUP BY publisher`, measurement, condition)
	//todo
	fmt.Println("cmd = " + cmd)

	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	var ret []readers.Message
	//查询数据库 得到结果集resp
	resp, err := repo.client.Query(q)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if resp.Error() != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, resp.Error())
	}

	if len(resp.Results) < 1 || len(resp.Results[0].Series) < 1 {
		return readers.MessagesPage{}, nil
	}
	//解析结果
	result := resp.Results[0].Series[0]
	for _, v := range result.Values {
		//解析fields字段的值
		temp := parseMessage(measurement, result.Columns, v)
		//解析tag字段的值
		parseMessageTags(result.Tags, &temp)
		ret = append(ret, temp)
	}

	return readers.MessagesPage{
		Total:    uint64(len(ret)),
		Offset:   0,
		Limit:    0,
		Messages: ret,
	}, nil
}

/**
desc：获取指定chanID和name的设备在指定时间段内泵站开启的时间
param:
	chanIDs: 要查询的chanID集合
	query: 查询条件，如name
*/
func (repo *influxRepository) PumpRunningSeconds(chanIDs []string, query map[string]string) (readers.MessagesPage, error) {
	measurement, ok := query[format]
	if !ok {
		measurement = defMeasurement
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	//将chanIDs集合和query查询条件放进sql语句中
	condition := fmtConditionByChanIDs(chanIDs, query)

	cmd := fmt.Sprintf(`SELECT INTEGRAL(value) as value FROM %s WHERE %s GROUP BY group by "publisher" , "name"`, measurement, condition)
	//todo
	fmt.Println("cmd = " + cmd)

	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	var ret []readers.Message

	resp, err := repo.client.Query(q)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if resp.Error() != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, resp.Error())
	}

	if len(resp.Results) < 1 || len(resp.Results[0].Series) < 1 {
		return readers.MessagesPage{}, nil
	}

	result := resp.Results[0].Series[0]
	for _, v := range result.Values {
		//解析fields字段的值
		temp := parseMessage(measurement, result.Columns, v)
		//解析tag字段的值
		parseMessageTags(result.Tags, &temp)
		ret = append(ret, temp)
	}

	return readers.MessagesPage{
		Total:    uint64(len(ret)),
		Offset:   0,
		Limit:    0,
		Messages: ret,
	}, nil
}

/**
desc：获取指定chanID和name的设备在指定时间段内
param:
	chanIDs: 要查询的chanID集合
	query: 查询条件，如name
	offset,limit: 分页条件
	aggregationType:sql语句中执行的函数, 如sum, max
	interval: 时间间隔
*/
func (repo *influxRepository) GetMessageByPublisher(chanID string, offset, limit uint64, aggregationType string, interval string, query map[string]string) (readers.MessagesPage, error) {
	//若aggregationType为空 说明sql中没有函数  调用ReadAll
	if aggregationType == "" {
		return repo.ReadAll(chanID, offset, limit, query)
	}
	measurement, ok := query[format]
	if !ok {
		measurement = defMeasurement
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	condition := fmtCondition(chanID, query)

	cmd := fmt.Sprintf(`SELECT %s ("value") as value, last("unit") as unit FROM %s WHERE %s GROUP BY time("%s")"name"`, aggregationType, measurement, condition, interval)
	//todo
	fmt.Println("cmd = " + cmd)
	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	var ret []readers.Message
	//查询数据库
	resp, err := repo.client.Query(q)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if resp.Error() != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, resp.Error())
	}

	if len(resp.Results) < 1 || len(resp.Results[0].Series) < 1 {
		return readers.MessagesPage{}, nil
	}
	//解析结果
	result := resp.Results[0].Series[0]
	for _, v := range result.Values {
		ret = append(ret, parseMessage(measurement, result.Columns, v))
	}

	total, err := repo.count(measurement, condition)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return readers.MessagesPage{
		Total:    total,
		Offset:   0,
		Limit:    0,
		Messages: ret,
	}, nil
}

func (repo *influxRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	measurement, ok := query[format]
	if !ok {
		measurement = defMeasurement
	}
	// Remove format filter and format the rest properly.
	delete(query, format)
	condition := fmtCondition(chanID, query)

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE %s ORDER BY time DESC LIMIT %d OFFSET %d`, measurement, condition, limit, offset)
	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	var ret []readers.Message

	resp, err := repo.client.Query(q)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if resp.Error() != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, resp.Error())
	}

	if len(resp.Results) < 1 || len(resp.Results[0].Series) < 1 {
		return readers.MessagesPage{}, nil
	}

	result := resp.Results[0].Series[0]
	for _, v := range result.Values {
		ret = append(ret, parseMessage(measurement, result.Columns, v))
	}

	total, err := repo.count(measurement, condition)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return readers.MessagesPage{
		Total:    total,
		Offset:   offset,
		Limit:    limit,
		Messages: ret,
	}, nil
}

func (repo *influxRepository) count(measurement, condition string) (uint64, error) {
	cmd := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, measurement, condition)
	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	resp, err := repo.client.Query(q)
	if err != nil {
		return 0, err
	}
	if resp.Error() != nil {
		return 0, resp.Error()
	}

	if len(resp.Results) < 1 ||
		len(resp.Results[0].Series) < 1 ||
		len(resp.Results[0].Series[0].Values) < 1 {
		return 0, nil
	}

	countIndex := 0
	for i, col := range resp.Results[0].Series[0].Columns {
		if col == countCol {
			countIndex = i
			break
		}
	}

	result := resp.Results[0].Series[0].Values[0]
	if len(result) < countIndex+1 {
		return 0, nil
	}

	count, ok := result[countIndex].(json.Number)
	if !ok {
		return 0, nil
	}

	return strconv.ParseUint(count.String(), 10, 64)
}
func fmtConditionByChanIDs(chanIDs []string, query map[string]string) string {
	condition := ""
	for i := range chanIDs {
		if condition == "" {
			condition = fmt.Sprintf(`channel='%s'`, chanIDs[i])
		} else {
			condition = fmt.Sprintf(`%s OR channel ='%s'`, condition, chanIDs[i])
		}
	}
	//condition := fmt.Sprintf(`channel='%s'`, chanID)
	for name, value := range query {
		switch name {
		case "name":
			nameList := strings.Split(value, ",")
			if len(nameList) > 0 {
				condition = fmt.Sprintf(`%s and ( `, condition)
				for i := range nameList {
					condition = fmt.Sprintf(`%s "%s"='%s'`, condition, name, value)
					if i != len(nameList)-1 {
						condition = fmt.Sprintf(`%s or `, condition)
					}
				}
				condition = fmt.Sprintf(`%s ) `, condition)
			}
		case
			"channel",
			"subtopic",
			"publisher",
			//"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND "%s"='%s'`, condition, name, value)
		case "v":
			condition = fmt.Sprintf(`%s AND value = %s`, condition, value)
		case "vb":
			condition = fmt.Sprintf(`%s AND boolValue = %s`, condition, value)
		case "vs":
			condition = fmt.Sprintf(`%s AND stringValue = '%s'`, condition, value)
		case "vd":
			condition = fmt.Sprintf(`%s AND dataValue = '%s'`, condition, value)
		case "from":
			fVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			iVal := int64(fVal * 1e9)
			condition = fmt.Sprintf(`%s AND time >= %d`, condition, iVal)
		case "to":
			fVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			iVal := int64(fVal * 1e9)
			condition = fmt.Sprintf(`%s AND time < %d`, condition, iVal)
		}
	}
	return condition
}
func fmtCondition(chanID string, query map[string]string) string {
	condition := fmt.Sprintf(`channel='%s'`, chanID)
	for name, value := range query {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND "%s"='%s'`, condition, name, value)
		case "v":
			condition = fmt.Sprintf(`%s AND value = %s`, condition, value)
		case "vb":
			condition = fmt.Sprintf(`%s AND boolValue = %s`, condition, value)
		case "vs":
			condition = fmt.Sprintf(`%s AND stringValue = '%s'`, condition, value)
		case "vd":
			condition = fmt.Sprintf(`%s AND dataValue = '%s'`, condition, value)
		case "from":
			fVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			iVal := int64(fVal * 1e9)
			condition = fmt.Sprintf(`%s AND time >= %d`, condition, iVal)
		case "to":
			fVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			iVal := int64(fVal * 1e9)
			condition = fmt.Sprintf(`%s AND time < %d`, condition, iVal)
		}
	}
	return condition
}

func parseMessageTags(tags map[string]string, message *interface{}) {
	//*senml.Message
	s := (*message).(senml.Message)
	s.Publisher = tags["publisher"]
	s.Name = tags["name"]
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

func parseMessage(measurement string, names []string, fields []interface{}) interface{} {
	switch measurement {
	case defMeasurement:
		return parseSenml(names, fields)
	default:
		return parseJSON(names, fields)
	}
}

func parseSenml(names []string, fields []interface{}) interface{} {
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

func parseJSON(names []string, fields []interface{}) interface{} {
	ret := make(map[string]interface{})
	for i, n := range names {
		ret[n] = fields[i]
	}

	return jsont.ParseFlat(ret)
}
