# Proxyfiltert
> 过滤请求静态内容走本地，其余转发至其他地址。

此项目主要用作随身WiFi的web资源调试，模拟前后端分离，本地修改调试静态资源，将后端接口转发到目标服务器。


## 如何使用

1. 拷贝静态资源放进项目 `web `文件夹中      
2. 配置`config.ini`
   1. `server`内的`target_host`表示要转发到的目标主机，默认`192.168.0.1`
   2. `listening_port` 为本地监听端口，默认为`8080`
3. 运行 `Proxyfiltert.exe` 会监听配置端口，默认 `8080 `端口      
4. 浏览器访问 [http://127.0.0.1:8080](http://127.0.0.1:8080) 即可      
5. 本地未找到的资源与请求会转发到 配置的目标服务器默认`192.168.0.1`，请确认设备网络可以正常连接      

# 编译

```
# Windows
go build -o Proxyfiltert.exe main.go

# linux/macos
go build -o Proxyfiltert main.go

```

