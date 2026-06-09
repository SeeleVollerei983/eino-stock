export class SettingConfig {
  constructor(source = {}) {
    Object.assign(this, { ID:0, DarkTheme:false, OpenAiEnable:true, EnableFund:false, EnableAgent:true, EnableNews:false, HttpProxyEnabled:false, CrawlTimeOut:30, ...source })
  }
  static createFrom(source) { return new SettingConfig(source) }
}
export class AIConfig {
  constructor(source = {}) {
    Object.assign(this, { ID:1, Name:"default", Model:"deepseek-chat", ApiKey:"", BaseUrl:"https://api.deepseek.com", ...source })
  }
}
export const data = { SettingConfig, AIConfig }
export const models = {}
