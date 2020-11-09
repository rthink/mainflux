package alert

import (
	"bytes"
	"fmt"
	"github.com/mainflux/mainflux/alert/expr"
	"github.com/mainflux/mainflux/graphql"
	"github.com/mainflux/mainflux/graphql/asset"
	"github.com/mainflux/mainflux/logger"
	"strconv"
	"strings"
	"text/template"
)

var (
	waterLevelExcept      = "WATER_LEVEL_EXCEPT"
	reportTimeOut         = "REPORT_TIME_OUT"
	waterLevelUpExcept    = "WATER_LEVEL_UP_EXCEPT"
)

type AlertContent struct {
	Title       string
	EventType   graphql.EventType
	Content     Content
	Notice      map[string]Notice
	Threshold   string
	CurrVal     string
	Description string
	NeedAlert   bool
}

func CompareWaterLevelExcept(uploadTime, waterLevel string, edgeDevice asset.EdgeDevice, logger logger.Logger) *AlertContent {
	rule := getRule(waterLevelExcept)
	if rule == nil {
		logger.Warn(fmt.Sprintf("not found rule name:%s", waterLevelExcept))
		return nil
	}

	// 组装表达式参数
	exprParamsMap := make(map[string]interface{})
	waterLevelInt64, _ := strconv.ParseInt(waterLevel, 10, 64)
	exprParamsMap[strings.ToLower("currVal")] = waterLevelInt64
	rangeLowerLimit, _ := strconv.ParseInt(edgeDevice.Read[0].RangeLowerLimit, 10, 64)
	exprParamsMap[strings.ToLower("rangeLowerLimit")] = rangeLowerLimit
	rangeUpperLimit, _ := strconv.ParseInt(edgeDevice.Read[0].RangeUpperLimit, 10, 64)
	exprParamsMap[strings.ToLower("rangeUpperLimit")] = rangeUpperLimit
	criticalLowerLimit, _ := strconv.ParseInt(edgeDevice.Read[0].CriticalLowerLimit, 10, 64)
	exprParamsMap[strings.ToLower("criticalLowerLimit")] = criticalLowerLimit
	criticalUpperLimit, _ := strconv.ParseInt(edgeDevice.Read[0].CriticalUpperLimit, 10, 64)
	exprParamsMap[strings.ToLower("criticalUpperLimit")] = criticalUpperLimit

	// 创建AlertContent并执行表达式
	alertContent := AlertContent{}
	alertContent.Title = rule.Title
	alertContent.EventType = rule.EventType
	alertContent.Notice = rule.Notice
	alertContent.NeedAlert = false
	alertContent.CurrVal = waterLevel
	alertContent.executorExpr(*rule, exprParamsMap)
	// 生成Description
	paramsMap := make(map[string]interface{})
	paramsMap["assetName"] = edgeDevice.Asset.Name
	paramsMap["dateTime"] = uploadTime
	paramsMap["currVal"] = alertContent.CurrVal
	paramsMap["threshold"] = alertContent.Threshold
	description := assemblyDesc(alertContent.Content.Description, paramsMap)
	alertContent.Description = description

	return &alertContent
}

func CompareReportTimeOut(uploadTime, continueReport string, edgeDevice asset.EdgeDevice , logger logger.Logger) *AlertContent {
	rule := getRule(reportTimeOut)
	if rule == nil {
		logger.Warn(fmt.Sprintf("not found rule name:%s", reportTimeOut))
		return nil
	}

	// 组装表达式参数
	exprParamsMap := make(map[string]interface{})
	continueReportInt64, _ := strconv.ParseInt(continueReport, 10, 64)
	exprParamsMap[strings.ToLower("currVal")] = continueReportInt64

	// 创建AlertContent并执行表达式
	alertContent := AlertContent{}
	alertContent.Title = rule.Title
	alertContent.EventType = rule.EventType
	alertContent.Notice = rule.Notice
	alertContent.NeedAlert = false
	alertContent.CurrVal = continueReport
	alertContent.executorExpr(*rule, exprParamsMap)
	// 生成Description
	paramsMap := make(map[string]interface{})
	paramsMap["assetName"] = edgeDevice.Asset.Name
	paramsMap["dateTime"] = uploadTime
	description := assemblyDesc(alertContent.Content.Description, paramsMap)
	alertContent.Description = description
	return &alertContent
}

func CompareWaterLevelUpExcept(uploadTime, upPercent string, edgeDevice asset.EdgeDevice , logger logger.Logger) *AlertContent {
	rule := getRule(waterLevelUpExcept)
	if rule == nil {
		logger.Warn(fmt.Sprintf("not found rule name:%s", waterLevelUpExcept))
		return nil
	}

	// 组装表达式参数
	exprParamsMap := make(map[string]interface{})
	upPercentFloat64, _ := strconv.ParseFloat(upPercent, 64)
	exprParamsMap[strings.ToLower("currVal")] = upPercentFloat64

	// 创建AlertContent并执行表达式
	alertContent := AlertContent{}
	alertContent.Title = rule.Title
	alertContent.EventType = rule.EventType
	alertContent.Notice = rule.Notice
	alertContent.NeedAlert = false
	alertContent.CurrVal = upPercent
	alertContent.executorExpr(*rule, exprParamsMap)
	// 生成Description
	paramsMap := make(map[string]interface{})
	paramsMap["assetName"] = edgeDevice.Asset.Name
	paramsMap["dateTime"] = uploadTime
	paramsMap["currVal"] = alertContent.CurrVal
	description := assemblyDesc(alertContent.Content.Description, paramsMap)
	alertContent.Description = description
	return &alertContent
}

func (ac *AlertContent) executorExpr(rule Rule, exprParamsMap map[string]interface{}) {
	contents := rule.Contents
	for i := 0; i < len(contents); i++ {
		expr, e := expr.Compile(contents[i].Expr)
		if e != nil {
			fmt.Println("Compile error:" + e.Error())
		}

		res := expr.Eval(func(key string) interface{}{
			ac.Threshold = convertToString(exprParamsMap[strings.ToLower(key)])
			return exprParamsMap[strings.ToLower(key)]
		})
		if res.(bool) {
			ac.Content = contents[i]
			ac.NeedAlert = true
		}
	}
}

func convertToString(val interface{}) string {
	switch v := val.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v,10)
	}
	return val.(string)
}

func assemblyDesc(desc string, paramsMap map[string]interface{}) string {
	buf := new(bytes.Buffer)
	tmpl, _ := template.New("desc").Parse(desc)
	tmpl.Execute(buf, paramsMap)
	return string(buf.Bytes()[:])
}