package tools

import (
	"github.com/cloudwego/eino/components/tool"
)

func GetAllTools() []tool.BaseTool {
	return []tool.BaseTool{
		NewSelectAStockTool(),
		NewMinuteDataTool(),
		NewStockDetailTool(),
		NewStockNoticeTool(),
		NewResearchReportTool(),
		NewGlobalStockIndexesTool(),
		NewMarketDataTool(),
		NewIndustryValuationTool(),
		NewIndustryMoneyRankTool(),
		NewUplimitHotPlatesTool(),
		NewLongTigerListTool(),
		NewQueryEconomicDataTool(),
		NewNewsListTool(),
		NewMutualTop10Tool(),
		NewWebSearchTool(),
	NewExpertAnalysisTool(),
	}
}

func ToolNames() []string {
	var names []string
	for _, t := range GetAllTools() {
		info, _ := t.Info(nil)
		if info != nil {
			names = append(names, info.Name)
		}
	}
	return names
}

