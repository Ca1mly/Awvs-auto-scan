package awvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	config Config
	client *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
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
	payload := map[string]interface{}{
		"target_id":  targetID,
		"profile_id": scanType,
		"schedule": map[string]interface{}{
			"disable":        false,
			"start_date":     nil,
			"time_sensitive": false,
		},
		"scan_speed": c.config.ScanSpeed,
	}

	_, err := c.post("/api/v1/scans", payload)
	if err != nil {
		return fmt.Errorf("启动扫描失败: %v", err)
	}

	return nil
}

// 删除所有目标
func (c *Client) DeleteAllTargets() error {
	req, err := http.NewRequest("DELETE", c.config.APIURL+"/api/v1/targets", nil)
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("删除失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// 发送POST请求的辅助函数
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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", result["target_id"]), nil
}
