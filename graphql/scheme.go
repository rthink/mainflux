package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mainflux/mainflux/logger"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	schemaMap sync.Map
)

func init() {
	loadRuleFromFile()
}

type graphqlConfig struct {
	url        string
	timeout    int
	log        logger.Logger
}

var cfg graphqlConfig

func InitGraphql(url, timeout string, log logger.Logger) {
	cfg.url = url
	val, _ := strconv.Atoi(timeout)
	cfg.timeout = val
	cfg.log = log
	// 5分钟执行一次
	//c := time.Tick(5 * 60 * time.Second)
	//for {
	//	<- c
	//	initUser();
	//	initEdgeDevice()
	//}
}

func Query(name string) []byte {
	val, ok := schemaMap.Load(name)
	if !ok {
		return nil;
	}

	schema := val.(*schema)
	// query只允许配置一条语句
	grammar := schema.Grammar

	client := &http.Client{Timeout: time.Duration(cfg.timeout) * time.Second}
	//fmt.Println("grammar:" + grammar)
	request, _ := http.NewRequest("POST", cfg.url, bytes.NewBuffer([]byte(grammar)))
	request.Header.Add("tenant", "rthink");
	request.Header.Add("Content-Type","application/json;charset=utf-8")
	request.Header.Add("Authorization","Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDQwODA0NjAsImlhdCI6MTYwNDA0NDQ2MCwiaXNzIjoibWFpbmZsdXguYXV0aG4iLCJzdWIiOiJhZG1pbkBiZW5neXVuLmlvIiwidHlwZSI6MH0.tCG7Czxx7XImEZJasjS3xAyu17yz-uLC194IJQDmO7w")
	resp, e := client.Do(request)
	if e != nil {
		return nil
	}
	defer resp.Body.Close()

	// 返回格式 {"data":{}}
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	// 返回data的value
	var bodyMap map[string]interface{}
	json.Unmarshal(body, &bodyMap)
	if _, ok := bodyMap["data"]; ok  {
		data, _ := json.Marshal(bodyMap["data"])
		return data
	}
	return nil
}

func Mutation(name string, paramsMap map[string]interface{}) []byte {
	val, ok := schemaMap.Load(name)
	if !ok {
		return nil;
	}

	schema := val.(*schema)
	// query只允许配置一条语句
	grammar := schema.Grammar
	buf := new(bytes.Buffer)
	tmpl, _ := template.New("mutation").Parse(grammar)
	tmpl.Execute(buf, paramsMap)

	client := &http.Client{Timeout: time.Duration(cfg.timeout) * time.Second}
	//fmt.Println("buf:" + buf)
	request, _ := http.NewRequest("POST", cfg.url, bytes.NewBuffer(buf.Bytes()))
	request.Header.Add("tenant", "rthink");
	request.Header.Add("Content-Type","application/json;charset=utf-8")
	request.Header.Add("Authorization","Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDQwODA0NjAsImlhdCI6MTYwNDA0NDQ2MCwiaXNzIjoibWFpbmZsdXguYXV0aG4iLCJzdWIiOiJhZG1pbkBiZW5neXVuLmlvIiwidHlwZSI6MH0.tCG7Czxx7XImEZJasjS3xAyu17yz-uLC194IJQDmO7w")
	resp, e := client.Do(request)
	if e != nil {
		return nil
	}
	defer resp.Body.Close()

	// 返回格式 {"data":{}}
	body, _ := ioutil.ReadAll(resp.Body)
	var bodyMap map[string]interface{}
	json.Unmarshal(body, &bodyMap)
	if _, ok := bodyMap["data"]; ok  {
		data, _ := json.Marshal(bodyMap["data"])
		return data
	}
	return nil
}

type schema struct {
	Name      string    `json:"name"`
	Grammar   string    `json:"grammar"`
}

func loadRuleFromFile() error {
	// 读取文件
	f, e := os.Open("graphql/schema.json")
	if e != nil {
		cfg.log.Error(fmt.Sprintf("failed to load schema from file:%s", e.Error()))
		return e
	}
	defer f.Close()

	// 解析json文件中内容
	var allSchemas []schema
	e = json.NewDecoder(f).Decode(&allSchemas)
	if e != nil {
		cfg.log.Error(fmt.Sprintf("failed to parse schema file:%s", e.Error()))
		return e
	}

	// 遍历并放在入map, 相当于put
	for i := 0; i < len(allSchemas); i++ {
		schema := &allSchemas[i]
		schemaMap.Store(schema.Name, schema)
	}
	return nil
}


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
type EventType string
const (
	AlaramEvent               EventType = "alaramEvent"
)

