package main

import (
    "bytes"
    "encoding/json"
    "fmt"
	"io"
    "io/ioutil"
    "log"
    "net/http"

    "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

// 配置结构体
type Config struct {
    AccessKeyID     string `json:"AccessKeyID"`
    AccessKeySecret string `json:"AccessKeySecret"`
    DomainName      string `json:"DomainName"`
    Record          string `json:"Record"`
    RecordType      string `json:"RecordType"`
}

// 错误处理辅助函数
func handleError(err error, message string) {
    if err != nil {
        log.Fatalf("%s: %v", message, err)
    }
}

// 获取本地外网 IP 地址
func getExternalIP() (string, error) {
    resp, err := http.Get("http://icanhazip.com")
    if err != nil {
        return "", fmt.Errorf("failed to get external IP: %w", err)
    }
    defer resp.Body.Close()

    var ip bytes.Buffer
    if _, err := io.Copy(&ip, resp.Body); err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }

    return ip.String(), nil
}

// 读取配置文件
func loadConfig(filename string) (Config, error) {
    var config Config
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return config, fmt.Errorf("failed to read config file: %w", err)
    }
    if err := json.Unmarshal(data, &config); err != nil {
        return config, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return config, nil
}

// 更新 DNS 记录
func updateDNSRecord(client *alidns.Client, config Config, newIP string) (string, error) {
    // 查询当前的 DNS 记录
    describeRequest := alidns.CreateDescribeDomainRecordsRequest()
    describeRequest.DomainName = config.DomainName
    describeResponse, err := client.DescribeDomainRecords(describeRequest)
    if err != nil {
        return "", fmt.Errorf("failed to describe domain records: %w", err)
    }

    var recordID, currentIP string
    for _, r := range describeResponse.DomainRecords.Record {
        if r.RR == config.Record && r.Type == config.RecordType {
            recordID = r.RecordId
            currentIP = r.Value // 获取当前记录的 IP
            break
        }
    }

    if recordID == "" {
        return "", fmt.Errorf("record %s not found in domain %s", config.Record, config.DomainName)
    }

    // 更新 DNS 记录
    updateRequest := alidns.CreateUpdateDomainRecordRequest()
    updateRequest.RecordId = recordID
    updateRequest.RR = config.Record
    updateRequest.Type = config.RecordType
    updateRequest.Value = newIP

    if _, err = client.UpdateDomainRecord(updateRequest); err != nil {
        return "", fmt.Errorf("failed to update domain record: %w", err)
    }

    return currentIP, nil
}

func main() {
    // 读取配置文件
    config, err := loadConfig("config.json")
    handleError(err, "Error loading config")

    // 获取本地外网 IP 地址
    newIP, err := getExternalIP()
    handleError(err, "Error getting external IP")

    // 创建阿里云 DNS 客户端
    client, err := alidns.NewClientWithAccessKey("cn-hangzhou", config.AccessKeyID, config.AccessKeySecret)
    handleError(err, "Failed to create client")

    fmt.Printf("New IP to update: %s\n", newIP) // 打印新 IP

    // 调用更新函数
    currentIP, err := updateDNSRecord(client, config, newIP)
    handleError(err, "Failed to update DNS record")

    fmt.Printf("Current IP: %s\n", currentIP) // 打印当前 IP
}
