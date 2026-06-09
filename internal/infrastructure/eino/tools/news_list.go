package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

type NewsListTool struct{ client *resty.Client }

func NewNewsListTool() *NewsListTool {
	return &NewsListTool{client: resty.New().SetTimeout(15 * time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *NewsListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetNewsList",
		Desc: "获取财经新闻资讯。支持来源: 财联社电报(默认)、新浪财经。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keyword": {Type: schema.String, Desc: "关键词或来源", Required: false},
		}),
	}, nil
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max { return s }
	return string([]rune(s)[:max]) + "..."
}

func (t *NewsListTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct { Keyword string `json:"keyword"` }
	json.Unmarshal([]byte(input), &p)

	var result strings.Builder

	// Cailianpress 财联社电报 (default when no keyword)
	if p.Keyword == "" || strings.Contains(p.Keyword, "财联社") || strings.Contains(p.Keyword, "电报") {
		resp, err := t.client.R().SetContext(ctx).
			SetHeader("Referer", "https://www.cls.cn/").
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
			Get("https://www.cls.cn/api/cache?app=CailianpressWeb&name=telegraph&os=web&sv=8.7.9")
		if err == nil {
			var cls struct {
				Errno int `json:"errno"`
				Data  struct { RollData []map[string]any `json:"roll_data"` } `json:"data"`
			}
			if json.Unmarshal(resp.Body(), &cls) == nil && cls.Errno == 0 {
				result.WriteString("## 财联社电报\n\n| 时间 | 内容 |\n|---|---|\n")
				for i, news := range cls.Data.RollData {
					if i >= 15 { break }
					ctime, _ := news["ctime"].(float64)
					tm := time.Unix(int64(ctime), 0).Format("15:04")
					title, _ := news["title"].(string)
					content, _ := news["content"].(string)
					text := title; if text == "" { text = content }
					result.WriteString(fmt.Sprintf("| %s | %s |\n", tm, truncate(text, 80)))
				}
			}
		}
	}

	// Sina 新浪财经
	if strings.Contains(p.Keyword, "新浪") {
		resp, err := t.client.R().SetContext(ctx).
			SetHeader("Referer", "https://finance.sina.com.cn/").
			Get(fmt.Sprintf("https://feed.mix.sina.com.cn/api/roll/get?pageid=153&lid=2516&k=&num=15&page=1"))
		if err == nil {
			var sina struct {
				Result struct { Data []map[string]any `json:"data"` } `json:"result"`
			}
			if json.Unmarshal(resp.Body(), &sina) == nil {
				result.WriteString("\n## 新浪财经\n\n| 标题 | 时间 |\n|---|---|\n")
				for _, item := range sina.Result.Data {
					title, _ := item["title"].(string)
					ct := item["ctime"]
					ctStr := fmt.Sprintf("%v", ct)
					if len(ctStr) >= 16 { ctStr = ctStr[11:16] }
					result.WriteString(fmt.Sprintf("| %s | %s |\n", title, ctStr))
				}
			}
		}
	}

	if result.Len() == 0 { result.WriteString("暂无新闻数据") }
	return result.String(), nil
}
var _ tool.InvokableTool = (*NewsListTool)(nil)
