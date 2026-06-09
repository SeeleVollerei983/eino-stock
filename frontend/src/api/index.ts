const BASE = "/api"

async function get(url: string) {
  const r = await fetch(BASE + url)
  if (!r.ok) throw new Error(`HTTP ${r.status}`)
  return r.json()
}
async function post(url: string) {
  const r = await fetch(BASE + url, { method: "POST" })
  if (!r.ok) throw new Error(`HTTP ${r.status}`)
  return r.json()
}

export const api = {
  // Stock basic
  searchStocks: (keyword: string, limit = 20) => get(`/market/stocks?keyword=${encodeURIComponent(keyword)}&limit=${limit}`),
  quote: (code: string) => get(`/market/quote/${code}`),
  kline: (code: string, ktype = "101", limit = 120) => get(`/market/kline/${code}?ktype=${ktype}&limit=${limit}`),
  minute: (code: string) => get(`/tool/minute?code=${code}`),
  detail: (code: string) => get(`/tool/detail?code=${code}`),
  notice: (code: string) => get(`/tool/notice?code=${code}`),
  report: (code: string, days = 30) => get(`/tool/report?code=${code}&days=${days}`),

  // Market overview
  globalIndexes: () => get("/tool/global-indexes"),
  hotPlates: () => get("/tool/hot-plates"),
  industryMoneyRank: () => get("/tool/industry-money-rank"),
  industryValuation: () => get("/tool/industry-valuation?bkName="),
  longTiger: () => get("/tool/long-tiger"),
  hotStrategy: () => get("/screen/hot-strategy"),

  // News & economic data
  newsList: (keyword?: string) => get(`/tool/news-list${keyword ? "?keyword="+encodeURIComponent(keyword) : ""}`),
  economicData: (flag = "all") => get(`/tool/economic-data?flag=${flag}`),
  mutualTop10: (mutualType = "001", tradeDate?: string) => {
    let p = `?mutualType=${mutualType}`
    if (tradeDate) p += "&tradeDate=" + tradeDate
    return get("/tool/mutual-top10" + p)
  },

  // Stock screening
  screen: (q: string) => get(`/tool/screen?q=${encodeURIComponent(q)}`),
  screenV2: (q: string) => get(`/tool/screen-v2?q=${encodeURIComponent(q)}`),
  aiScreen: (q: string) => get(`/ai/screen?q=${encodeURIComponent(q)}`),
  aiParallel: (q: string) => get(`/ai/parallel?q=${encodeURIComponent(q)}`),
  searchBk: (keyword: string) => get(`/screen/bk/${encodeURIComponent(keyword)}`),
  searchEtf: (keyword: string) => get(`/screen/etf/${encodeURIComponent(keyword)}`),

  // Follow / watchlist
  followList: () => get("/follow/list"),
  followAdd: (code: string, name: string) => post(`/follow/add?code=${code}&name=${encodeURIComponent(name)}`),
  followRemove: (code: string) => post(`/follow/remove?code=${code}`),

  // Settings
  settingsGet: () => get("/settings/get"),
  settingsSet: (key: string, value: string) => get(`/settings/set?key=${encodeURIComponent(key)}&value=${encodeURIComponent(value)}`),

  // Market search
  bkSearch: (keyword: string, limit = 20) => get(`/screen/bk/${encodeURIComponent(keyword)}?limit=${limit}`),
  etfSearch: (keyword: string, limit = 20) => get(`/screen/etf/${encodeURIComponent(keyword)}?limit=${limit}`),

  // F10 (coming soon)
  f10Profile: (code: string) => get(`/f10/profile?code=${code}`),
  f10Finance: (code: string) => get(`/f10/finance?code=${code}`),
  f10Holders: (code: string) => get(`/f10/holders?code=${code}`),
}
