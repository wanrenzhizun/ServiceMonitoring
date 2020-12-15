本系统采用go语言编写，数据存储采用blot内部存储，不需要连接数据，作为一个简单的web服务监控软件，我不希望有太多的其他模块，那样显得极其臃肿。

本系统支持请求返回数据自定义校验，当任务失败时可以通过钉钉或者邮箱通知，微信通知暂时不在考虑范围。

安装：
     直接点击下载对应系统安装包 [点击下载](https://github.com/wanrenzhizun/ServiceMonitoring/releases)，解压执行ServiceMonitoring文件即可
     默认使用web端口8085，可在./config/user.ini 里面配置
     默认使用./config/user.ini配置文件，
     可指定配置文件运行：
     ./ServiceMonitoring -u yourPath/xxx.ini

自定义校验格式如下：

{

    "type": "re",        //匹配规则：re代表正则匹配，eq代表相等
    "condition": "200",     //匹配条件
    "field": "statusCode",   //需匹配的属性，有：proto；header；statusCode；status；body几个属性
    "link": "and"       //连接条件 and或者or，当多个条件（返回结果判断为数组）时会使用

}

截图：
![登录](https://raw.githubusercontent.com/wanrenzhizun/ServiceMonitoring/master/screen/1.png)
![首页](https://raw.githubusercontent.com/wanrenzhizun/ServiceMonitoring/master/screen/2.png)
![服务](https://raw.githubusercontent.com/wanrenzhizun/ServiceMonitoring/master/screen/3.png)
![日志](https://raw.githubusercontent.com/wanrenzhizun/ServiceMonitoring/master/screen/4.png)

服务状态，RUN启动，STOP 停止，HOLD 暂挂（当服务告警通知后会进行暂挂，挂起时间为1个小时）
