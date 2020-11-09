package alert

import (
	"github.com/mainflux/mainflux/graphql"
	"sync"
	"time"
)

type Rule struct {
	Id       		string
	Name  			string
	Title      		string
	EventType 		graphql.EventType
	Contents 		[]Content
	Notice 			map[string]Notice
}

type Content struct {
	Type         graphql.AlarmType      `json:"type"`
	Title        string		    		`json:"title"`
	Expr         string        			`json:"expr"`
	Description  string        			`json:"description"`
}

type Notice struct {
	Url          string           `json:"url"`
	Template     string           `json:"template"`
	Timeout      time.Duration    `json:"timeout"`
}


type RuleRepository interface {
	// 查询所有的规则
	ListAll() ([]Rule, error)
}


var (
	ruleMap sync.Map
	ruleRepo RuleRepository
)

func getRule(name string) *Rule {
	val, ok := ruleMap.Load(name)
	if ok {
		return val.(*Rule)
	}
	return nil
}

func loadRuleFromDb() error {
	allRules, _ := ruleRepo.ListAll()
	// 遍历并放入map
	for i := 0; i < len(allRules); i++ {
		rule := &allRules[i]
		ruleMap.Store(rule.Name, rule)
	}

	return nil
}

