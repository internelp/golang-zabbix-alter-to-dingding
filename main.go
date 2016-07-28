package main

//执行方式：
//alter -to=接收人 -agent=应用id	-color=消息头部颜色	-corpid=corpid		   -corpsecret=corpsecret
//alter -to=@all  -agent=29481187 -color=FFE61A1A 	-corpid=dingd123465865 -corpsecret=zC5Jbed9S
//CorpID和CorpSecret可以在钉钉后台找到
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type MsgInfo struct {
	//消息属性和内容
	To, Agentid, Corpid, Corpsecret, Msg, Url string
}

var msgInfo MsgInfo

type Alter struct {
	From         string `json:"from"`
	Time         string `json:"time"`
	Level        string `json:"level"`
	Name         string `json:"name"`
	Key          string `json:"key"`
	Value        string `json:"value"`
	Now          string `json:"now"`
	ID           string `json:"id"`
	IP           string `json:"ip"`
	Color        string
	Age          string
	Status       string
	RecoveryTime string
}

func init() {
	flag.StringVar(&msgInfo.To, "to", "@all", "消息的接收人，可以在钉钉后台查看，可空。")
	flag.StringVar(&msgInfo.Agentid, "agentid", "", "AgentID，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Corpid, "corpid", "", "CorpID，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Corpsecret, "corpsecret", "", "CorpSecret，可以在钉钉后台查看，不可空。")
	flag.StringVar(&msgInfo.Msg, "msg", `{ "from": "千思网", "time": "2016.07.28 17:00:05", "level": "Warning", "name": "这是一个千思网（qiansw.com）提供的ZABBIX钉钉报警插件。", "key": "icmpping", "value": "30ms", "now": "56ms", "id": "1637", "ip": "8.8.8.8", "color":"FF4A934A", "age":"3m", "recoveryTime":"2016.07.28 17:03:05", "status":"OK" }`, "Json格式的文本消息内容，不可空。")
	flag.StringVar(&msgInfo.Url, "url", "http://www.qiansw.com/golang-zabbix-alter-to-dingding.html", "消息内容点击后跳转到的URL，可空。")
	flag.Parse()

}

func makeMsg(msg string) string {
	//	根据json文本创建消息体

	var alter Alter
	err := json.Unmarshal([]byte(msg), &alter)
	if err != nil {
		log.Fatal(err)
	}

	if alter.Status == "PROBLEM" {
		JsonMsg := fmt.Sprintf(`
		{
		"touser":"%s",
		"agentid":"%s",
 		"msgtype": "oa",
	     "oa": {
	        "message_url": "%s",
	        "head": {
	            "bgcolor": "%s"
	        },
	        "body": {
	            "title": "%s",
	            "form": [
	                {
	                    "key": "告警级别：",
	                    "value": "%s"
	                },
	                {
	                    "key": "故障时间：",
	                    "value": "%s"
	                },
					{
	                    "key": "故障时长：",
	                    "value": "%s"
	                },
					{
	                    "key": "IP地址：",
	                    "value": "%s"
	                },

	                {
	                    "key": "检测项：",
	                    "value": "%s"
	                }
	            ],
	            "rich": {
	                "num": "%s"
	            },
	            "content": "",
	            "author": "[%s%s(%s)]"
	        }
	    }
	}`,
			msgInfo.To,
			msgInfo.Agentid,
			msgInfo.Url,
			alter.Color,
			alter.Name,
			alter.Level,
			alter.Time,
			alter.Age,
			alter.IP,
			alter.Key,
			alter.Now,
			alter.From,
			"故障",
			alter.ID)

		return JsonMsg
	} else if alter.Status == "OK" {
		JsonMsg := fmt.Sprintf(`
{
		"touser":"%s",
		"agentid":"%s",
 		"msgtype": "oa",
	     "oa": {
	        "message_url": "%s",
	        "head": {
	            "bgcolor": "%s"
	        },
	        "body": {
	            "title": "%s",
	            "form": [
	                {
	                    "key": "故障时间：",
	                    "value": "%s"
	                },
	                {
	                    "key": "恢复时间：",
	                    "value": "%s"
	                },
					{
	                    "key": "故障时长：",
	                    "value": "%s"
	                },
					{
	                    "key": "IP地址：",
	                    "value": "%s"
	                },

	                {
	                    "key": "检测项：",
	                    "value": "%s"
	                }
	            ],
	            "rich": {
	                "num": "%s"
	            },
	            "content": "",
	            "author": "[%s%s(%s)]"
	        }
	    }
	}`,
			msgInfo.To,
			msgInfo.Agentid,
			msgInfo.Url,
			alter.Color,
			alter.Name,
			alter.Time,
			alter.RecoveryTime,
			alter.Age,
			alter.IP,
			alter.Key,
			alter.Now,
			alter.From,
			"恢复",
			alter.ID)

		return JsonMsg
	} else {
		JsonMsg := fmt.Sprintf(`
{
		"touser":"%s",
		"agentid":"%s",
 		"msgtype": "oa",
	     "oa": {
	        "message_url": "%s",
	        "head": {
	            "bgcolor": "%s"
	        },
	        "body": {
	            "title": "%s",
	            "form": [
	                {
	                    "key": "告警级别：",
	                    "value": "%s"
	                },
	                {
	                    "key": "故障时间：",
	                    "value": "%s"
	                },
					{
	                    "key": "故障时长：",
	                    "value": "%s"
	                },
					{
	                    "key": "IP地址：",
	                    "value": "%s"
	                },

	                {
	                    "key": "检测项：",
	                    "value": "%s"
	                }
	            ],
	            "rich": {
	                "num": "%s"
	            },
	            "content": "ZABBIX动作配置有误，请至千思网（qiansw.com）查看具体配置文档。",
	            "author": "[%s%s(%s)]"
	        }
	    }
	}`,
			msgInfo.To,
			msgInfo.Agentid,
			msgInfo.Url,
			alter.Color,
			alter.Name,
			alter.Level,
			alter.Time,
			alter.Age,
			alter.IP,
			alter.Key,
			alter.Now,
			alter.From,
			alter.Status,
			alter.ID)

		return JsonMsg
	}
}

func getToken(corpid, corpsecret string) (token string) { //根据id和secret获取AccessToken

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
		log.Fatal(err)
		return
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	var m ResToken
	err1 := json.Unmarshal(result, &m)

	if err1 == nil {
		if m.Errcode == 0 {
			return m.Access_token
		} else {
			log.Fatal("AccessToken获取失败，", m.Errmsg)
			return
		}
		return
	} else {
		log.Fatal("Token解析失败！")

		return
	}

}
func sendMsg(token, msg string) (status bool) { //发送OA消息，,返回成功或失败

	JsonMsg := fmt.Sprintf(msg)

	body := bytes.NewBuffer([]byte(JsonMsg))
	url := fmt.Sprintf("https://oapi.dingtalk.com/message/send?access_token=%s", token)
	//	fmt.Println(url)
	res, err := http.Post(url, "application/json;charset=utf-8", body)
	if err != nil {
		log.Fatal(err)
		return
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", result)

	return
}

func main() {
	sendMsg(getToken(msgInfo.Corpid, msgInfo.Corpsecret), makeMsg(msgInfo.Msg))
}
