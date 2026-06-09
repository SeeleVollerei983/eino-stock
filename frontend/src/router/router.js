import {createMemoryHistory, createRouter, createWebHashHistory} from 'vue-router'

import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'
import marketView from '../components/market.vue'
import agentChat from '../components/agent-chat.vue'
import research from '../components/researchIndex.vue'
import klineAnalysis from '../components/kline-analysis.vue'

const routes = [
    { path: '/', component: stockView, name: 'stock'},
    { path: '/settings', component: settingsView, name: 'settings' },
    { path: '/market', component: marketView, name: 'market' },
    { path: '/agent', component: agentChat, name: 'agent' },
    { path: '/research', component: research, name: 'research' },
    { path: '/kline-analysis', component: klineAnalysis, name: 'klineAnalysis' },
]

const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

export default router
