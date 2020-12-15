package po

import (
	"time"
)

type RequestLog struct {
	ID           int64 `storm:"id" json:"id,string,omitempty"`
	ServeId      int64 `storm:"index" json:"serveId"` //对应ServeInfo的ID
	ServeName    string `json:"serveName"` //对应ServeInfo的名称
	Url          string `json:"url"`//请求地址
	Params       string `json:"params"`//请求参数
	RequestTime  int `json:"requestTime"`//请求参数
	Success      bool `json:"success"`//请求是否成功
	ResponseBody string `json:"responseBody"`//返回结果整体
	CreatedAt    time.Time `storm:"index" json:"createdAt"`
}

type ServeInfo struct {
	ID        int64 `storm:"id" json:"id,string,omitempty"`
	Name      string `json:"name"`//名称
	Group     string `storm:"index" json:"group"` //分组
	AlarmType string `json:"alarmType"`//告警方式，目前支持钉钉或者邮件
	AllowFail int `json:"allowFail,string,omitempty"`//允许失败次数
	TokenKey  string `json:"tokenKey"`//使用钉钉通知时的授权token
	DingKey   string `json:"dingKey"`//使用钉钉通知时的关键字
	Email     string `json:"email"`//指定的通知邮箱
	Url       string `json:"url"`//请求连接
	Type      string `json:"type"`//请求方式
	Header    string `json:"header"`//附加请求头
	Params    string `json:"params"`//请求参数
	Rule      string `json:"rule"`//请求结果判断规则，支持正则表达式
	Status    string `json:"status"`//服务状态，RUN启动，STOP 停止，HOLD 暂挂（当服务告警通知后会进行暂挂，挂起时间为1个小时）
	Cron      string `json:"cron"`//执行间隔，cron表达式
	CreatedAt time.Time `storm:"index" json:"createdAt"` //创建时间
}

type Page struct {
	PageSize int `json:"pageSize"`//每页显示数量
	PageNumber int `json:"pageNumber"`//当前页面
	Total int  `json:"total"`//总数据量
	Query string  `json:"query"`//查询参数
	Rows interface{} `json:"rows"`//返回结果
}

type IndexInfo struct {
	ServeCount int `json:"serveCount"`//每页显示数量
	RequestCount int `json:"requestCount"`//当前页面
	SuccessCount int  `json:"successCount"`//总数据量
	FailCount int  `json:"failCount"`//查询参数
}

type IndexChart struct {
	Data []int `json:"data"`//每页显示数量
	Labels []string `json:"labels"`//当前页面
}
