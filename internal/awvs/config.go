package awvs

type Config struct {
	APIURL     string `json:"api_url"`
	APIKey     string `json:"api_key"`
	ProxyURL   string `json:"proxy_url,omitempty"`   // 代理地址
	ThreadNum  int    `json:"thread_num,omitempty"`  // 线程数量
	ScanSpeed  string `json:"scan_speed,omitempty"`  // 扫描速度
	ReportType string `json:"report_type,omitempty"` // 报告类型
}

// 扫描配置ID映射
var ScanTypeMap = map[string]string{
	"完全扫描":           "full",
	"扫描高风险漏洞":        "high_risk",
	"扫描XSS漏洞":        "xss",
	"扫描SQL注入漏洞":      "sql_injection",
	"弱口令检测":          "weak_password",
	"仅爬虫(可配合被动扫描)":   "crawl",
	"扫描已知漏洞":         "cve",
	"仅添加目标":          "add_only",
	"Apache-Log4j扫描": "log4j",
	"Bug Bounty高频漏洞": "bug_bounty",
	"常见CVE扫描":        "common_cve",
	"Spring4Shell扫描": "spring4shell",
}

// 扫描速度选项
var SpeedOptions = []string{
	"sequential", // 顺序扫描
	"slow",       // 慢速
	"moderate",   // 中等
	"fast",       // 快速
}

// 报告类型选项
var ReportTypes = []string{
	"HTML",
	"PDF",
	"XML",
}

// 默认配置
var DefaultConfig = Config{
	ThreadNum:  10,
	ScanSpeed:  "moderate",
	ReportType: "HTML",
}
