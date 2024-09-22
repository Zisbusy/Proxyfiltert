package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

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

func handler(w http.ResponseWriter, r *http.Request) {
	// 检查请求的文件是否存在于 web 文件夹中
	webDir := "web"
	requestedFile := filepath.Join(webDir, r.URL.Path)
	fmt.Println("请求文件:", requestedFile)

	if fileExists(requestedFile) {
		fmt.Print("本地文件存在,返回本地文件")
		http.ServeFile(w, r, requestedFile)
	} else {
		fmt.Println("未找到本地文件,转发请求到 192.168.0.1")
		forwardRequest(r, w)
	}
}

func forwardRequest(r *http.Request, w http.ResponseWriter) {
	// 创建目标服务器的新 URL
	targetURL, err := url.Parse("http://192.168.0.1")
	if err != nil {
		fmt.Println("解析目标 URL 时出错:", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	// 拼接完整的目标 URL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	// 使用 http.Client 发送请求
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		fmt.Println("创建新请求时出错:", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	// 从原始请求复制头部信息到新请求
	for k, v := range r.Header {
		req.Header[k] = v
	}

	// 发送请求并读取响应
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求时出错:", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 将响应复制回客户端
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("启动过滤代理服务，监听 8080 端口...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
