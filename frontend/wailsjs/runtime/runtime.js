export function EventsOn() { return ()=>{} }
export function EventsEmit() {}
export function EventsOff() {}
export function WindowFullscreen() {}
export function WindowUnfullscreen() {}
export function WindowSetTitle(t) { document.title=t }
export function WindowReload() { location.reload() }
export function Environment() { return Promise.resolve({platform:"windows"}) }
export function OpenURL(u) { window.open(u,"_blank") }
export function BrowserOpenURL(url) { window.open(url, '_blank') }
