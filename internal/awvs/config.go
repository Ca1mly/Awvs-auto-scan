package awvs

type Config struct {
	APIURL       string `json:"api_url"`
	APIKey       string `json:"api_key"`
	ProxyEnabled string `json:"proxy_enabled,omitempty"` // 代理开关
	ProxyIP      string `json:"ip,omitempty"`
	ProxyPort    string `json:"port,omitempty"`        // 代理地址
	ThreadNum    int    `json:"thread_num,omitempty"`  // 线程数量
	ScanSpeed    string `json:"scan_speed,omitempty"`  // 扫描速度
	ReportType   string `json:"report_type,omitempty"` // 报告类型
}

// 扫描配置ID映射
var ScanTypeMap = map[string]string{
	"完全扫描":           "11111111-1111-1111-1111-111111111111",
	"扫描高风险漏洞":        "11111111-1111-1111-1111-111111111112",
	"扫描XSS漏洞":        "11111111-1111-1111-1111-111111111116",
	"扫描SQL注入漏洞":      "11111111-1111-1111-1111-111111111113",
	"弱口令检测":          "11111111-1111-1111-1111-111111111115",
	"仅爬虫(可配合被动扫描)":   "11111111-1111-1111-1111-111111111117",
	"扫描已知漏洞":         "11111111-1111-1111-1111-111111111113",
	"仅添加目标":          "11111111-1111-1111-1111-111111111120",
	"Apache-Log4j扫描": "11111111-1111-1111-1111-111111111114",
	"Bug Bounty高频漏洞": "11111111-1111-1111-1111-111111111118",
	"常见CVE扫描":        "11111111-1111-1111-1111-111111111121",
	"Spring4Shell扫描": "11111111-1111-1111-1111-111111111122",
}

// 实际的 AWVS 扫描模板 ID
const (
	FullScan      = "11111111-1111-1111-1111-111111111111" // Full Scan
	HighRisk      = "11111111-1111-1111-1111-111111111112" // High Risk Vulnerabilities
	XSSScan       = "11111111-1111-1111-1111-111111111116" // Cross-site Scripting
	SQLInjection  = "11111111-1111-1111-1111-111111111113" // SQL Injection
	WeakPasswords = "11111111-1111-1111-1111-111111111115" // Weak Passwords
	CrawlOnly     = "11111111-1111-1111-1111-111111111117" // Crawl Only
	MalwareEval   = "11111111-1111-1111-1111-111111111113" // Known Vulnerabilities
	AddTargetOnly = "11111111-1111-1111-1111-111111111120" // Add Target Only
	Log4j         = "11111111-1111-1111-1111-111111111114" // Log4j
	BugBounty     = "11111111-1111-1111-1111-111111111118" // Bug Bounty
	CommonCVE     = "11111111-1111-1111-1111-111111111121" // Common CVEs
	Spring4Shell  = "11111111-1111-1111-1111-111111111122" // Spring4Shell
)

// 扫描速度选项
var SpeedOptions = []string{
	"sequential", // 顺序扫描
	"slow",       // 慢速
	"moderate",   // 中等
	"fast",       // 快速
}

// ProxyEnabled
var ProxyEnabled = []string{
	"False", // false
	"True",  // true
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

// 扫描配置描述
var ScanTypeDescriptions = map[string]string{
	"完全扫描":           "执行完整的漏洞扫描，包括所有类型的漏洞检测",
	"扫描高风险漏洞":        "仅扫描高风险漏洞，如命令注入、文件包含等",
	"扫描XSS漏洞":        "专门扫描跨站脚本攻击漏洞",
	"扫描SQL注入漏洞":      "专门扫描SQL注入漏洞",
	"弱口令检测":          "检测常见的弱密码问题",
	"仅爬虫(可配合被动扫描)":   "只执行爬虫功能，不进行漏洞扫描",
	"扫描已知漏洞":         "扫描已知的CVE漏洞",
	"仅添加目标":          "只添加目标，不执行扫描",
	"Apache-Log4j扫描": "专门扫描Log4j相关漏洞",
	"Bug Bounty高频漏洞": "扫描漏洞赏金计划中常见的漏洞",
	"常见CVE扫描":        "扫描常见的CVE漏洞",
	"Spring4Shell扫描": "专门扫描Spring4Shell漏洞",
}
