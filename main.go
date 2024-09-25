package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// 配置结构体
type Config struct {
	TargetHost    string `ini:"target_host"`
	ListeningPort string `ini:"listening_port"`
}

// fileExists 检查文件是否存在且不是一个目录
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println("文件不存在:", filename)
		return false
	}
	if err != nil {
		fmt.Println("检查文件时出错:", filename, err)
		return false
	}
	return !info.IsDir()
}

func handler(w http.ResponseWriter, r *http.Request, targetHost string) {
	// 检查请求的文件是否存在于 web 文件夹中
	webDir := "web"
	uri := r.RequestURI
	fmt.Println("请求:", uri)
	requestedFile := filepath.Join(webDir, r.URL.Path)
	if uri == "/" {
		requestedFile = filepath.Join(webDir, "/index.html")
	}
	if strings.Contains(r.URL.Path, "reqproc") || strings.Contains(r.URL.Path, "goform") {
		forwardRequest(r, w, targetHost)
	} else if fileExists(requestedFile) {
		// fmt.Println("请求文件:", requestedFile)
		if r.URL.Path == "/index.html" {
			r.URL.Path = "/"
		}
		http.ServeFile(w, r, requestedFile)
	} else {
		fmt.Println("未找到本地文件,转发请求到 " + targetHost)
		forwardRequest(r, w, targetHost)
	}
}

func forwardRequest(r *http.Request, w http.ResponseWriter, targetHost string) {
	// 解析目标服务器的 URL
	targetURL, err := url.Parse("http://" + targetHost + r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		// 对于 GET 请求，将查询参数附加到目标 URL
		queryParams := r.URL.Query()
		targetURL.RawQuery = targetURL.RawQuery + "&" + queryParams.Encode()
	}

	// 复制原始请求到新请求
	newReq, err := http.NewRequest(r.Method, targetURL.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制原始请求的头部信息和主体内容
	newReq.Header = r.Header
	newReq.GetBody = r.GetBody // GetBody 是 http.Request 的一个方法，用于获取请求体

	// 特殊处理 GET 和 POST 请求
	if r.Method == http.MethodPost {
		// 对于 POST 请求，根据 Content-Type 处理请求体
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		newReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		newReq.ContentLength = int64(len(bodyBytes))
	}

	// 发送请求到目标服务器
	client := &http.Client{}
	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将响应内容写回客户端
	w.Write(respBody)
}

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	var config Config
	err = cfg.Section("server").MapTo(&config)
	if err != nil {
		log.Fatalf("无法解析配置: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config.TargetHost)
	})

	// http.HandleFunc("/", handler, config.TargetHost)
	fmt.Println("启动过滤代理服务目标服务器：" + config.TargetHost + "，监听 " + config.ListeningPort + " 端口...")
	if err := http.ListenAndServe(":"+config.ListeningPort, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
