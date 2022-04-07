package main

import (
	"log"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/imroc/req"
	"github.com/robertkrimen/otto"
)

//常量
const (
	username  string = "账号"
	password  string = "密码"
	loginurl  string = "https://passport.eteams.cn/papi/passport/login/appLogin"         //登陆页面
	checkurl  string = "https://weapp.eteams.cn/api/app/attend/web/sign/getAttendStatus" //检测打卡
	attendurl string = "https://weapp.eteams.cn/api/app/attend/web/sign/sign"            //打卡接口
	jsurl     string = "https://gitee.com/a403/eteams_auto/raw/master/getsignscret.js"   //加密文件
	ppapi     string = "http://pushplus.hxtrip.com/send"                                 //微信推送
)

//变量
var (
	jsessionid, tenantkey, uid, ETEAMSID, message, timecardStatus string
	eteamsuid                                                     int64
)

//微信推送
func PushPlus(content string) {
	param := req.Param{
		"token":    "xxxxxxxxxxxxxxxxxxxxxxxxxx",
		"title":    "eteams自动签到",
		"template": "html",
		"content":  content,
	}
	_, err := req.Get(ppapi, param)
	if err != nil {
		panic(err)
	}
}

//签名
func SignSecret(b string) string {
	header := req.Header{"User-Agent": "yyds"}
	queryparam := req.QueryParam{
		"username": username,
		"password": password,
		"time":     time.Now().Format("2006-01-02_15-04-05"),
	}
	resp, _ := req.Get(jsurl, header, queryparam)
	result := resp.Bytes()
	jsvm := otto.New()
	_, err := jsvm.Run(string(result))
	if err != nil {
		panic(err)
	}
	signvalue, _ := jsvm.Call("signSecret", nil, b)
	return signvalue.String()
}

//登陆页面
func Login() {
	header := req.Header{
		"Content-Type": "application/json;charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (iPhone; CPU iPhone OS 15_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) weapp/4.3.9/public//1.0/",
	}
	queryparam := req.QueryParam{
		"username":   username,
		"password":   password,
		"loginTyp":   "app_account",
		"imei":       "F1A480AB-1A98-44D2-8200-91220A4529F4",
		"version":    "15.1.1",
		"secondImei": "",
		"adviceInfo": "iPhone",
	}
	resp, _ := req.Post(loginurl, header, queryparam)
	result := resp.Bytes()
	jsessionid, tenantkey, eteamsuid, uid, ETEAMSID = GetCookie(result)
	if jsessionid != "" && tenantkey != "" && eteamsuid != 0 && uid != "" && ETEAMSID != "" {
		log.Println("登陆成功！")
	} else {
		log.Println("登陆失败！")
	}
}

//解析JSON
func GetCookie(result []byte) (string, string, int64, string, string) {
	jsonresult, _ := simplejson.NewJson(result)
	jsessionid, _ := jsonresult.Get("jsessionid").String()
	tenantkey, _ := jsonresult.Get("tenantkey").String()
	eteamsuid := jsonresult.Get("eteamsuid").MustInt64()
	uid, _ := jsonresult.Get("uid").String()
	ETEAMSID, _ := jsonresult.Get("ETEAMSID").String()
	return jsessionid, tenantkey, eteamsuid, uid, ETEAMSID
}

//检测打卡类型
func Check() string {
	header := req.Header{
		"Cookie":       "ETEAMSID=" + ETEAMSID + "; ETEAMSID=" + ETEAMSID + "; ETEAMSID=" + ETEAMSID + "; langType=zh_CN; " + "JSESSIONID=" + jsessionid,
		"Content-Type": "application/json;charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (iPhone; CPU iPhone OS 15_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) weapp/4.3.9/public//1.0/",
	}
	resp, _ := req.Get(checkurl, header)
	result := resp.Bytes()
	jsonresult, _ := simplejson.NewJson(result)
	timecardStatus, _ = jsonresult.Get("data").Get("signStatus").String()
	return timecardStatus
}

//打卡
func Attendance() {
	header := req.Header{
		"Cookie":       "ETEAMSID=" + ETEAMSID + "; ETEAMSID=" + ETEAMSID + "; ETEAMSID=" + ETEAMSID + "; langType=zh_CN; " + "JSESSIONID=" + jsessionid,
		"Content-Type": "application/json;charset=UTF-8",
		"User-Agent":   "Mozilla/5.0 (iPhone; CPU iPhone OS 15_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) weapp/4.3.9/public//1.0/",
	}
	param := req.Param{
		"checkAddress": "公司名称",
		"latitude":     "00.000000",
		"longitude":    "00.000000",
		"type":         timecardStatus,
		"sign":         SignSecret("web" + uid + ETEAMSID + tenantkey),
		"userId":       uid,
	}
	queryparam := req.QueryParam{
		"device":    "Mobile",
		"engine":    "WebKit",
		"browser":   "Weapp",
		"os":        "iOS",
		"osVersion": "15.1.1",
		"version":   "4.3.9",
		"language":  "zh_CN",
	}
	resp, _ := req.Post(attendurl, header, req.BodyJSON(param), queryparam)
	result := resp.Bytes()
	jsonresult, _ := simplejson.NewJson(result)
	status, _ := jsonresult.Get("status").Bool()
	message = timecardStatus
	if status {
		log.Println("打卡成功！")
		switch message {
		case "CHECKIN":
			message = "签到成功！"
		case "CHECKOUT":
			message = "签退成功！"
		}
	} else {
		log.Println("打卡失败！")
		message = "签到异常！"
	}
}

func main() {
	Login()
	Check()
	Attendance()
	PushPlus(message)
}
