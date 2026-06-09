<script setup lang="ts">
import {onBeforeUnmount, onMounted, ref, watch, computed} from "vue";
import {GetBKFundFlowListByDate, GetBKFundFlowTopListByDate, GetAllBKCodes} from "../../wailsjs/go/main/App";
import echarts from "echarts";

const props = defineProps({
  darkTheme: {
    type: Boolean,
    default: false
  },
  chartHeight: {
    type: Number,
    default: 600
  }
})

const chartRef = ref(null)
const topList = ref<any[]>([])
const loading = ref(false)
const refreshInterval = ref<any>(null)
let chart: echarts.ECharts | null = null

// 涓存椂娣诲姞鐨勬澘鍧?
const extraSectors = ref<any[]>([])
const addCodeInput = ref('')
const allBKCodes = ref<any[]>([])
const addCodeOptions = computed(() => {
  const existCodes = new Set([
    ...inflowList.value.map((i: any) => i.code),
    ...outflowList.value.map((i: any) => i.code),
    ...extraSectors.value.map((i: any) => i.code)
  ])
  return allBKCodes.value
    .filter((i: any) => !existCodes.has(i.code))
    .map((i: any) => ({label: `${i.name} (${i.code})`, value: i.code, name: i.name}))
})

// 榛樿褰撳ぉ鏃ユ湡锛屾牸寮?YYYY-MM-DD
const today = new Date()
const todayStr = today.getFullYear() + '-' +
  String(today.getMonth() + 1).padStart(2, '0') + '-' +
  String(today.getDate()).padStart(2, '0')
const selectedDate = ref(todayStr)

// 鏄惁鏌ョ湅鐨勬槸浠婂ぉ
const isToday = computed(() => selectedDate.value === todayStr)

// 娴佸叆鍓?0
const inflowList = computed(() => topList.value.filter((item: any) => item.netInflow > 0).slice(0, 20))
// 娴佸嚭鍓?0
const outflowList = computed(() => topList.value.filter((item: any) => item.netInflow < 0).slice(-20).reverse())

onMounted(async () => {
  try {
    // 鍔犺浇鎵€鏈夋澘鍧椾唬鐮侊紙渚涗复鏃舵坊鍔犱娇鐢級
    const codes = await GetAllBKCodes()
    if (codes && Array.isArray(codes)) allBKCodes.value = codes

    await loadAllData()
    // 浜ゆ槗鏃堕棿姣忓垎閽熷埛鏂帮紙浠呭綋澶╋級
    refreshInterval.value = setInterval(async () => {
      if (isToday.value && isTradingTime()) {
        await loadAllData()
      }
    }, 60000)
  } catch (e) {
    console.error('onMounted error:', e)
  }
})

onBeforeUnmount(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
  if (chart) {
    chart.dispose()
  }
})

function isTradingTime(): boolean {
  const now = new Date()
  const h = now.getHours()
  const m = now.getMinutes()
  const t = h * 60 + m
  // 9:30 - 11:30 鎴?13:00 - 15:00
  return (t >= 570 && t <= 690) || (t >= 780 && t <= 900)
}

// 娴佸叆鍓?0鐢ㄧ孩鑹茬郴锛屾祦鍑哄墠20鐢ㄧ豢鑹茬郴
function getSeriesColor(index: number, isInflow: boolean): string {
  const redShades = [
    '#ee6666', '#d14a61', '#fc8452', '#e8534e', '#c23531',
    '#f4755e', '#d4816a', '#e76f51', '#ef6c4a', '#d9534f',
    '#e74c3c', '#ff6b6b', '#ee5a24', '#f19066', '#e55039',
    '#eb4d4b', '#fc5c65', '#ff6348', '#e71d36', '#c0392b'
  ]
  const greenShades = [
    '#00da3c', '#3ba272', '#91cc75', '#2ecc71', '#27ae60',
    '#52be80', '#58d68d', '#45b39d', '#1abc9c', '#16a085',
    '#239b56', '#28b463', '#73c0de', '#48b8d0', '#87cefa',
    '#1dd1a1', '#10ac84', '#0abde3', '#01a3a4', '#00cec9'
  ]
  return isInflow ? redShades[index % redShades.length] : greenShades[index % greenShades.length]
}

// 鏃ユ湡绂佺敤锛氫笉鑳介€夋湭鏉ユ棩鏈?
function isDateDisabled(ts: number): boolean {
  return ts > Date.now()
}

async function loadAllData() {
  loading.value = true
  try {
    const date = selectedDate.value
    // 鎸夋棩鏈熻幏鍙栨澘鍧楁帓鍚?
    const res = await GetBKFundFlowTopListByDate(date, 500)
    if (!res || !Array.isArray(res) || res.length === 0) {
      topList.value = []
      return
    }
    topList.value = res

    // 娴佸叆鍓?0 鍜?娴佸嚭鍓?0
    const inflowTop = res.filter((item: any) => item.netInflow > 0).slice(0, 20)
    const outflowTop = res.filter((item: any) => item.netInflow < 0).slice(-20).reverse()
    const sectors = [...inflowTop, ...outflowTop, ...extraSectors.value]

    const allData = await Promise.all(
      sectors.map(async (item: any) => {
        const points = await GetBKFundFlowListByDate(item.code, date)
        return {
          code: item.code,
          name: item.name,
          isInflow: item.isInflow !== undefined ? item.isInflow : item.netInflow > 0,
          points: points || []
        }
      })
    )

    renderChart(allData)
  } catch (e) {
    console.error('loadAllData error:', e)
  } finally {
    loading.value = false
  }
}

function renderChart(allData: { code: string; name: string; isInflow: boolean; points: any[] }[]) {
  if (!allData || allData.length === 0) return

  // 鏀堕泦鎵€鏈夋椂闂寸偣锛屽幓閲嶅苟鎺掑簭
  const timeSet = new Set<string>()
  for (const sector of allData) {
    for (const pt of sector.points) {
      const t = extractTime(pt.snapTime)
      if (t) timeSet.add(t)
    }
  }
  const times = Array.from(timeSet).sort()

  // 鍒嗗埆缁熻娴佸叆/娴佸嚭鏉垮潡鐨勭储寮曪紝鐢ㄤ簬閰嶈壊
  let inflowIdx = 0
  let outflowIdx = 0

  // 鏋勫缓姣忎釜鏉垮潡鐨?series 鏁版嵁锛屽榻愬埌缁熶竴鏃堕棿杞?
  const seriesList: any[] = []
  for (let i = 0; i < allData.length; i++) {
    const sector = allData[i]
    const dataMap = new Map<string, number>()
    for (const pt of sector.points) {
      const t = extractTime(pt.snapTime)
      if (t) dataMap.set(t, pt.netInflow || 0)
    }

    // 瀵归綈鍒扮粺涓€鏃堕棿杞达紝缂哄け鐨勬椂闂寸偣鐢?null
    const values = times.map(t => dataMap.has(t) ? dataMap.get(t)! : null)

    const color = getSeriesColor(sector.isInflow ? inflowIdx++ : outflowIdx++, sector.isInflow)

    seriesList.push({
      name: sector.name,
      type: 'line',
      data: values,
      smooth: true,
      showSymbol: false,
      lineStyle: {width: 2, color},
      itemStyle: {color},
      emphasis: {
        lineStyle: {width: 3},
        focus: 'series'
      },
      // 鎶樼嚎鏈熬鏄剧ず鍚嶇О
      endLabel: {
        show: true,
        formatter: '{a}',
        color,
        fontSize: 12,
        fontWeight: 'bold',
        distance: 8
      }
    })
  }

  if (!chart && chartRef.value) {
    chart = echarts.init(chartRef.value)
  }
  if (!chart) return

  const textColor = props.darkTheme ? '#aaa' : '#666'
  const bgColor = props.darkTheme ? '#1a1a2e' : '#fff'
  const dateLabel = selectedDate.value

  const option: echarts.EChartsOption = {
    backgroundColor: bgColor,
    title: {
      text: `${dateLabel} 鏉垮潡璧勯噾娴佸悜 - 澶氭澘鍧楀姣擿,
      left: '20px',
      textStyle: {color: props.darkTheme ? '#ccc' : '#456', fontSize: 16}
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {type: 'cross'},
      borderWidth: 1,
      borderColor: props.darkTheme ? '#456' : '#ddd',
      backgroundColor: props.darkTheme ? 'rgba(30,30,60,0.9)' : 'rgba(255,255,255,0.95)',
      padding: 10,
      textStyle: {color: props.darkTheme ? '#ccc' : '#333', fontSize: 12},
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ''
        const inflowItems = params.filter((p: any) => p.value != null && p.value > 0)
          .sort((a: any, b: any) => (b.value || 0) - (a.value || 0))
        const outflowItems = params.filter((p: any) => p.value != null && p.value < 0)
          .sort((a: any, b: any) => (a.value || 0) - (b.value || 0))

        const renderList = (items: any[]) => items.map((p: any) => {
          const val = (p.value / 100000000).toFixed(2)
          const sign = p.value > 0 ? '+' : ''
          return `${p.marker} ${p.seriesName} <b>${sign}${val}</b>`
        }).join('<br/>')

        let html = `<b>${params[0].axisValue}</b><br/>`
        html += '<div style="display:flex;gap:20px;">'
        html += `<div><div style="color:#ee6666;font-weight:bold;margin-bottom:4px;">娴佸叆</div>${renderList(inflowItems) || '-'}</div>`
        html += `<div><div style="color:#00da3c;font-weight:bold;margin-bottom:4px;">娴佸嚭</div>${renderList(outflowItems) || '-'}</div>`
        html += '</div>'
        return html
      }
    },
    legend: {
      type: 'plain',
      left: 0,
      top: 30,
      orient: 'horizontal',
      align: 'left',
      itemWidth: 14,
      itemHeight: 10,
      itemGap: 8,
      textStyle: {color: textColor, fontSize: 11},
      icon: 'roundRect'
    },
    grid: {
      left: '8%',
      right: '12%',
      top: 120,
      height: '52%'
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLine: {onZero: false, lineStyle: {color: props.darkTheme ? '#444' : '#ccc'}},
      splitLine: {show: false},
      axisLabel: {
        color: textColor,
        rotate: 30,
        fontSize: 11,
        interval: times.length <= 30 ? 0 : Math.floor(times.length / 12)
      },
      axisTick: {lineStyle: {color: props.darkTheme ? '#444' : '#ccc'}}
    },
    yAxis: {
      name: '鍑€娴佸叆/浜垮厓',
      type: 'value',
      nameTextStyle: {color: textColor, fontSize: 11},
      axisLine: {show: true, lineStyle: {color: props.darkTheme ? '#444' : '#ccc'}},
      splitLine: {lineStyle: {color: props.darkTheme ? '#333' : '#eee', type: 'dashed'}},
      axisLabel: {
        color: textColor,
        fontSize: 11,
        formatter: (v: number) => (v / 100000000).toFixed(2)
      }
    },
    series: seriesList,
    dataZoom: [
      {
        type: 'inside',
        xAxisIndex: [0],
        start: 0,
        end: 100
      },
      {
        show: true,
        xAxisIndex: [0],
        type: 'slider',
        top: '88%',
        start: 0,
        end: 100,
        borderColor: props.darkTheme ? '#444' : '#ccc',
        fillerColor: props.darkTheme ? 'rgba(100,100,200,0.2)' : 'rgba(100,100,200,0.15)',
        handleStyle: {color: props.darkTheme ? '#666' : '#999'},
        textStyle: {color: textColor}
      }
    ]
  }

  chart.setOption(option, true)
  chart.resize()
}

function extractTime(snapTime: string): string {
  if (!snapTime || typeof snapTime !== 'string') return ''
  if (snapTime.length >= 16) return snapTime.substring(11, 16)
  return String(snapTime)
}

function addExtraSector(code: string) {
  if (!code) return
  const found = allBKCodes.value.find((i: any) => i.code === code)
  if (!found) return
  // 閬垮厤閲嶅
  if (extraSectors.value.some((i: any) => i.code === code)) return
  // 鍙帓闄ゅ凡鏄剧ず鍦ㄥ浘涓婄殑锛堟祦鍏ュ墠20+娴佸嚭鍓?0锛?
  const displayCodes = new Set([
    ...inflowList.value.map((i: any) => i.code),
    ...outflowList.value.map((i: any) => i.code)
  ])
  if (displayCodes.has(code)) {
    // 宸插湪鍥句腑鏄剧ず锛屾棤闇€閲嶅娣诲姞
    addCodeInput.value = ''
    return
  }
  extraSectors.value.push({code: found.code, name: found.name, netInflow: 0, isInflow: true})
  addCodeInput.value = ''
  loadAllData()
}

function removeExtraSector(code: string) {
  extraSectors.value = extraSectors.value.filter((i: any) => i.code !== code)
  loadAllData()
}

function onDateChange() {
  loadAllData()
}

watch(() => props.darkTheme, () => {
  loadAllData()
})

watch(() => props.chartHeight, () => {
  if (chart) chart.resize()
})
</script>

<template>
  <div style="width: 100%">
    <!-- 鎺у埗鏍?-->
    <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px; flex-wrap: wrap;">
      <n-tag :bordered="false" type="error" size="small">绾㈣壊绯?= 娴佸叆鍓?0</n-tag>
      <n-tag :bordered="false" type="success" size="small">缁胯壊绯?= 娴佸嚭鍓?0</n-tag>
      <n-date-picker
          v-model:formatted-value="selectedDate"
          type="date"
          value-format="yyyy-MM-dd"
          :is-date-disabled="isDateDisabled"
          style="width: 150px"
          @update:formatted-value="onDateChange"
      />
      <n-button size="small" :loading="loading" @click="loadAllData">
        鍒锋柊
      </n-button>
      <n-select
          v-model:value="addCodeInput"
          :options="addCodeOptions"
          filterable
          placeholder="娣诲姞鏉垮潡"
          style="width: 180px"
          @update:value="addExtraSector"
      />
      <n-tag
          v-for="item in extraSectors"
          :key="item.code"
          closable
          size="small"
          type="warning"
          @close="removeExtraSector(item.code)"
      >
        {{ item.name }}
      </n-tag>
      <n-text v-if="isToday" :depth="3" style="font-size: 12px; margin-left: auto;">
        浜ゆ槗鏃堕棿鑷姩姣忓垎閽熷埛鏂?(9:30-15:00)
      </n-text>
      <n-text v-else depth="3" style="font-size: 12px; margin-left: auto;">
        鍘嗗彶鏁版嵁
      </n-text>
    </div>

    <!-- 鎶樼嚎鍥?-->
    <div ref="chartRef" style="width: 100%;" :style="{height: chartHeight + 'px'}"></div>

    <!-- 鏉垮潡璧勯噾鎺掑悕琛ㄦ牸 - 骞舵帓灞曠ず -->
    <div style="margin-top: 20px; display: flex; gap: 20px; align-items: flex-start;">
      <!-- 娴佸叆鎺掑悕 -->
      <div style="flex: 1; min-width: 0;">
        <n-h3 :style="{color: '#ee6666'}">娴佸叆 Top 20</n-h3>
        <n-table striped size="small">
          <n-thead>
            <n-tr>
              <n-th>鎺掑悕</n-th>
              <n-th>鏉垮潡鍚嶇О</n-th>
              <n-th>鍑€娴佸叆/浜垮厓</n-th>
            </n-tr>
          </n-thead>
          <n-tbody>
            <n-tr v-for="(item, idx) in inflowList" :key="item.code">
              <n-td>{{ idx + 1 }}</n-td>
              <n-td>
                <n-tag :bordered="false" type="error" size="small">{{ item.name }}</n-tag>
              </n-td>
              <n-td>
                <n-text type="error">+{{ (item.netInflow / 100000000).toFixed(2) }}</n-text>
              </n-td>
            </n-tr>
            <n-tr v-if="inflowList.length === 0">
              <n-td colspan="3" style="text-align: center; color: #999;">鏆傛棤鏁版嵁</n-td>
            </n-tr>
          </n-tbody>
        </n-table>
      </div>
      <!-- 娴佸嚭鎺掑悕 -->
      <div style="flex: 1; min-width: 0;">
        <n-h3 :style="{color: '#00da3c'}">娴佸嚭 Top 20</n-h3>
        <n-table striped size="small">
          <n-thead>
            <n-tr>
              <n-th>鎺掑悕</n-th>
              <n-th>鏉垮潡鍚嶇О</n-th>
              <n-th>鍑€娴佸叆/浜垮厓</n-th>
            </n-tr>
          </n-thead>
          <n-tbody>
            <n-tr v-for="(item, idx) in outflowList" :key="item.code">
              <n-td>{{ idx + 1 }}</n-td>
              <n-td>
                <n-tag :bordered="false" type="success" size="small">{{ item.name }}</n-tag>
              </n-td>
              <n-td>
                <n-text type="success">{{ (item.netInflow / 100000000).toFixed(2) }}</n-text>
              </n-td>
            </n-tr>
            <n-tr v-if="outflowList.length === 0">
              <n-td colspan="3" style="text-align: center; color: #999;">鏆傛棤鏁版嵁</n-td>
            </n-tr>
          </n-tbody>
        </n-table>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>

