package rule

import (
	"encoding/json"
	"fmt"
	"github.com/mainflux/mainflux/graphql"
	"os"
	"sync"
	"time"
)

var (
	ruleMap sync.Map
)

func init() {
	loadRuleFromFile()
}

type rule struct {
	Name      	string                			`json:"name"`
	Title       string							`json:"title"`
	TableName 	graphql.TableName	  			`json:"tableName"`
	Contents  	[]Content    		  			`json:"contents"`
	Notice    	map[string]Notice     `json:"notice"`
}

type Content struct {
	Type         string        `json:"type"`
	Title        string		   `json:"title"`
	Expr         string        `json:"expr"`
	Description  string        `json:"description"`
}

type Notice struct {
	Url          string           `json:"url"`
	Template     string           `json:"template"`
	Timeout      time.Duration    `json:"timeout"`
}


func getRule(name string) *rule {
	val, ok := ruleMap.Load(name)
	if ok {
		return val.(*rule)
	}
	return nil
}

func loadRuleFromFile() error {
	// 读取文件
	f, e := os.Open("alert/rule/rule.json")
	if e != nil {
		fmt.Println("failed to load rules from file:", e.Error())
		return e
	}
	defer f.Close()

	// 解析json文件中内容
	var allRules []rule
	e = json.NewDecoder(f).Decode(&allRules)
	if e != nil {
		fmt.Println("failed to parse rule file:", e.Error())
		return e
	}

	// 遍历并放在入map
	for i := 0; i < len(allRules); i++ {
		rule := &allRules[i]
		ruleMap.Store(rule.Name, rule)
	}

	return nil
}