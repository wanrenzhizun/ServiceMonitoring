package resolve

import (
	"encoding/json"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/blinkbean/dingtalk"
	"github.com/bwmarrin/snowflake"
	"github.com/muesli/cache2go"
	"github.com/tidwall/gjson"
	"project/server/config"
	"project/server/cron"
	"project/server/dbUtil"
	"project/server/po"
	"project/server/util"
	"regexp"
	"strings"
	"time"
)

var failCache *cache2go.CacheTable
var holdCache *cache2go.CacheTable

func init() {
	failCache = cache2go.Cache("FAIL_CACHE")
	holdCache = cache2go.Cache("HOLD_CACHE")
	holdCache.SetAboutToDeleteItemCallback(func(item *cache2go.CacheItem) {
		fmt.Printf("取消挂起：%v\n", item.Data())
		var id = item.Key().(int64)
		err := dbUtil.Execute(func(db *storm.DB) error {
			er := db.UpdateField(&po.ServeInfo{ID: id}, "Status", "RUN")
			return er
		})
		if err == nil {
			_, _ = failCache.Delete(id)
			cron.Run(id,nil)
		}
	})
	holdCache.SetAddedItemCallback(func(item *cache2go.CacheItem) {
		fmt.Printf("任务挂起：%v\n", item.Data())
		var id = item.Key().(int64)
		err := dbUtil.Execute(func(db *storm.DB) error {
			er := db.UpdateField(&po.ServeInfo{ID: id}, "Status", "HOLD")
			return er
		})
		if err == nil {
			cron.Hold(id)
		}
	})
}

func Resolve(info *po.ServeInfo) po.RequestLog {
	var reqLog po.RequestLog
	switch info.Type {
	case "GET":
		reqLog = http(info)
		break
	case "POST":
		reqLog = http(info)
		break
	case "PUT":
		reqLog = http(info)
		break
	case "ORCLE":
		break
	case "MYSQL":
		break
	default:
		break

	}
	return reqLog
}
func http(info *po.ServeInfo) po.RequestLog {
	beforeTime := time.Now()
	res, err := util.HttpDo(info.Type, info.Url, info.Params, info.Header)
	afterTime := time.Now()
	checkErr(err)
	var reqLog po.RequestLog
	result := gjson.Parse(info.Rule)
	if result.IsArray() {
		success := true
		result.ForEach(func(key, value gjson.Result) bool {
			match := value.Get("type").String()
			link := value.Get("link").String()
			resData, err := json.Marshal(res)
			val := gjson.GetBytes(resData,value.Get("field").String()).String()
			if match == "re" {
				re := regexp.MustCompile(value.Get("condition").String())
				checkErr(err)
				success = re.MatchString(val)
				if !success && link == "and" {
					return false
				}
			}else {
				success = result.Get("condition").String() == val
				if !success && link == "and" {
					return false
				}
			}
			return true
		})
		reqLog = resolveResult(info,success,res)
	}else {
		match := result.Get("type").String()
		resData, err := json.Marshal(res)
		val := gjson.GetBytes(resData,result.Get("field").String()).String()
		if match == "re" {
			re := regexp.MustCompile(result.Get("condition").String())
			checkErr(err)
			success := re.MatchString(val)
			reqLog = resolveResult(info,success,res)
		}else {
			condition := result.Get("condition").String()
			if  len(condition) > 0{
				success := result.Get("condition").String() == val
				reqLog = resolveResult(info,success,res)
			}else {
				success := res.StatusCode == 200
				reqLog = resolveResult(info,success,res)
			}

		}
	}
	reqLog.RequestTime = int(afterTime.UnixNano()/1e6 - beforeTime.UnixNano()/1e6)
	return reqLog


}

func resolveResult(info *po.ServeInfo, success bool, res *util.ResponseJson) po.RequestLog {
	if success && !gjson.Valid(res.Body) {
		res.Body = "响应成功，返回非json内容，截取如下：\n" + res.Body[0:50]
	}
	var reqLog po.RequestLog
	node, _ := snowflake.NewNode(1)
	id := node.Generate()
	reqLog.ID = id.Int64()
	reqLog.Url = info.Url
	reqLog.Params = info.Params
	reqLog.CreatedAt = time.Now()
	resData, err := json.Marshal(res)
	checkErr(err)
	reqLog.ResponseBody = gjson.ParseBytes(resData).String()
	reqLog.ServeId = info.ID
	reqLog.ServeName = info.Name
	reqLog.Success = success

	if success {
		_, _ = failCache.Delete(info.ID)
	}else {
		if failCache.Exists(info.ID){
			value, _ := failCache.Value(info.ID)
			failCache.Add(info.ID,0,value.Data().(int) + 1)
		}else {
			failCache.Add(info.ID,0,1)
		}
		value, _ := failCache.Value(info.ID)
		if  value.Data().(int) > info.AllowFail{
			switch info.AlarmType {
			case "DING":
				dingMsg(info,reqLog)
				break
			case "EMAIL":
				emailMsg(info,reqLog)
				break
			case "ALL":
				dingMsg(info,reqLog)
				emailMsg(info,reqLog)
				break
			default:
				break
			}
			waitTime, err := config.Cfg.Section("serve").Key("waitTime").Duration()
			if err != nil{
				waitTime = 60
			}
			holdCache.Add(info.ID,waitTime * time.Minute,info.Name)

		}

	}
	return reqLog
}

func dingMsg(info *po.ServeInfo ,reqLog po.RequestLog)  {
	dingToken := info.TokenKey
	dingTokens := strings.Split(dingToken,",")
	cli := dingtalk.InitDingTalk(dingTokens, info.DingKey)
	msg := "【" + info.Name + "】的服务异常结果信息如下：\n" + reqLog.ResponseBody
	_ = cli.SendTextMessage(msg)
}

func emailMsg(info *po.ServeInfo,reqLog po.RequestLog)  {
	email := info.Email
	emails := strings.Split(email,",")
	//邮件主题为"Hello"
	subject := "【" + info.Name + "】的服务异常通知"
	// 邮件正文
	body := "【" + info.Name + "】的服务异常结果信息如下：<br>" + reqLog.ResponseBody
	_ = util.SendMail(emails, subject, body)
}

func checkErr(err error) {
	if err != nil{
		fmt.Println(err.Error())
	}
}

