package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mainflux/mainflux/alert/rule"
	"github.com/mainflux/mainflux/graphql"
	"github.com/mainflux/mainflux/graphql/asset"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type consumer struct {
	log         logger.Logger
}

var (
	sms    string = "sms"
)

//TODO 跨语言传输, 为了保证数据的可靠性，统一使用string类型
type alertData struct {
	// 设备ID
	DeviceId            string
	// 终端上传时间 yyyy-MM-dd hh:mm:ss
	UploadTime          string
	// 液位值-数字
	WaterLevel          string
	// 是否连续上传-布尔
	ContinueReport      string
	// 与上次比较上涨百分比
	UpPercent           string

}

func Receive(sub  messaging.Subscriber, logger logger.Logger) {
	c := consumer{logger}
	if err := sub.Subscribe("java-nats", c.handler); err != nil {
		logger.Error(fmt.Sprintf("sub error %s", err.Error()))
	}
}

func (c *consumer) handler(msg messaging.Message) error {
	payload := msg.Payload
	var alertData alertData
	json.Unmarshal([]byte(payload), &alertData)
	go runAlert(alertData, c.log)
	return nil
}

func runAlert(alertData alertData, logger logger.Logger) {
	// 获取设备信息
	deviceId := alertData.DeviceId
	edgeDevice := asset.GetEdgeDeviceById(deviceId)
	if edgeDevice == nil {
		logger.Warn(fmt.Sprintf("not found EdgeDevice id:", deviceId))
		return
	}
	// 根据配置的规则判断是否需要报警
	wlExceptAlertContent := rule.CompareWaterLevelExcept(alertData.WaterLevel, *edgeDevice, logger)
	afterRunAlert(*edgeDevice, alertData.UploadTime, *wlExceptAlertContent)

	reportAlertContent := rule.CompareReportTimeOut(alertData.ContinueReport, *edgeDevice, logger)
	afterRunAlert(*edgeDevice, alertData.UploadTime, *reportAlertContent)

	wlUpAlertContent := rule.CompareWaterLevelUpExcept(alertData.UpPercent, *edgeDevice, logger)
	afterRunAlert(*edgeDevice, alertData.UploadTime, *wlUpAlertContent)
}

func afterRunAlert(edgeDevice asset.EdgeDevice, uploadTime string, alertContent rule.AlertContent) {
	if alertContent.NeedAlert {
		// 保存数据库
		if alertContent.TableName == graphql.AlaramEvent {
			go saveAlarmEvent(alertContent)
		}
		if _, ok := alertContent.Notice[sms]; ok {
			go noticePrometheusAlert(edgeDevice, uploadTime, alertContent)
		}
	}
}

func saveAlarmEvent(alertContent rule.AlertContent) {

}

func noticePrometheusAlert(edgeDevice asset.EdgeDevice, uploadTime string, alertContent rule.AlertContent) {
	if len(asset.GetPhones()) == 0 {
		return
	}
	notice := alertContent.Notice[sms]

	paramsMap := make(map[string]string)
	paramsMap["alertName"] = alertContent.Title
	paramsMap["assetName"] = edgeDevice.Name
	paramsMap["dateTime"] = uploadTime
	paramsMap["desc"] = alertContent.Content.Title

	paramsBuf := new(bytes.Buffer)
	tmpl, _ := template.New("sms").Parse(notice.Template)
	tmpl.Execute(paramsBuf, paramsMap)

	url := notice.Url + strings.Join(asset.GetPhones(), ",")
	//fmt.Println("url:", url, "params:", paramsBuf)
	client := &http.Client{Timeout: notice.Timeout * time.Second}
	request, _ := http.NewRequest("POST", url, paramsBuf)
	resp, e := client.Do(request)
	if e != nil {
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var bodyMap map[string]interface{}
	json.Unmarshal(body, &bodyMap)
}