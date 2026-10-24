<template>
  <LineChart
    v-if="loaded"
    :key="theme.global.name"
    :data="data"
    :options="<any>options"
  />
  <div
    v-else
    class="chart-pending"
    role="status"
  >
    {{ $t('loading') }}
  </div>
</template>

<script lang="ts">
import { PropType, ref } from 'vue'
import { useTheme } from 'vuetify'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Filler,
} from 'chart.js'
import { HumanReadable } from '@/plugins/utils'
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Filler
)
ChartJS.defaults.font.family = 'Avenir Next, Segoe UI, PingFang SC, Vazirmatn, sans-serif'

// One resource reading. Memory reports a pair, cpu is a percent, and the
// network and disk counters are totals a delta is taken from.
interface Usage {
  current: number
  total: number
}

interface NetUsage {
  recv: number
  sent: number
  precv: number
  psent: number
}

interface DiskIo {
  read: number
  write: number
}

// The reading a chart draws. Only the section its own type needs is read, and
// every reading carries that section whole.
interface TilesData {
  cpu: number
  mem: Usage
  net: NetUsage
  dio: DiskIo
}

// The scales below are linear, so chart.js hands each tick its numeric value.
type TickCallback = (label: number) => string

// Only the chart.js options these two presets actually set. The chart itself
// takes the far wider shape chart.js defines, which is why it is cast on the
// way in.
interface AxisOptions {
  min?: number
  max?: number
  grid: { color: string }
  beginAtZero?: boolean
  ticks: {
    beginAtZero?: boolean
    steps?: number
    stepValue?: number
    max?: number
    count?: number
    color?: string
    callback?: TickCallback
  }
}

interface LineChartOptions {
  animation: boolean
  responsive: boolean
  maintainAspectRatio: boolean
  interaction: { intersect: boolean, mode: string }
  elements: {
    point: { pointStyle: boolean }
    line: { tension: number, borderWidth: number }
  }
  plugins: {
    tooltip: { enabled: boolean }
    legend: { display: boolean }
  }
  scales: { y: AxisOptions, x?: AxisOptions }
}

// The series the chart draws, built fresh on every reading.
interface LineChartData {
  labels: string[]
  datasets: {
    label: string
    backgroundColor: string
    borderColor: string
    fill: boolean
    data: number[]
  }[]
}

export default {
  components: {
    LineChart: Line
  },
  props: {
    tilesData: { type: Object as PropType<TilesData>, required: true },
    type: { type: String, required: true }
  },
  data() {
    return {
      theme: useTheme(),
      loaded: false,
      labels: new Array(20).fill(''),
      // Empty until the first reading arrives; the checks below skip that one,
      // since a delta needs two.
      oldValues: { net: <NetUsage>{}, dio: <DiskIo>{} },
      options1: <LineChartOptions>{
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: 'index',
        },
        elements: {
          point: { pointStyle: false },
          line: { tension: 0.2, borderWidth: 2 },
        },
        plugins: {
          tooltip: {
            enabled: false
          },
          legend: {
              display: false,
          }
        },
        scales: {
          y: {
            min: 0,
            max: 100,
            grid: {
              color: '#777777',
            },
            beginAtZero: true,
            ticks: {
                beginAtZero: true,
                steps: 10,
                stepValue: 5,
                max: 100
            }
          }
        }
      },
      optionsNet: <LineChartOptions>{
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: 'index',
        },
        elements: {
          point: { pointStyle: false },
          line: { tension: 0.2, borderWidth: 2 },
        },
        plugins: {
          tooltip: {
            enabled: false
          },
          legend: {
              display: false,
          }
        },
        scales: {
          y: {
            grid: {
              color: '#777777',
            },
            beginAtZero: true,
            ticks: {
              callback: (label:number) => { return parseInt(String(label)).toString() },
              count: 10
            }
          }
        }
      },
      data: ref(<LineChartData>{})
    }
  },
  computed: {
    chartColors() {
      // Vuetify types a theme colour as a string or one of its object forms;
      // the palette this app defines holds CSS colour strings throughout.
      const onSurface = <string>this.theme.current.colors['on-surface']
      return {
        text: onSurface,
        primary: <string>this.theme.current.colors['chart-1'],
        secondary: <string>this.theme.current.colors['chart-2'],
        gridY: <string>this.theme.current.colors.outline,
        gridX: 'transparent',
      }
    },
    options() {
      const { text, gridY, gridX } = this.chartColors
      // Build a new options object instead of mutating component state from a computed
      const applyTheme = (o:LineChartOptions, callback?: TickCallback): LineChartOptions => ({
        ...o,
        scales: {
          ...o.scales,
          y: {
            ...o.scales.y,
            grid: { ...o.scales.y.grid, color: gridY },
            ticks: { ...o.scales.y.ticks, color: text, ...(callback ? { callback } : {}) },
          },
          x: { grid: { color: gridX }, ticks: { color: text } },
        },
      })
      switch (this.$props.type){
        case "h-net":
          return applyTheme(this.optionsNet, (label:number) => label == 0 ? "0" : HumanReadable.sizeFormat(label,0))
        case "hp-net":
          return applyTheme(this.optionsNet, (label:number) => label == 0 ? "0" : HumanReadable.packetFormat(label,0))
        case "h-dio":
          return applyTheme(this.optionsNet, (label:number) => label == 0 ? "0" : HumanReadable.sizeFormat(label,0))
      }
      return applyTheme(this.options1)
    }
  },
  watch: {
    tilesData(v:TilesData) {
      switch (this.$props.type) {
        case 'h-cpu':
          this.updateData1(v.cpu)
          break
        case 'h-mem':
          this.updateData1(v.mem.current*100/v.mem.total)
          break
        case 'h-net':
          if (this.oldValues.net.sent) {
            const downSpeed = (v.net.recv-this.oldValues.net.recv)/2  // Each 2 sec
            const upSpeed = (v.net.sent-this.oldValues.net.sent)/2  // Each 2 sec
            this.updateData2(upSpeed,downSpeed)
          }
          this.oldValues.net = v.net
          break
        case 'hp-net':
          if (this.oldValues.net.psent) {
            const downSpeed = (v.net.precv-this.oldValues.net.precv)/2  // Each 2 sec
            const upSpeed = (v.net.psent-this.oldValues.net.psent)/2  // Each 2 sec
            this.updateData2(upSpeed,downSpeed)
          }
          this.oldValues.net = v.net
          break
        case 'h-dio':
          if (this.oldValues.dio.read) {
            const downSpeed = (v.dio.read-this.oldValues.dio.read)/2  // Each 2 sec
            const upSpeed = (v.dio.write-this.oldValues.dio.write)/2  // Each 2 sec
            this.updateData2(upSpeed,downSpeed)
          }
          this.oldValues.dio = v.dio
          break
      }
    }
  },
  methods: {
    updateData1(value1: number) {
      const newData = <number[]>[]
      if (this.data.datasets){
        newData.push(...this.data.datasets[0].data,value1)
      }
      if (newData.length>20) newData.shift()
      this.data = {
        labels: this.labels,
        datasets: [
          {
            label: '',
            backgroundColor: this.chartColors.primary + '12',
            borderColor: this.chartColors.primary,
            fill: true,
            data: newData
          }
        ],
      }
      this.loaded = true
    },
    updateData2(value1: number, value2:number) {
      const newData1 = <number[]>[]
      const newData2 = <number[]>[]
      if (this.data.datasets){
        newData1.push(...this.data.datasets[0].data,value1)
        newData2.push(...this.data.datasets[1].data,value2)
      }
      if (newData1.length>20) newData1.shift()
      if (newData2.length>20) newData2.shift()
      this.data = {
        labels: this.labels,
        datasets: [
          {
            label: '',
            backgroundColor: this.chartColors.primary + '12',
            borderColor: this.chartColors.primary,
            fill: true,
            data: newData1
          },
          {
            label: '',
            backgroundColor: this.chartColors.secondary + '12',
            borderColor: this.chartColors.secondary,
            fill: true,
            data: newData2
          }
        ],
      }
      this.loaded = true
    }
  }
}
</script>

<style scoped>
.chart-pending { display: grid; place-items: center; width: 100%; height: 100%; color: var(--panel-muted); font-size: .75rem; border-bottom: 1px solid var(--panel-border); }
</style>
