<script setup>
import { h, ref, onMounted } from "vue";
import { RouterLink, useRouter } from 'vue-router'
import { NIcon, darkTheme } from 'naive-ui'
import {
  StarOutline, NewspaperOutline, StatsChartOutline,
  SettingsOutline, ChatbubblesOutline, FlaskOutline,
} from '@vicons/ionicons5'
import { api } from './api'

const router = useRouter()
const enableDarkTheme = ref(null)
const activeKey = ref('stock')

function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = [
  { label: () => h(RouterLink, { to: { name: 'stock' } }, { default: () => '自选' }), key: 'stock', icon: renderIcon(StarOutline) },
  { label: () => h(RouterLink, { to: { name: 'market' } }, { default: () => '市场' }), key: 'market', icon: renderIcon(NewspaperOutline) },
  { label: () => h(RouterLink, { to: { name: 'klineAnalysis' } }, { default: () => 'K线' }), key: 'klineAnalysis', icon: renderIcon(StatsChartOutline) },
  { label: () => h(RouterLink, { to: { name: 'agent' } }, { default: () => 'AI' }), key: 'agent', icon: renderIcon(ChatbubblesOutline) },
  { label: () => h(RouterLink, { to: { name: 'research' } }, { default: () => '研究' }), key: 'research', icon: renderIcon(FlaskOutline) },
  { label: () => h(RouterLink, { to: { name: 'settings' } }, { default: () => '设置' }), key: 'settings', icon: renderIcon(SettingsOutline) },
]

onMounted(async () => {
  try {
    const s = await api.settingsGet()
    if (s?.darkTheme) enableDarkTheme.value = darkTheme
  } catch {}
})
</script>
<template>
  <n-config-provider :theme="enableDarkTheme">
    <n-message-provider>
      <n-notification-provider>
        <n-modal-provider>
          <n-dialog-provider>
            <n-layout has-sider style="height:100vh">
              <n-layout-sider bordered width=200 content-style="padding:0;">
                <n-menu v-model:value="activeKey" :options="menuOptions" />
              </n-layout-sider>
              <n-layout>
                <n-layout-content style="padding:16px;height:100vh;overflow-y:auto;">
                  <router-view />
                </n-layout-content>
              </n-layout>
            </n-layout>
          </n-dialog-provider>
        </n-modal-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>
