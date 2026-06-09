package f10

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

const emF10BaseURL = "https://datacenter.eastmoney.com/securities/api/data/v1/get"

type GenericResp struct {
	Version string   `json:"version"`
	Result  *F10Data `json:"result"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Code    int      `json:"code"`
}

type F10Data struct {
	Count int              `json:"count"`
	Data  []map[string]any `json:"data"`
}

type Client struct {
	http *resty.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		http: resty.New().SetTimeout(timeout).
			SetHeader("Host", "datacenter.eastmoney.com").
			SetHeader("Referer", "https://emweb.securities.eastmoney.com/").
			SetHeader("Origin", "https://emweb.securities.eastmoney.com").
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:148.0) Gecko/20100101 Firefox/148.0"),
	}
}

func (c *Client) request(url string, result any) error {
	resp, err := c.http.R().Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode())
	}
	return json.Unmarshal(resp.Body(), result)
}

func normalizeCode(code string) string {
	if len(code) == 0 {
		return code
	}
	// If already has dot suffix (e.g., 600519.SH), return as-is
	for _, sep := range []string{".SH", ".SZ", ".BJ", ".sh", ".sz", ".bj"} {
		for i := 0; i < len(code)-3; i++ {
			if code[i:i+3] == sep {
				return code
			}
		}
	}
	// Strip prefix
	if len(code) > 6 {
		code = code[len(code)-6:]
	}
	// Pure number
	switch code[0] {
	case '6', '9':
		return code + ".SH"
	case '0', '3', '4', '8':
		return code + ".SZ"
	case '5':
		return code + ".SH"
	}
	return code + ".SZ"
}

func buildURL(reportName, columns, filter string) string {
	v := url.Values{}
	v.Set("reportName", reportName)
	v.Set("columns", columns)
	v.Set("filter", filter)
	v.Set("pageNumber", "1")
	v.Set("pageSize", "10")
	v.Set("sortTypes", "-1")
	v.Set("sortColumns", "REPORT_DATE")
	v.Set("source", "HSF10")
	v.Set("client", "PC")
	v.Set("v", fmt.Sprintf("%d", time.Now().Unix()))
	return emF10BaseURL + "?" + v.Encode()
}

func (c *Client) LatestFinance(stockCode string) (*GenericResp, error) {
	code := normalizeCode(stockCode)
	cols := "SECUCODE,SECURITY_CODE,SECURITY_NAME_ABBR,REPORT_DATE,REPORT_TYPE,EPSJB,EPSKCJB,EPSXS,BPS,MGZBGJ,MGWFPLR,MGJYXJJE,TOTAL_OPERATEINCOME,TOTAL_OPERATEINCOME_LAST,PARENT_NETPROFIT,PARENT_NETPROFIT_LAST,KCFJCXSYJLR,KCFJCXSYJLR_LAST,ROEJQ,ROEJQ_LAST,XSMLL,XSMLL_LAST,ZCFZL,ZCFZL_LAST,YYZSRGDHBZC,YYZSRGDHBZC_LAST,NETPROFITRPHBZC,NETPROFITRPHBZC_LAST,KFJLRGDHBZC,KFJLRGDHBZC_LAST,TOTALOPERATEREVETZ,TOTALOPERATEREVETZ_LAST,PARENTNETPROFITTZ,PARENTNETPROFITTZ_LAST,KCFJCXSYJLRTZ,KCFJCXSYJLRTZ_LAST,TOTAL_SHARE,FREE_SHARE,EPSJB_PL,BPS_PL,FORMERNAME"
	filter := fmt.Sprintf("(SECUCODE=%q)", code)
	url := buildURL("RPT_PCF10_FINANCEMAINFINADATA", cols, filter)
	var resp GenericResp
	err := c.request(url, &resp)
	return &resp, err
}

func (c *Client) QtrFinance(stockCode string) (*GenericResp, error) {
	code := normalizeCode(stockCode)
	cols := "SECUCODE,SECURITY_CODE,SECURITY_NAME_ABBR,ORG_CODE,REPORT_DATE,EPSJB,BPS,PER_CAPITAL_RESERVE,PER_UNASSIGN_PROFIT,PER_NETCASH,TOTALOPERATEREVE,GROSS_PROFIT,PARENTNETPROFIT,DEDU_PARENT_PROFIT,TOTALOPERATEREVETZ,PARENTNETPROFITTZ,DPNP_YOY_RATIO,YYZSRGDHBZC,NETPROFITRPHBZC,KFJLRGDHBZC,ROE_DILUTED,JROA,NET_PROFIT_RATIO,GROSS_PROFIT_RATIO"
	filter := fmt.Sprintf("(SECUCODE=%q)", code)
	url := buildURL("RPT_F10_QTR_MAINFINADATA", cols, filter)
	var resp GenericResp
	err := c.request(url, &resp)
	return &resp, err
}

func (c *Client) HolderTrend(stockCode string) (*GenericResp, error) {
	code := normalizeCode(stockCode)
	filter := fmt.Sprintf("(SECUCODE=%q)(INDICATORTYPE=1)(DATETYPE=3)", code)
	url := emF10BaseURL + "?reportName=RPT_CUSTOM_DMSK_TREND&columns=ALL&quoteColumns=&filter=" + url.QueryEscape(filter) + "&pageNumber=1&pageSize=&sortTypes=1&sortColumns=TRADE_DATE&source=HSF10&client=PC&v=" + fmt.Sprintf("%d", time.Now().Unix())
	var resp GenericResp
	err := c.request(url, &resp)
	return &resp, err
}

func (c *Client) OrgPredict(stockCode string) (*GenericResp, error) {
	code := normalizeCode(stockCode)
	cols := "SECUCODE,SECURITY_CODE,SECURITY_NAME_ABBR,PUBLISH_DATE,ORG_CODE,ORG_NAME_ABBR,YEAR1,YEAR_MARK1,EPS1,PE1,YEAR2,YEAR_MARK2,EPS2,PE2,YEAR3,YEAR_MARK3,EPS3,PE3,YEAR4,YEAR_MARK4,EPS4,PE4"
	filter := fmt.Sprintf("(SECUCODE=%q)", code)
	url := buildURL("RPT_HSF10_RES_ORGPREDICT", cols, filter)
	var resp GenericResp
	err := c.request(url, &resp)
	return &resp, err
}

func (c *Client) PredictSummary(stockCode string) (*GenericResp, error) {
	code := normalizeCode(stockCode)
	cols := "SECUCODE,SECURITY_CODE,SECURITY_NAME_ABBR,YEAR,YEAR_MARK,EPS,EPS_RATIO,PE"
	filter := fmt.Sprintf("(SECUCODE=%q)", code)
	url := emF10BaseURL + "?reportName=RPT_HSF10_RESPREDICT_STATISTICS&columns=" + url.QueryEscape(cols) + "&filter=" + url.QueryEscape(filter) + "&pageNumber=1&pageSize=200&sortTypes=1&sortColumns=RANK&source=HSF10&client=PC&v=" + fmt.Sprintf("%d", time.Now().Unix())
	var resp GenericResp
	err := c.request(url, &resp)
	return &resp, err
}
