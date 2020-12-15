package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/bwmarrin/snowflake"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"project/server/config"
	"project/server/cron"
	"project/server/dbUtil"
	"project/server/po"
	"project/server/resolve"
	"strings"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Token    struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"`
		Timestamp   int64  `json:"timestamp"`
	} `json:"token"`
}

var ExpiresTime = 30 * time.Minute

func init() {
	initJob()
}

func initJob()  {
	var servers []po.ServeInfo
	_ = dbUtil.Execute(func(db *storm.DB) error {
		er := db.Select(q.Eq("Status","HOLD")).Find(&servers)
		return er
	})
	for _,serve := range servers {
		_ = dbUtil.Execute(func(db *storm.DB) error {
			er := db.UpdateField(&serve,"Status","RUN")
			return er
		})
	}

	var serves []po.ServeInfo
	err := dbUtil.Execute(func(db *storm.DB) error {
		err := db.Select(q.Eq("Status", "RUN")).Find(&serves)
		return err
	})
	if err == nil {
		for _,serve := range serves {
			cron.Run(serve.ID,parseJob)
		}
	}
}

func ValidateLogin(c *gin.Context) (User, error) {
	//这一部分可以替换成从session/cookie中获取，
	var user User
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &user)

	if err != nil {
		panic(err)
	}

	username := user.Username
	password := user.Password

	flag := false

	if len(username) != 0 && config.Cfg != nil {
		flag = true
	}

	var pwd string
	if flag {
		pwd = config.Cfg.Section(username).Key("password").String()
	}

	if len(pwd) == 0 {
		flag = false
	}

	if flag && password == pwd {
		user.Nickname = config.Cfg.Section(username).Key("nickname").String()
		user.Email = config.Cfg.Section(username).Key("email").String()
		flag = true
	}
	if flag {
		newUser, err2 := CreateToken(user, c)
		if err2 != nil {
			return newUser, err2
		}
		user.Password = ""

		return user, nil
	} else {
		return user, errors.New("用户验证不合法")
	}

}

func ReshToken(token *jwt.Token, c *gin.Context) {
	claims := token.Claims.(jwt.MapClaims)
	now := time.Now().Add(10 * time.Minute).Unix()
	willValid := claims.VerifyExpiresAt(now, true)
	if !willValid {
		fmt.Println("刷新token")
		newClaims := &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ExpiresTime).Unix(), // 过期时间，必须设置
			Issuer:    fmt.Sprintf("%v", claims["iss"]),
			Subject:   fmt.Sprintf("%v", claims["sub"]), // 可不必设置，也可以填充用户名，
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims) //生成token
		accessToken, err := token.SignedString([]byte("vector.sign"))
		if err != nil {
			panic(err)
		}
		hostDoman := c.Request.Host
		comma := strings.Index(hostDoman, ":")
		if comma > 0 {
			hostDoman = hostDoman[:comma]
		}
		c.SetCookie("AccessToken", accessToken, 3600, "/", hostDoman, false, true)

	}
}

func CreateToken(user User, c *gin.Context) (User, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ExpiresTime).Unix(), // 过期时间，必须设置
		Issuer:    user.Username,
		Subject:   user.Nickname, // 可不必设置，也可以填充用户名，
	}
	expired := time.Now().Add(ExpiresTime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //生成token
	accessToken, err := token.SignedString([]byte("vector.sign"))
	user.Token.ExpiresAt = expired
	user.Token.AccessToken = accessToken
	user.Token.Timestamp = time.Now().Unix()
	if err != nil {
		return user, err
	}
	hostDoman := c.Request.Host
	comma := strings.Index(hostDoman, ":")
	if comma > 0 {
		hostDoman = hostDoman[:comma]
	}
	c.SetCookie("AccessToken", accessToken, 3600, "/", hostDoman, false, true)
	return user, nil
}

func Validate(router *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		//这一部分可以替换成从session/cookie中获取，
		authorization, err := c.Cookie("AccessToken")
		token, err := jwt.Parse(authorization, func(token *jwt.Token) (i interface{}, e error) {
			return []byte("vector.sign"), nil
		})
		if err != nil {
			// 第一种
			//if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			//  fmt.Println("+++")
			//  return
			//}
			//fmt.Println([]byte("vector.sign"))

			// 第二种
			if err, ok := err.(*jwt.ValidationError); ok {
				if err.Errors&jwt.ValidationErrorMalformed != 0 {
					fmt.Println(err)
				}
				if err.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					fmt.Println(err)
				}
			}
			GotoLogin(c, router)
			return
		}

		if !token.Valid {
			GotoLogin(c, router)
		} else {
			ReshToken(token, c)
		}

	}
}

func GotoLogin(c *gin.Context, router *gin.Engine) {
	c.Abort()
	c.Request.URL.Path = "/goLogin"
	router.HandleContext(c)
}

func UpdateInfo(c *gin.Context) error {
	authorization, err := c.Cookie("AccessToken")
	if err != nil {
		return err
	}
	user, err := GetUserFromCookie(authorization)
	var newUser User
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	_ = json.Unmarshal(data, &newUser)

	config.Cfg.Section(user.Username).Key("nickname").SetValue(newUser.Nickname)
	config.Cfg.Section(user.Username).Key("email").SetValue(newUser.Email)
	err = config.Cfg.SaveTo(config.UserFilePath)
	return nil

}

func UpdatePwd(c *gin.Context) error {
	authorization, err := c.Cookie("AccessToken")
	if err != nil {
		return err
	}
	user, err := GetUserFromCookie(authorization)
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	result := gjson.ParseBytes(data)

	oldPwd := result.Get("oldPassword").String()
	newPwd := result.Get("newPassword").String()

	pwd := config.Cfg.Section(user.Username).Key("password").String()

	if oldPwd != pwd {
		return errors.New("原密码不匹配！")
	}

	config.Cfg.Section(user.Username).Key("password").SetValue(newPwd)
	err = config.Cfg.SaveTo(config.UserFilePath)
	return nil

}

func GetUserFromCookie(name string) (User, error) {
	token, err := jwt.Parse(name, func(token *jwt.Token) (i interface{}, e error) {
		return []byte("vector.sign"), nil
	})
	var user User
	if err != nil {
		if err, ok := err.(*jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorMalformed != 0 {
				fmt.Println(err)
			}
			if err.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				fmt.Println(err)
			}
		}
		return user, err
	}
	claims := token.Claims.(jwt.MapClaims)

	user.Username = fmt.Sprintf("%v", claims["iss"])
	user.Nickname = config.Cfg.Section(user.Username).Key("nickname").String()
	user.Email = config.Cfg.Section(user.Username).Key("email").String()
	return user, nil
}

func InsertServeInfo(c *gin.Context) error {
	var serve po.ServeInfo

	data, err := ioutil.ReadAll(c.Request.Body)
	checkErr(err)

	_ = json.Unmarshal(data, &serve)

	checkErr(err)
	serve.CreatedAt = time.Now()
	if serve.ID != 0 {
		return UpdateServeInfo(serve)
	}else{
		node, err := snowflake.NewNode(1)
		id := node.Generate()
		serve.ID = id.Int64()
		err = dbUtil.Execute(func(db *storm.DB) error{
			er := db.Save(&serve)
			checkErr(er)
			return er
		})
		return err
	}

}
func UpdateServeInfo(serve po.ServeInfo) error {
	var err error
	err = dbUtil.Execute(func(db *storm.DB) error{
		er := db.Update(&serve)
		return er
	})

	return err
}
func UpdateServeField(c *gin.Context) error {

	data, err := ioutil.ReadAll(c.Request.Body)
	checkErr(err)

	result := gjson.ParseBytes(data)
	if result.IsArray() {
		result.ForEach(func(key,value gjson.Result) bool {
			var server po.ServeInfo
			server.ID = value.Get("id").Int()
			err = dbUtil.Execute(func(db *storm.DB) error {
				er := db.UpdateField(&server, value.Get("field").String(), value.Get("value").String())
				return er
			})
			cron.ChangeJob(value.Get("id").Int(),value.Get("value").String(), parseJob)
			return true
		})
	} else {
		var server po.ServeInfo
		server.ID = result.Get("id").Int()
		err = dbUtil.Execute(func(db *storm.DB) error{
			er := db.UpdateField(&server, result.Get("field").String(), result.Get("value").String())
			return er
		})
		cron.ChangeJob(result.Get("id").Int(),result.Get("value").String(), parseJob)
		checkErr(err)
	}
	return err


}
func DeleteServeInfo(c *gin.Context) error {
	var serve po.ServeInfo

	data, err := ioutil.ReadAll(c.Request.Body)
	checkErr(err)

	_ = json.Unmarshal(data, &serve)
	err = dbUtil.Execute(func(db *storm.DB) error{
		er := db.DeleteStruct(&serve)
		return er
	})
	checkErr(err)
	cron.Remove(serve.ID)
	return err
}

//func GetServeInfo(pageSize int, pageNum int, matchers ...q.Matcher) ([]ServeInfo, int) {
func GetServeInfo(c *gin.Context) po.Page {
	var page po.Page
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &page)

	matchers := initQuery(page.Query)

	var serve po.ServeInfo
	var total int
	err = dbUtil.Execute(func(db *storm.DB) error {
		var er error
		if matchers == nil {
			total, er = db.Count(&serve)
		}else {
			total, er = db.Select(matchers...).Count(&serve)
		}
		return er
	})
	//checkErr(err2)
	serves := []po.ServeInfo{}
	if total > 0 {
		if matchers == nil {
			err = dbUtil.Execute(func(db *storm.DB) error{
				er := db.Select().OrderBy("CreatedAt").Reverse().Skip(page.PageSize * (page.PageNumber - 1)).Limit(page.PageSize).Find(&serves)
				return er
			})
			checkErr(err)
		}else {
			err = dbUtil.Execute(func(db *storm.DB) error{
				er := db.Select(matchers...).OrderBy("CreatedAt").Reverse().Skip(page.PageSize * (page.PageNumber - 1)).Limit(page.PageSize).Find(&serves)
				return er
			})
			checkErr(err)
		}

	}
	page.Total = total
	page.Rows = serves
	return page

}

func SuccessChart() po.IndexChart {
	var serveInfos []po.ServeInfo
	_ = dbUtil.Execute(func(db *storm.DB) error {
		er := db.All(&serveInfos)
		return er
	})

	labelsCount := make([]string,len(serveInfos))
	dataCount := make([]int,len(serveInfos))
	for i,serveInfo := range serveInfos {
		labelsCount[i] = serveInfo.Name
		_ = dbUtil.ExecuteLog(func(db *storm.DB) error {
			count, err := db.Select(q.Eq("ServeId",serveInfo.ID)).Count(&po.RequestLog{})
			success, err := db.Select(q.Eq("ServeId",serveInfo.ID),q.Eq("Success",true)).Count(&po.RequestLog{})
			if count == 0 {
				dataCount[i] = 0
			}else {
				dataCount[i] = success*100.00/count
			}
			return err
		})
	}
	return po.IndexChart{
		Data:   dataCount,
		Labels: labelsCount,
	}
}

func TimeChart() po.IndexChart {
	var requestLogs []po.RequestLog
	_ = dbUtil.ExecuteLog(func(db *storm.DB) error {
		er := db.Select().OrderBy("CreatedAt").Reverse().Limit(50).Find(&requestLogs)
		return er
	})

	labelsCount := make([]string,len(requestLogs))
	dataCount := make([]int,len(requestLogs))
	for i,requestLog := range requestLogs {
		labelsCount[i] = requestLog.CreatedAt.Format("2006-01-02 15:04:05")
		dataCount[i] = requestLog.RequestTime
	}
	return po.IndexChart{
		Data:   dataCount,
		Labels: labelsCount,
	}
}

func IndexInfo() po.IndexInfo {
	var indexInfo po.IndexInfo
	_ = dbUtil.Execute(func(db *storm.DB) error {
		count, er := db.Count(&po.ServeInfo{})
		indexInfo.ServeCount = count
		return er
	})
	_ = dbUtil.ExecuteLog(func(db *storm.DB) error {
		count, err := db.Count(&po.RequestLog{})
		countSuccess, err := db.Select(q.Eq("Success",true)).Count(&po.RequestLog{})
		countFail, err := db.Select(q.Eq("Success",false)).Count(&po.RequestLog{})
		indexInfo.RequestCount = count
		indexInfo.SuccessCount = countSuccess
		indexInfo.FailCount = countFail
		return err
	})
	return indexInfo
}

func ClearRequestLog() error {

	var reqLog po.RequestLog
	var err error
	err = dbUtil.ExecuteLog(func(db *storm.DB) error{
		er := db.Drop(&reqLog)
		return er
	})
	dbUtil.DeleteLogFile()
	return err

}

func GetRequestLog(c *gin.Context) po.Page {
	var page po.Page
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &page)

	matchers := initQuery(page.Query)

	var reqLog po.RequestLog
	var total int
	err = dbUtil.ExecuteLog(func(db *storm.DB) error {
		var er error
		if matchers == nil {
			total, er = db.Count(&reqLog)
		}else {
			total, er = db.Select(matchers...).Count(&reqLog)
		}
		return er
	})
	//checkErr(err2)
	reqLogs := []po.RequestLog{}
	if total > 0 {
		if matchers == nil {
			err = dbUtil.ExecuteLog(func(db *storm.DB) error {
				er := db.Select().OrderBy("CreatedAt").Reverse().Skip(page.PageSize * (page.PageNumber - 1)).Limit(page.PageSize).Find(&reqLogs)
				return er
			})
			checkErr(err)
		}else {
			err = dbUtil.ExecuteLog(func(db *storm.DB) error{
				er := db.Select(matchers...).OrderBy("CreatedAt").Reverse().Skip(page.PageSize * (page.PageNumber - 1)).Limit(page.PageSize).Find(&reqLogs)
				return er
			})
			checkErr(err)
		}

	}
	page.Total = total
	page.Rows = reqLogs
	return page

}

func GetServe(id int64) *po.ServeInfo {
	var serve po.ServeInfo
	var err error
	err = dbUtil.Execute(func(db *storm.DB) error{
		er := db.One("ID",id,&serve)
		return er
	})
	checkErr(err)
	return &serve
}

func parseJob(id int64,job *cron.MyJob)  {
	server := GetServe(id)
	if server != nil {
		entryID, err := job.Job.AddFunc(server.Cron, func() {
			reqLog := resolve.Resolve(server)
			if reqLog.ID != 0 {
				_ = dbUtil.ExecuteLog(func(db *storm.DB) error {
					er := db.Save(&reqLog)
					return er
				})
			}
		})
		checkErr(err)
		job.Job.Start()
		job.EntryID = entryID
	}
}

func initQuery(query string) []q.Matcher {

	result := gjson.Parse(query)
	var matchers []q.Matcher
	if result.IsArray() {
		result.ForEach(func(key,line gjson.Result) bool {
			ruleType := line.Get("type").String()
			var matcher q.Matcher
			switch ruleType {
			case "re":
				matcher = q.Re(line.Get("field").Str,line.Get("value").Str)
				break
			case "eq":
				matcher = q.Eq(line.Get("field").Str,line.Get("value").Str)
				break
			case "lte":
				matcher = q.Lte(line.Get("field").Str,line.Get("value").Time())
				break
			case "gte":
				matcher = q.Gte(line.Get("field").Str,line.Get("value").Time())
				break
			case "in":
				matcher = q.In(line.Get("field").Str,line.Get("value").Array())
				break
			default:
				break

			}
			matchers = append(matchers, matcher)
			return true
		})
	}
	return matchers

}

func checkErr(err error) {
	if err != nil && err.Error() != "not found"{
		panic(err)
	}
}
