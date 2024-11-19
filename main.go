package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type WebhookDataMessage struct {
	Url     string                 `json:"url,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
}
type WebhookData struct {
	Ipv4Ip    string              `json:"ipv4Ip,omitempty"`
	Ipv4Hosts string              `json:"ipv4Hosts,omitempty"`
	Message   *WebhookDataMessage `json:"message,omitempty"`
}

var ipStore sync.Map
var filePath = "./data/ip_store.txt"
var ignoreHosts = make(map[string]bool)
var port = "8080"

func main() {
	log.Println("⏳ 开始运行...")
	envFilePath := os.Getenv("DDNS_GO_HOSTS_PATH")
	if len(envFilePath) > 0 {
		filePath = envFilePath
	}
	envAddr := os.Getenv("DDNS_GO_HOSTS_PORT")
	if len(envAddr) > 0 {
		port = envAddr
	}
	envIgnoreHosts := os.Getenv("DDNS_GO_HOSTS_IGNORE")
	if len(envIgnoreHosts) > 0 {
		split := strings.Split(envIgnoreHosts, ",")
		for _, s := range split {
			ignoreHosts[s] = true
		}
	}

	readIpStore()
	// 启动服务
	runServer()
	// 结束运行
	log.Println("❌ 结束运行")
}
func readIpStore() {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	readAll, err := io.ReadAll(file)
	if err != nil {
		return
	}
	split := strings.Split(string(readAll), "\n")
	for _, s := range split {
		split1 := strings.Split(s, " ")
		if len(split1) == 2 {
			ipStore.Store(split1[1], split1[0])
		}
	}
}

func ipStoreToHosts() string {
	fileData := make([]string, 0, 20)
	ipStore.Range(func(key, value interface{}) bool {
		fileData = append(fileData, fmt.Sprintf("%s %s", value, key))
		return true
	})
	slices.SortFunc(fileData, func(a, b string) int {
		return strings.Compare(a, b)
	})
	return strings.Join(fileData, "\n")
}
func webhookDataSaveStore(ipv4Hosts, ipv4Ip string) {
	if len(ipv4Hosts) > 0 && len(ipv4Ip) > 0 {
		ipv4HostList := strings.Split(ipv4Hosts, ",")
		for _, s := range ipv4HostList {
			if strings.HasPrefix(s, "*") {
				continue
			}
			if ignoreHosts[s] {
				continue
			}
			ipStore.Store(s, ipv4Ip)
		}
		_ = os.WriteFile(filePath, []byte(ipStoreToHosts()), 0666)
	}
}

func runServer() {
	handler := http.NewServeMux()
	handler.HandleFunc("/webhook", ddnsWebhook)
	handler.HandleFunc("/hosts", getHosts)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
	if err != nil {
		log.Panic("http.ListenAndServe err:", err)
	}
}

func ddnsWebhook(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	readAll, err := io.ReadAll(request.Body)
	if err != nil {
		_, _ = writer.Write([]byte("error"))
		return
	}
	var ddd WebhookData
	if err = json.Unmarshal(readAll, &ddd); err != nil {
		_, _ = writer.Write([]byte("error"))
		return
	}
	webhookDataSaveStore(ddd.Ipv4Hosts, ddd.Ipv4Hosts)
	if ddd.Message != nil {
		SendMessageByUrl(ddd.Message)
	}
	_, _ = writer.Write([]byte("ok"))
}
func getHosts(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()
	_, err := writer.Write([]byte(ipStoreToHosts()))
	if err != nil {
		return
	}
}

var messageClient = &http.Client{
	Timeout: time.Second * 30,
}

func SendMessageByUrl(message *WebhookDataMessage) {
	marshal, err := json.Marshal(message.Body)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", message.Url, bytes.NewReader(marshal))
	if err != nil {
		log.Println("发送失败", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range message.Headers {
		req.Header.Set(k, v)
	}
	rsp, err := messageClient.Do(req)
	if err != nil {
		log.Println("发送失败", err)
		return
	}
	defer func() {
		_ = rsp.Body.Close()
	}()
	if rsp.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(rsp.Body)
		log.Printf("发送失败 [%d] [%s] %s", rsp.StatusCode, string(all), marshal)
		return
	}
}

// func daemon() {
// 	// 监听停止信号
// 	stop := make(chan os.Signal, 1)
// 	signal.Notify(stop, syscall.SIGTERM)
//
// 	// 启动完成
// 	log.Println("启动完成")
// 	// 等待停止信号
// 	<-stop
// 	log.Println("收到停止信号")
// }
