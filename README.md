# eteams_auto
eteams自动签到

## pushplus获取token（微信推送）
http://pushplus.hxtrip.com/  
一对一推送-->token值

### 配置教程
修改以下参数：  
username        账号  
password        密码  
token           pushplus一对一推送  
checkAddress    公司名称  
latitude        纬度  
longitude       精度  

#### Tips
修改完配置文件之后可以使用Go编译放到VPS上定时执行  
编译：  
go build eteams.go
