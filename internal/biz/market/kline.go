package market

import "context"

type KLine struct {
	Day           string            `json:"day"`
	Open          string            `json:"open"`
	Close         string            `json:"close"`
	High          string            `json:"high"`
	Low           string            `json:"low"`
	Volume        string            `json:"volume"`
	Amount        string            `json:"amount"`
	ChangePercent string            `json:"changePercent"`
	ChangeValue   string            `json:"changeValue"`
	Amplitude     string            `json:"amplitude"`
	TurnoverRate  string            `json:"turnoverRate"`
	MA            map[string]string `json:"ma,omitempty"`
}

type KLineType string

const (
	KLine1Min     KLineType = "1"
	KLine5Min     KLineType = "5"
	KLine15Min    KLineType = "15"
	KLine30Min    KLineType = "30"
	KLine60Min    KLineType = "60"
	KLine120Min   KLineType = "120"
	KLineDay      KLineType = "101"
	KLineWeek     KLineType = "102"
	KLineMonth    KLineType = "103"
	KLineQuarter  KLineType = "104"
	KLineYear     KLineType = "106"
)

type KLineProvider interface {
	GetKLines(ctx context.Context, code string, ktype KLineType, limit int) ([]*KLine, error)
}