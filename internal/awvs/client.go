package awvs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	config Config
	client *http.Client
}

func NewClient(config Config) *Client {
	// 创建一个忽略证书验证的 HTTP 客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略证书验证
		},
	}

	// 如果配置了代理
	// No proxy configuration needed here as it should be part of the AWVS scan configuration

	return &Client{
		config: config,
		client: &http.Client{
			Transport: tr,
		},
	}
}

// 添加扫描目标
func (c *Client) AddTarget(url string) (string, error) {
	payload := map[string]interface{}{
		"address":     url,
		"description": "Added by Go AWVS Scanner",
		"criticality": "10",
	}

	targetID, err := c.post("/api/v1/targets", payload)
	if err != nil {
		return "", fmt.Errorf("添加目标失败: %v", err)
	}

	return targetID, nil
}

// 开始扫描
func (c *Client) StartScan(targetID, scanType string) error {
	// 添加日志输出来调试
	fmt.Printf("使用扫描类型: %s\n", scanType)

	payload := map[string]interface{}{
		"target_id":  targetID,
		"profile_id": scanType,
		"schedule": map[string]interface{}{
			"disable":        false,
			"start_date":     nil,
			"time_sensitive": false,
		},
		"scan_speed":              c.config.ScanSpeed,
		"user_authorized_to_scan": "yes",
		"proxy": map[string]interface{}{
			"address": c.config.ProxyIP,
			"port":    c.config.ProxyPort,
		},
	}

	// 打印完整的请求负载
	jsonData, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Printf("请求负载: %s\n", string(jsonData))

	// 使用正确的路径
	resp, err := c.post("/api/v1/scans", payload)
	if err != nil {
		// 打印详细的错误信息
		if resp != "" {
			return fmt.Errorf("启动扫描失败: %v (响应: %s)", err, resp)
		}
		return fmt.Errorf("启动扫描失败: %v", err)
	}

	return nil
}

// 删除所有目标
func (c *Client) DeleteAllTargets() error {
	// 先获取所有目标
	targets, err := c.GetTargets()
	if err != nil {
		return fmt.Errorf("获取目标失败: %v", err)
	}

	// 逐个删除目标
	for _, target := range targets {
		targetID := fmt.Sprintf("%v", target["target_id"])
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/targets/%s", c.config.APIURL, targetID), nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Auth", c.config.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("删除目标失败，状态码: %d", resp.StatusCode)
		}
	}

	return nil
}

// 删除所有扫描任务
func (c *Client) DeleteAllScans() error {
	// 先获取所有扫描任务
	req, err := http.NewRequest("GET", c.config.APIURL+"/api/v1/scans", nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Auth", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Scans []map[string]interface{} `json:"scans"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// 逐个删除扫描任务
	for _, scan := range result.Scans {
		scanID := fmt.Sprintf("%v", scan["scan_id"])
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/scans/%s", c.config.APIURL, scanID), nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Auth", c.config.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("删除扫描任务失败，状态码: %d", resp.StatusCode)
		}
	}

	return nil
}

// 获取所有目标
func (c *Client) GetTargets() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.config.APIURL+"/api/v1/targets", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Targets []map[string]interface{} `json:"targets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Targets, nil
}

// 修改 post 方法以返回响应内容
func (c *Client) post(path string, payload interface{}) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.config.APIURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Auth", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return string(body), fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return string(body), nil
}
