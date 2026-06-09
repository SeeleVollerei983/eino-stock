package market

import "strings"

// NormalizeStockCodes 标准化股票代码格式。
func NormalizeStockCodes(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		code = strings.ToLower(code)
		// 去掉 . 后缀
		code = strings.ReplaceAll(code, ".sh", "")
		code = strings.ReplaceAll(code, ".sz", "")
		code = strings.ReplaceAll(code, ".bj", "")
		// 已经带前缀
		if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "bj") {
			out = append(out, code)
			continue
		}
		// 纯数字
		if len(code) >= 6 && code[0] >= '0' && code[0] <= '9' {
			switch code[0] {
			case '6':
				out = append(out, "sh"+code)
			case '0', '3':
				out = append(out, "sz"+code)
			case '8', '9':
				out = append(out, "bj"+code)
			default:
				out = append(out, code)
			}
		} else {
			out = append(out, code)
		}
	}
	return out
}
