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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"awvs-scan/internal/awvs"
)

func main() {
	fmt.Println("开始运行...")

	a := app.New()
	w := a.NewWindow("AWVS 批量扫描工具")
	w.Resize(fyne.NewSize(800, 800))

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

	// 创建日志显示区域
	logText := widget.NewMultiLineEntry()
	logText.Disable() // 使其只读
	logText.Wrapping = fyne.TextWrapWord
	logText.MultiLine = true
	logText.SetMinRowsVisible(15) // 设置最小可见行数

	// 使用更大的容器来显示日志
	logScroll := container.NewScroll(logText)
	logScroll.SetMinSize(fyne.NewSize(780, 300)) // 设置最小大小

	// 创建一个固定大小的容器来包装滚动区域
	logContainer := container.NewVBox(
		widget.NewLabel("运行日志："),
		logScroll,
	)

	// 更新日志显示的函数
	updateLog := func(msg string) {
		currentTime := time.Now().Format("15:04:05")
		logMsg := fmt.Sprintf("[%s] %s\n", currentTime, msg)
		logText.SetText(logText.Text + logMsg)
	}

	// 创建扫描类型选择
	var scanTypes []string
	for scanType := range awvs.ScanTypeMap {
		scanTypes = append(scanTypes, scanType)
	}
	scanTypeSelect := widget.NewSelect(scanTypes, func(selected string) {
		// 当选择改变时显示描述
		if desc, ok := awvs.ScanTypeDescriptions[selected]; ok {
			updateLog(fmt.Sprintf("选择扫描类型: %s (%s)", selected, desc))
		}
	})
	scanTypeSelect.SetSelected("完全扫描")
	scanTypeSelect.PlaceHolder = "选择扫描类型"

	// 创建按钮
	btnConfig := widget.NewButton("配置", func() {
		showConfigDialog(w)
	})

	btnAddScan := widget.NewButton("批量添加URL到扫描器", func() {
		urls := getURLs(urlInput.Text)
		if len(urls) == 0 {
			dialog.ShowInformation("提示", "请先输入或导入URL", w)
			return
		}

		scanType := scanTypeSelect.Selected
		if scanType == "" {
			dialog.ShowInformation("提示", "请选择扫描类型", w)
			return
		}

		profileID := awvs.ScanTypeMap[scanType]
		config, err := loadConfig()
		if err != nil {
			dialog.ShowError(fmt.Errorf("加载配置失败: %v", err), w)
			return
		}

		client := awvs.NewClient(config)

		// 创建不可关闭的进度窗口
		progressWindow := fyne.CurrentApp().NewWindow("扫描进度")
		progressWindow.SetFixedSize(true) // 禁止调整大小
		progressWindow.CenterOnScreen()

		progress := widget.NewProgressBar()
		progressLabel := widget.NewLabel(fmt.Sprintf("正在处理 %d 个目标...", len(urls)))

		// 添加取消按钮
		cancelChan := make(chan struct{})
		btnCancel := widget.NewButton("取消", func() {
			close(cancelChan)
			progressWindow.Close()
		})

		progressContent := container.NewVBox(
			progressLabel,
			progress,
			btnCancel,
		)

		progressWindow.SetContent(progressContent)
		progressWindow.Show()

		go func() {
			defer progressWindow.Close()

			for i, url := range urls {
				select {
				case <-cancelChan:
					updateLog("用户取消了扫描")
					return
				default:
					progress.SetValue(float64(i) / float64(len(urls)))
					progressLabel.SetText(fmt.Sprintf("正在处理: %s (%d/%d)", url, i+1, len(urls)))
					updateLog(fmt.Sprintf("正在处理: %s", url))

					targetID, err := client.AddTarget(url)
					if err != nil {
						updateLog(fmt.Sprintf("添加目标失败 %s: %v", url, err))
						continue
					}

					if err := client.StartScan(targetID, profileID); err != nil {
						updateLog(fmt.Sprintf("启动扫描失败 %s: %v", url, err))
						continue
					}

					updateLog(fmt.Sprintf("成功添加并启动扫描: %s", url))
				}
			}

			updateLog(fmt.Sprintf("完成！已添加 %d 个目标到扫描队列", len(urls)))
			dialog.ShowInformation("完成", fmt.Sprintf("已添加 %d 个目标到扫描队列", len(urls)), w)
		}()
	})

	btnDeleteAll := widget.NewButton("删除所有目标和任务", func() {
		dialog.ShowConfirm("确认", "确定要删除所有目标和任务吗？", func(ok bool) {
			if !ok {
				return
			}

			config, err := loadConfig()
			if err != nil {
				dialog.ShowError(fmt.Errorf("加载配置失败: %v", err), w)
				return
			}

			client := awvs.NewClient(config)
			updateLog("正在删除所有目标和任务...")

			if err := client.DeleteAllTargets(); err != nil {
				updateLog(fmt.Sprintf("删除失败: %v", err))
				dialog.ShowError(err, w)
				return
			}

			updateLog("删除成功")
			dialog.ShowInformation("成功", "已删除所有目标和任务", w)
		}, w)
	})

	btnDeleteTasks := widget.NewButton("仅删除扫描任务", func() {
		dialog.ShowConfirm("确认", "确定要删除所有扫描任务吗？", func(ok bool) {
			if !ok {
				return
			}

			config, err := loadConfig()
			if err != nil {
				dialog.ShowError(fmt.Errorf("加载配置失败: %v", err), w)
				return
			}

			client := awvs.NewClient(config)
			updateLog("正在删除所有扫描任务...")

			if err := client.DeleteAllScans(); err != nil {
				updateLog(fmt.Sprintf("删除失败: %v", err))
				dialog.ShowError(err, w)
				return
			}

			updateLog("删除成功")
			dialog.ShowInformation("成功", "已删除所有扫描任务", w)
		}, w)
	})

	btnScanExisting := widget.NewButton("扫描已有目标", func() {
		config, err := loadConfig()
		if err != nil {
			dialog.ShowError(fmt.Errorf("加载配置失败: %v", err), w)
			return
		}

		client := awvs.NewClient(config)
		updateLog("正在获取现有目标...")

		targets, err := client.GetTargets()
		if err != nil {
			updateLog(fmt.Sprintf("获取目标失败: %v", err))
			dialog.ShowError(err, w)
			return
		}

		if len(targets) == 0 {
			dialog.ShowInformation("提示", "没有找到任何目标", w)
			return
		}

		// 创建目标列表窗口
		targetWindow := fyne.CurrentApp().NewWindow("选择目标")
		targetWindow.Resize(fyne.NewSize(600, 400))

		var targetItems []string
		var targetIDs []string
		for _, target := range targets {
			address := fmt.Sprintf("%v", target["address"])
			targetID := fmt.Sprintf("%v", target["target_id"])
			targetItems = append(targetItems, address)
			targetIDs = append(targetIDs, targetID)
		}

		list := widget.NewList(
			func() int { return len(targetItems) },
			func() fyne.CanvasObject { return widget.NewLabel("template") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(targetItems[id])
			},
		)

		list.OnSelected = func(id widget.ListItemID) {
			targetID := targetIDs[id]
			selectedType := scanTypeSelect.Selected
			profileID := awvs.ScanTypeMap[selectedType]

			// 添加日志输出
			updateLog(fmt.Sprintf("选择的扫描类型: %s, 对应的 profile_id: %s", selectedType, profileID))

			updateLog(fmt.Sprintf("正在启动扫描: %s", targetItems[id]))
			if err := client.StartScan(targetID, profileID); err != nil {
				updateLog(fmt.Sprintf("启动扫描失败: %v", err))
				dialog.ShowError(err, w)
				return
			}

			updateLog(fmt.Sprintf("成功启动扫描: %s", targetItems[id]))
			dialog.ShowInformation("成功", "已启动扫描", w)
			targetWindow.Close()
		}

		targetWindow.SetContent(list)
		targetWindow.CenterOnScreen()
		targetWindow.Show()
	})

	// 创建免责声明
	disclaimer := widget.NewLabel("免责声明：本工具仅用于安全自查，请勿用于非法测试")
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Alignment = fyne.TextAlignCenter

	// 创建免责声明容器
	disclaimerContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewWithoutLayout(
			container.NewHBox(
				widget.NewLabel("—"),
				disclaimer,
				widget.NewLabel("—"),
			),
		),
		layout.NewSpacer(),
	)

	// 然后再创建布局
	content := container.NewVBox(
		title,
		btnConfig,
		container.NewHBox(widget.NewLabel("扫描类型：")),
		scanTypeSelect,
		container.NewHBox(widget.NewLabel("目标URL：")),
		urlInput,
		btnSelectFile,
		btnAddScan,
		btnDeleteAll,
		btnDeleteTasks,
		btnScanExisting,
		logContainer, // 使用 logContainer 替代之前的日志相关组件
		layout.NewSpacer(),
		disclaimerContainer,
	)

	// 设置窗口最小大小以确保内容显示完整
	w.Resize(fyne.NewSize(800, 900))

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

// 添加新的配置相关函数
func showConfigDialog(parent fyne.Window) {
	w := fyne.CurrentApp().NewWindow("AWVS配置")

	// 基本设置
	apiURLEntry := widget.NewEntry()
	apiURLEntry.SetPlaceHolder("https://your-awvs-host:3443")

	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("apikey")

	// 代理设置 - 用于 AWVS 扫描时的流量代理
	proxyURLEntry := widget.NewEntry()
	proxyURLEntry.SetPlaceHolder("http://127.0.0.1:8080") // 用于配置 AWVS 的扫描代理

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
