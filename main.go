package main

//执行方式：
//alert -to=接收人 -agent=应用id	-color=消息头部颜色	-corpid=corpid		   -corpsecret=corpsecret
//alert -to=@all  -agent=29481187 -color=FFE61A1A 	-corpid=dingd123465865 -corpsecret=zC5Jbed9S
//CorpID和CorpSecret可以在钉钉后台找到
import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

type MsgInfo struct {
	//消息属性和内容
	To, Agentid, Corpid, Corpsecret, Msg, Url, Style string
}

var msgInfo MsgInfo
var logPath string

type Alert struct {
	From                   string `json:"from" xml:"from"`
	Time                   string `json:"time" xml:"time"`
	Level                  string `json:"level" xml:"level"`
	Name                   string `json:"name" xml:"name"`
	Key                    string `json:"key" xml:"key"`
	Value                  string `json:"value" xml:"value"`
	Now                    string `json:"now" xml:"now"`
	ID                     string `json:"id" xml:"id"`
	IP                     string `json:"ip" xml:"ip"`
	Color                  string `json:"color" xml:"color"`
	Url                    string `json:"url" xml:"url"`
	Age                    string `json:"age" xml:"age"`
	Status                 string `json:"status" xml:"status"`
	RecoveryTime           string `json:"recoveryTime" xml:"recoveryTime"`
	Acknowledgement        string `json:"acknowledgement" xml:"acknowledgement"`
	Acknowledgementhistory string `json:"acknowledgementhistory" xml:"acknowledgementhistory"`
}

type DingMsg struct {
	Touser  string `json:"touser"`
	Agentid string `json:"agentid"`
	Msgtype string `json:"msgtype"`
	Oa      struct {
		MessageURL string `json:"message_url"`
		Head       struct {
			Bgcolor string `json:"bgcolor"`
		} `json:"head"`
		Body struct {
			Title string `json:"title"`
			Form  [5]struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"form"`
			Rich struct {
				Num string `json:"num"`
			} `json:"rich"`
			Content string `json:"content"`
			Author  string `json:"author"`
		} `json:"body"`
	} `json:"oa"`
}

func log(message interface{}) {
	pc, file, line, _ := runtime.Caller(1)

	f := runtime.FuncForPC(pc).Name()
	now := time.Now().Format("15:04:05.000")
	date := time.Now().Format("2006-01-02")

	str := fmt.Sprintf("%s %s:%d [%s]: %v", now, file, line, f, message)

	fmt.Println(str)
	if logPath != "" {
		fname := fmt.Sprintf("%s/zabbix_to_dingding_%s.log", logPath, date)
		logfile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		defer logfile.Close()
		if err != nil {
			fmt.Printf("%s %s:%d [%s]: [日志创建错误] %v\r\n", now, file, line, f, err)
		}
		logfile.WriteString(str + "\r\n")
	}
}

func init() {

	flag.StringVar(&msgInfo.To, "to", "@all", "消息的接收人，可以在钉钉后台查看，可空。")
	flag.StringVar(&msgInfo.Agentid, "agentid", "", "AgentID，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Corpid, "corpid", "", "CorpID，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Corpsecret, "corpsecret", "", "CorpSecret，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Msg, "msg", `{ "from": "千思网", "time": "2016.07.28 17:00:05", "level": "Warning", "name": "这是一个千思网（qiansw.com）提供的ZABBIX钉钉报警插件。", "key": "icmpping", "value": "30ms", "now": "56ms", "id": "1637", "ip": "8.8.8.8", "color":"FF4A934A", "age":"3m", "recoveryTime":"2016.07.28 17:03:05", "status":"OK" }`, "Json格式的文本消息内容，不可空。")
	flag.StringVar(&msgInfo.Url, "url", "http://www.qiansw.com/golang-zabbix-alter-to-dingding.html", "消息内容点击后跳转到的URL，可空。")
	flag.StringVar(&msgInfo.Style, "style", "json", "Msg的格式，可选json和xml，推荐使用xml（支持消息中含双引号），可空。")
	flag.StringVar(&logPath, "log", "", "指定存放 log 的目录，不指定则不记录 log。")
	flag.Parse()

	pc, file, line, _ := runtime.Caller(1)

	f := runtime.FuncForPC(pc).Name()
	now := time.Now().Format("15:04:05.000")
	date := time.Now().Format("2006-01-02")

	if logPath != "" {
		fname := fmt.Sprintf("%s/zabbix_to_dingding_%s.log", logPath, date)
		logfile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		defer logfile.Close()
		if err != nil {
			fmt.Printf("%s %s:%d [%s]: [日志创建错误] %v\r\n", now, file, line, f, err)
			os.Exit(1)
		}
		logfile.WriteString("程序启动……" + "\r\n")
	}
	log("初始化完成。")
}

func makeMsg(msg string) string {
	//	根据json或xml文本创建消息体
	log("开始创建消息。")
	log(fmt.Sprintf(`来源消息：
%v`, msg))

	var alert Alert
	if msgInfo.Style == "xml" {
		log("来源消息格式为XML。")
		err := xml.Unmarshal([]byte(msg), &alert)
		if err != nil {
			log(fmt.Sprintf("XML 解析失败：%v", err))
			os.Exit(1)
		}
	} else if msgInfo.Style == "json" {
		log("来源消息格式为Json。")
		err := json.Unmarshal([]byte(msg), &alert)
		if err != nil {
			log(fmt.Sprintf("Json 解析失败：%v", err))
			os.Exit(1)
		}
	} else {
		log("未指定来源消息格式，默认使用Json解析。")
		err := json.Unmarshal([]byte(msg), &alert)
		if err != nil {
			log(fmt.Sprintf("Json 解析失败：%v", err))
			os.Exit(1)
		}
	}
	log("来源消息解析成功。")
	var dingMsg DingMsg
	//给dingMsg各元素赋值
	dingMsg.Touser = msgInfo.To
	dingMsg.Agentid = msgInfo.Agentid
	dingMsg.Msgtype = "oa"
	dingMsg.Oa.MessageURL = msgInfo.Url
	if alert.Url != "" {
		dingMsg.Oa.MessageURL = alert.Url
	}
	dingMsg.Oa.Head.Bgcolor = alert.Color
	dingMsg.Oa.Body.Title = alert.Name
	dingMsg.Oa.Body.Form[0].Key = "告警级别："
	dingMsg.Oa.Body.Form[1].Key = "故障时间："
	dingMsg.Oa.Body.Form[2].Key = "故障时长："
	dingMsg.Oa.Body.Form[3].Key = "IP地址："
	dingMsg.Oa.Body.Form[4].Key = "检测项："
	dingMsg.Oa.Body.Form[0].Value = alert.Level
	dingMsg.Oa.Body.Form[1].Value = alert.Time
	dingMsg.Oa.Body.Form[2].Value = alert.Age
	dingMsg.Oa.Body.Form[3].Value = alert.IP
	dingMsg.Oa.Body.Form[4].Value = alert.Key
	dingMsg.Oa.Body.Rich.Num = alert.Now
	if alert.Status == "PROBLEM" {
		//  故障处理
		dingMsg.Oa.Body.Author = fmt.Sprintf("[%s·%s(%s)]", alert.From, "故障", alert.ID)
		if strings.Replace(alert.Acknowledgement, " ", "", -1) == "Yes" {
			dingMsg.Oa.Body.Content = "故障已经被确认，" + alert.Acknowledgementhistory
		}
	} else if alert.Status == "OK" {
		//  恢复处理
		dingMsg.Oa.Body.Form[0].Key = "故障时间："
		dingMsg.Oa.Body.Form[1].Key = "恢复时间："
		dingMsg.Oa.Body.Form[0].Value = alert.Time
		dingMsg.Oa.Body.Form[1].Value = alert.RecoveryTime
		dingMsg.Oa.Body.Author = fmt.Sprintf("[%s·%s(%s)]", alert.From, "恢复", alert.ID)

	} else if alert.Status == "RESOLVED" {
		//  恢复处理
		dingMsg.Oa.Body.Form[0].Key = "故障时间："
		dingMsg.Oa.Body.Form[1].Key = "恢复时间："
		dingMsg.Oa.Body.Form[0].Value = alert.Time
		dingMsg.Oa.Body.Form[1].Value = alert.RecoveryTime
		dingMsg.Oa.Body.Author = fmt.Sprintf("[%s·%s(%s)]", alert.From, "恢复", alert.ID)

	} else if alert.Status == "msg" {
		dingMsg.Oa.Body.Title = alert.Name
		dingMsg.Oa.Body.Form[0].Key = ""
		dingMsg.Oa.Body.Form[1].Key = ""
		dingMsg.Oa.Body.Form[2].Key = ""
		dingMsg.Oa.Body.Form[3].Key = ""
		dingMsg.Oa.Body.Form[4].Key = ""
		dingMsg.Oa.Body.Form[0].Value = ""
		dingMsg.Oa.Body.Form[1].Value = ""
		dingMsg.Oa.Body.Form[2].Value = ""
		dingMsg.Oa.Body.Form[3].Value = ""
		dingMsg.Oa.Body.Form[4].Value = ""
		dingMsg.Oa.Body.Content = alert.Acknowledgementhistory
	} else {
		//  其他status状况处理
		dingMsg.Oa.MessageURL = "https://www.qiansw.com/golang-zabbix-alter-to-dingding.html"
		dingMsg.Oa.Body.Content = "ZABBIX动作配置有误，请至千思网[qiansw.com]或直接[点击此消息]查看具体配置文档。"
		dingMsg.Oa.Body.Author = fmt.Sprintf("[%s·%s(%s)]", alert.From, alert.Status, alert.ID)
		if strings.Replace(alert.Acknowledgement, " ", "", -1) == "Yes" {
			dingMsg.Oa.Body.Content = "故障已经被确认，" + alert.Acknowledgementhistory
		}
	}
	//	创建post给钉钉的Json文本
	JsonMsg, err := json.Marshal(dingMsg)
	if err != nil {
		log(err)
		os.Exit(1)
	}
	log(fmt.Sprintf("消息创建完成：%s\r\n", string(JsonMsg)))
	return string(JsonMsg)
}

func getToken(corpid, corpsecret string) (token string) { //根据id和secret获取AccessToken
	log("获取Token……")
	type ResToken struct {
		Access_token string
		Errcode      int
		Errmsg       string
	}

	urlstr := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?corpid=%s&corpsecret=%s", corpid, corpsecret)
	u, _ := url.Parse(urlstr)
	q := u.Query()
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())

	if err != nil {
		log(err)
		os.Exit(1)
		return
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log(err)
		os.Exit(1)
		return
	}

	var m ResToken
	err1 := json.Unmarshal(result, &m)

	if err1 == nil {
		if m.Errcode == 0 {
			log("成功！")
			return m.Access_token
		} else {
			log(fmt.Sprintf("AccessToken获取失败，", m.Errmsg))
			os.Exit(1)
			return
		}
		return
	} else {
		log("Token解析失败！")
		os.Exit(1)
		return
	}

}

func sendMsg(token, msg string) (status bool) { //发送OA消息，,返回成功或失败
	log(fmt.Sprintf("需要POST的内容：%v", msg))
	body := bytes.NewBuffer([]byte(msg))
	url := fmt.Sprintf("https://oapi.dingtalk.com/message/send?access_token=%s", token)
	//	fmt.Println(url)
	res, err := http.Post(url, "application/json;charset=utf-8", body)
	if err != nil {
		log(err)
		os.Exit(1)
		return
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log(err)
		os.Exit(1)
		return
	}
	log(fmt.Sprintf("钉钉接口返回消息：%s", result))
	return
}

func main() {
	sendMsg(getToken(msgInfo.Corpid, msgInfo.Corpsecret), makeMsg(msgInfo.Msg))
}
