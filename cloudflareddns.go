package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Config struct {
	CF_API_TOKEN string `json:"CF_API_TOKEN"`
	DOMAIN_NAME  string `json:"DOMAIN_NAME"`
	RECORD_NAME  string `json:"RECORD_NAME"`
}

var (
	cfApiToken string
	domainName string
	recordName string
)

func loadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return err
	}

	cfApiToken = config.CF_API_TOKEN
	domainName = config.DOMAIN_NAME
	recordName = config.RECORD_NAME

	return nil
}

type CloudflareZoneResponse struct {
	Result []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

type CloudflareDNSResponse struct {
	Result []struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"result"`
}

type UpdateDNSRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func getZoneID(domainName string) (string, error) {
	url := "https://api.cloudflare.com/client/v4/zones"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfApiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response CloudflareZoneResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	for _, zone := range response.Result {
		if zone.Name == domainName {
			return zone.Id, nil
		}
	}

	return "", fmt.Errorf("未找到域名 %s 的Zone ID", domainName)
}

func getDNSRecord(zoneID, recordName string) (string, string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, recordName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfApiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var response CloudflareDNSResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", err
	}

	if len(response.Result) == 0 {
		return "", "", fmt.Errorf("未找到DNS记录 %s", recordName)
	}

	return response.Result[0].Id, response.Result[0].Content, nil
}

func getExternalIP() (string, error) {
	resp, err := http.Get("https://ipinfo.io/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func updateDNSRecord(zoneID, recordID, recordName, newIP string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)
	updateRequest := UpdateDNSRequest{
		Type:    "A",
		Name:    recordName,
		Content: newIP,
		TTL:     1,
		Proxied: false,
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfApiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("更新DNS记录失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// 加载配置文件
	err := loadConfig("config.json")
	if err != nil {
		fmt.Println("加载配置文件失败:", err)
		return
	}

	// 打印读取的配置
	fmt.Printf("CF_API_TOKEN: %s\n", cfApiToken)
	fmt.Printf("DOMAIN_NAME: %s\n", domainName)
	fmt.Printf("RECORD_NAME: %s\n", recordName)

	zoneID, err := getZoneID(domainName)
	if err != nil {
		fmt.Println("获取Zone ID失败:", err)
		return
	}

	fmt.Printf("域名 %s 的Zone ID是: %s\n", domainName, zoneID)

	recordID, recordContent, err := getDNSRecord(zoneID, recordName)
	if err != nil {
		fmt.Println("获取DNS记录失败:", err)
		return
	}

	fmt.Printf("DNS记录 %s 的内容是: %s\n", recordName, recordContent)

	externalIP, err := getExternalIP()
	if err != nil {
		fmt.Println("获取外网IP失败:", err)
		return
	}

	fmt.Printf("本地外网IP地址是: %s\n", externalIP)

	if externalIP == recordContent {
		fmt.Println("外网IP与DNS记录匹配，无需更新。")
		return
	}

	fmt.Println("外网IP与DNS记录不匹配，正在更新DNS记录...")
	if err := updateDNSRecord(zoneID, recordID, recordName, externalIP); err != nil {
		fmt.Println("更新DNS记录失败:", err)
		return
	}

	fmt.Println("DNS记录更新成功。")
}
