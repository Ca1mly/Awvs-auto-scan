package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"awvs-scan/internal/awvs"
)

func main() {
	fmt.Println("开始运行...")

	a := app.New()
	w := a.NewWindow("AWVS 批量扫描工具")
	w.Resize(fyne.NewSize(800, 600))

	// 创建主标题
	title := widget.NewLabel("AWVS 14/15 批量扫描工具")

	// 创建URL输入区域
	urlInput := widget.NewMultiLineEntry()
	urlInput.SetPlaceHolder("请输入URL（每行一个）\n例如：\nhttp://example.com\nhttps://test.com")
	urlInput.Resize(fyne.NewSize(700, 200))

	// 创建文件选择按钮
	btnSelectFile := widget.NewButton("从文件导入URL", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			var urls []string
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				url := strings.TrimSpace(scanner.Text())
				if url != "" {
					urls = append(urls, url)
				}
			}

			if len(urls) > 0 {
				currentText := urlInput.Text
				if currentText != "" {
					currentText += "\n"
				}
				urlInput.SetText(currentText + strings.Join(urls, "\n"))
				dialog.ShowInformation("成功", fmt.Sprintf("已导入 %d 个URL", len(urls)), w)
			}
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
		fd.Show()
	})

	// 创建主要功能按钮
	btnAddScan := widget.NewButton("批量添加URL到扫描器", func() {
		urls := getURLs(urlInput.Text)
		if len(urls) == 0 {
			dialog.ShowInformation("提示", "请先输入或导入URL", w)
			return
		}
		showScanOptions(w, urls)
	})

	btnDeleteAll := widget.NewButton("删除所有目标和任务", func() {
		// TODO: 实现删除功能
		fmt.Println("删除所有目标和任务")
	})

	btnDeleteTasks := widget.NewButton("仅删除扫描任务", func() {
		// TODO: 实现删除任务功能
		fmt.Println("删除所有扫描任务")
	})

	btnScanExisting := widget.NewButton("扫描已有目标", func() {
		// TODO: 实现扫描已有目标功能
		fmt.Println("扫描已有目标")
	})

	// 添加配置按钮
	btnConfig := widget.NewButton("配置", func() {
		showConfigDialog(w)
	})

	// 创建免责声明
	disclaimer := widget.NewLabel("免责声明：本工具仅用于安全自查，请勿用于非法测试")

	// 使用垂直布局排列所有组件
	content := container.NewVBox(
		title,
		btnConfig,
		container.NewHBox(widget.NewLabel("目标URL：")),
		urlInput,
		btnSelectFile,
		btnAddScan,
		btnDeleteAll,
		btnDeleteTasks,
		btnScanExisting,
		disclaimer,
	)

	w.SetContent(content)
	w.CenterOnScreen()
	fmt.Println("窗口创建成功，即将显示...")
	w.ShowAndRun()
}

// 从文本中提取URL列表
func getURLs(text string) []string {
	var urls []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls
}

// 显示扫描选项窗口
func showScanOptions(parent fyne.Window, urls []string) {
	w := fyne.CurrentApp().NewWindow("选择扫描类型")
	w.Resize(fyne.NewSize(400, 500))

	options := []string{
		"完全扫描",
		"扫描高风险漏洞",
		"扫描XSS漏洞",
		"扫描SQL注入漏洞",
		"弱口令检测",
		"仅爬虫(可配合被动扫描)",
		"扫描已知漏洞",
		"仅添加目标",
		"Apache-Log4j扫描",
		"Bug Bounty高频漏洞",
		"常见CVE扫描",
		"Spring4Shell扫描",
	}

	list := widget.NewList(
		func() int { return len(options) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(options[id])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		scanType := options[id]
		profileID := awvs.ScanTypeMap[scanType]

		config, err := loadConfig()
		if err != nil {
			dialog.ShowError(fmt.Errorf("加载配置失败: %v", err), parent)
			return
		}

		client := awvs.NewClient(config)

		progress := dialog.NewProgress("扫描进度", fmt.Sprintf("正在处理 %d 个目标", len(urls)), parent)
		progress.Show()

		go func() {
			for i, url := range urls {
				progress.SetValue(float64(i) / float64(len(urls)))

				targetID, err := client.AddTarget(url)
				if err != nil {
					dialog.ShowError(fmt.Errorf("添加目标失败 %s: %v", url, err), parent)
					continue
				}

				if err := client.StartScan(targetID, profileID); err != nil {
					dialog.ShowError(fmt.Errorf("启动扫描失败 %s: %v", url, err), parent)
					continue
				}
			}

			progress.Hide()
			dialog.ShowInformation("完成", fmt.Sprintf("已添加 %d 个目标到扫描队列", len(urls)), parent)
		}()

		w.Close()
	}

	w.SetContent(list)
	w.CenterOnScreen()
	w.Show()
}

// 添加新的配置相关函数
func showConfigDialog(parent fyne.Window) {
	w := fyne.CurrentApp().NewWindow("AWVS配置")

	// 基本设置
	apiURLEntry := widget.NewEntry()
	apiURLEntry.SetPlaceHolder("https://your-awvs-host:3443")

	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("1986ad8c0a5b3df4d7028d5f3c06e936")

	// 代理设置
	proxyURLEntry := widget.NewEntry()
	proxyURLEntry.SetPlaceHolder("http://127.0.0.1:8080")

	// 线程数设置
	threadNumEntry := widget.NewEntry()
	threadNumEntry.SetPlaceHolder("10")

	// 扫描速度选择
	speedSelect := widget.NewSelect(awvs.SpeedOptions, nil)
	speedSelect.SetSelected("moderate")

	// 报告类型选择
	reportSelect := widget.NewSelect(awvs.ReportTypes, nil)
	reportSelect.SetSelected("HTML")

	// 加载现有配置
	if config, err := loadConfig(); err == nil {
		apiURLEntry.SetText(config.APIURL)
		apiKeyEntry.SetText(config.APIKey)
		if config.ProxyURL != "" {
			proxyURLEntry.SetText(config.ProxyURL)
		}
		if config.ThreadNum > 0 {
			threadNumEntry.SetText(fmt.Sprintf("%d", config.ThreadNum))
		}
		if config.ScanSpeed != "" {
			speedSelect.SetSelected(config.ScanSpeed)
		}
		if config.ReportType != "" {
			reportSelect.SetSelected(config.ReportType)
		}
	} else {
		// 使用默认配置
		threadNumEntry.SetText(fmt.Sprintf("%d", awvs.DefaultConfig.ThreadNum))
		speedSelect.SetSelected(awvs.DefaultConfig.ScanSpeed)
		reportSelect.SetSelected(awvs.DefaultConfig.ReportType)
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "API URL", Widget: apiURLEntry},
			{Text: "API Key", Widget: apiKeyEntry},
			{Text: "代理地址", Widget: proxyURLEntry},
			{Text: "线程数量", Widget: threadNumEntry},
			{Text: "扫描速度", Widget: speedSelect},
			{Text: "报告类型", Widget: reportSelect},
		},
		OnSubmit: func() {
			threadNum, err := strconv.Atoi(threadNumEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("线程数必须是数字"), w)
				return
			}

			config := awvs.Config{
				APIURL:     apiURLEntry.Text,
				APIKey:     apiKeyEntry.Text,
				ProxyURL:   proxyURLEntry.Text,
				ThreadNum:  threadNum,
				ScanSpeed:  speedSelect.Selected,
				ReportType: reportSelect.Selected,
			}

			if err := saveConfig(config); err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("成功", "配置已保存", w)
			w.Close()
		},
	}

	w.SetContent(form)
	w.Resize(fyne.NewSize(500, 400))
	w.CenterOnScreen()
	w.Show()
}

// 配置文件处理函数
func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "awvs-scan", "config.json")
}

func loadConfig() (awvs.Config, error) {
	var config awvs.Config

	configPath := getConfigPath()
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}

func saveConfig(config awvs.Config) error {
	configPath := getConfigPath()

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, data, 0644)
}
