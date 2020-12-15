package controller

import (
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"project/server/api"
	"project/server/config"
	"strings"
)

type MyTemplate struct {
	TemplateName string `uri:"templateName" binding:"required"`
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.HTMLRender = LoadTemplates(config.Cfg.Section("common").Key("baseDir").String() + "templates")
	router.Static("/log/",config.Cfg.Section("common").Key("baseDir").String() +  "static")
	router.GET("/goLogin", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{
			"title": "登录!",
		})
	})
	router.POST("/login", func(c *gin.Context) {
		user, err := api.ValidateLogin(c)
		if err == nil {
			c.JSON(http.StatusOK,user)
		}else {
			c.Abort()
			c.JSON(http.StatusUnauthorized,gin.H{"message":"用户名或密码错误"})
		}

	})

	router.POST("/logout", func(c *gin.Context) {
		c.SetCookie("AccessToken", "", 5, "/", "localhost", false, true)
		c.JSON(http.StatusOK,gin.H{"message":"退出成功"})
	})

	router.Use(api.Validate(router))  //使用validate()中间件身份验证

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "首页",
		})
	})
	router.GET("/index", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "首页",
		})
	})
	router.GET("/includes/:templateName", func(c *gin.Context) {
		var template MyTemplate
		if err := c.ShouldBindUri(&template); err != nil {
			c.JSON(400, gin.H{"msg": err})
			return
		}
		c.HTML(http.StatusOK, template.TemplateName, gin.H{
		})
	})

	router.POST("/updateUserInfo", func(c *gin.Context) {
		err := api.UpdateInfo(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"message":err.Error()})
		}else {
			c.JSON(http.StatusOK,gin.H{"message":"保存成功"})
		}
	})

	router.POST("/updateUserPwd", func(c *gin.Context) {
		err := api.UpdatePwd(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "修改成功"})
		}
	})

	router.POST("/serve/getServeInfos", func(c *gin.Context) {
		page := api.GetServeInfo(c)
		c.JSON(http.StatusOK, page)

	})


	router.POST("/serve/getReqLogs", func(c *gin.Context) {
		page := api.GetRequestLog(c)
		c.JSON(http.StatusOK, page)

	})

	router.POST("/serve/clearLogs", func(c *gin.Context) {
		err := api.ClearRequestLog()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "清除成功"})
		}

	})

	router.POST("/serve/addServeInfos", func(c *gin.Context) {
		err := api.InsertServeInfo(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
		}

	})

	router.POST("/serve/deleteServeInfos", func(c *gin.Context) {
		err := api.DeleteServeInfo(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
		}

	})

	router.POST("/serve/updateServeInfoField", func(c *gin.Context) {
		err := api.UpdateServeField(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
		}

	})

	router.POST("/index/indexInfo", func(c *gin.Context) {
		indexInfo := api.IndexInfo()
		c.JSON(http.StatusOK, indexInfo)
	})

	router.POST("/index/successChart", func(c *gin.Context) {
		successChart := api.SuccessChart()
		c.JSON(http.StatusOK, successChart)
	})

	router.POST("/index/timeChart", func(c *gin.Context) {
		timeChart := api.TimeChart()
		c.JSON(http.StatusOK, timeChart)
	})




	return router

}

func LoadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	layouts, err := filepath.Glob(templatesDir + "/layouts/*.html")
	if err != nil {
		panic(err.Error())
	}

	includes, err := GetFiles(templatesDir + "/includes",".html")
	if err != nil {
		panic(err.Error())
	}

	autonomys, err := GetFiles(templatesDir + "/autonomy",".html")
	if err != nil {
		panic(err.Error())
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFilesFuncs(strings.Replace(include,templatesDir + "/includes/","",1),template.FuncMap{
			"formatTitle": formatTitle,
		}, files...)
	}
	for _, autonomy := range autonomys {
		r.AddFromFiles(strings.Replace(autonomy,templatesDir + "/autonomy/","",1), autonomy)
	}


	return r
}


func GetFiles(folder string,match string) (matches []string, err error) {
	files, _ := ioutil.ReadDir(folder) //specify the current dir
	var filePaths []string
	var er error
	for _,file := range files{
		if file.IsDir(){
			var childFilePaths []string
			childFilePaths, er = GetFiles(folder + "/" + file.Name(),match)
			filePaths = append(filePaths, childFilePaths...)
			er = err
		}else{
			matched := strings.HasSuffix(folder+"/"+file.Name(),match)
			er = err
			if matched {
				filePaths = append(filePaths, folder + "/" + file.Name())
			}
		}
	}
	return filePaths,er
}

func formatTitle(title string) string {
	if len(title) == 0 {
		title = "阿伦"
	}
	return title
}


