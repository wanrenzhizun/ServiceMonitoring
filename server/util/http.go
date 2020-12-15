package util

import (
	"errors"
	"github.com/gin-gonic/gin/binding"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
)

type ResponseJson struct {
	Proto         string `json:"proto"` // e.g. "HTTP/1.0"
	Header        string `json:"header"`
	ContentLength int64  `json:"contentLength"`
	StatusCode    int    `json:"statusCode,string"` // e.g. 200
	Status        string `json:"status"`     // e.g. "200 OK"
	Body          string `json:"body"`
}

func HttpDo(method string, url string, paramJson string, headerJson string) (*ResponseJson, error) {
	client := &http.Client{}
	var reqBody string
	var contentType string
	var header gjson.Result
	if len(headerJson) != 0 {
		headerValid := gjson.Valid(headerJson)
		if !headerValid {
			return nil, errors.New("请设置headerJson为json格式")
		}
		header = gjson.Parse(headerJson)
		contentType = header.Get("Content-Type").Str
		if contentType == binding.MIMEJSON{
			contentType = contentType + ";charset=UTF-8"
			if len(paramJson) > 0 {
				paramValid := gjson.Valid(paramJson)
				if !paramValid {
					return nil, errors.New("请设置paramJson为json格式")
				}
				reqBody = paramJson
			}
		}else if contentType == binding.MIMEPOSTForm {
			if len(paramJson) > 0 {
				paramValid := gjson.Valid(paramJson)
				if !paramValid {
					return nil, errors.New("请设置paramJson为json格式")
				}
				param := gjson.Parse(paramJson)
				sendBody := http.Request{}
				_ = sendBody.ParseForm()
				param.ForEach(func(key, value gjson.Result) bool {
					sendBody.Form.Add(key.Str, value.Str)
					return true // keep iterating
				})
				reqBody = sendBody.Form.Encode()
			}
		}else {
			reqBody = paramJson
		}
	}else {
		contentType = binding.MIMEJSON
	}




	respJson := new(ResponseJson)

	/*gjson.Parse(paramJson).ForEach(func(key, value gjson.Result) bool {
		println(key.String())
		return true // keep iterating
	})*/


	req, err := http.NewRequest(method, url, strings.NewReader(reqBody))
	if err != nil {
		respJson.StatusCode = 500
		respJson.Status = "监控系统异常"
		respJson.Body = "{\"message\":\"系统异常\"}"
		return respJson, err
	}

	req.Header.Set("Content-Type", contentType)
	if len(headerJson) > 0 {
		header.ForEach(func(key, value gjson.Result) bool {
			if key.Str != "Content-Type"{
				req.Header.Set(key.Str, value.Str)
			}
			return true
		})
	}

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Cookie", "name=anny")

	resp, err := client.Do(req)
	if err != nil {
		respJson.StatusCode = 501
		respJson.Status = "fail"
		respJson.Body = "{\"message\":\"" + err.Error() + "\"}"
		return respJson, err
	}

	defer resp.Body.Close()

	return setResponse(resp)

}

func setResponse(resp *http.Response) (*ResponseJson, error) {
	resultJson := new(ResponseJson)
	resultJson.StatusCode = resp.StatusCode
	resultJson.Status = resp.Status
	body, err := ioutil.ReadAll(resp.Body)
	resultJson.Body = string(body)
	resultJson.ContentLength = resp.ContentLength
	resultJson.Proto = resp.Proto
	return resultJson, err
}
