package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	schemaMap sync.Map
)

func init() {
	loadRuleFromFile()
}

func Query(name string) []byte {
	val, ok := schemaMap.Load(name)
	if !ok {
		return nil;
	}

	schema := val.(*schema)
	// query只允许配置一条语句
	grammar := schema.Grammars[0]

	client := &http.Client{Timeout: 5 * time.Second}
	//fmt.Println("grammar:" + grammar)
	//"http://116.62.210.212:4001/api"
	request, _ := http.NewRequest("POST", "http://localhost:4001/api", bytes.NewBuffer([]byte(grammar)))
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

type schema struct {
	Name      string    `json:"name"`
	Grammars  []string  `json:"grammars"`
}

func loadRuleFromFile() error {
	// 读取文件
	f, e := os.Open("graphql/schema.json")
	if e != nil {
		fmt.Println("failed to load schema from file:", e.Error())
		return e
	}
	defer f.Close()

	// 解析json文件中内容
	var allSchemas []schema
	e = json.NewDecoder(f).Decode(&allSchemas)
	if e != nil {
		fmt.Println("failed to parse schema file:", e.Error())
		return e
	}

	// 遍历并放在入map, 相当于put
	for i := 0; i < len(allSchemas); i++ {
		schema := &allSchemas[i]
		schemaMap.Store(schema.Name, schema)
	}
	return nil
}
