<template>
	<div>
		<n-alert type="info" class="mb-20px">
			SMTP 已接收表示发件服务器接受了邮件，不代表邮件已进入客户收件箱。
			最终送达和退信尚未接入回执；打开、点击缺少可靠跟踪数据时显示为未知。 统计仅包含本机 Mail
			的发送记录，按邮件封数累计，不是去重客户数。
		</n-alert>
		<div class="native-metrics">
			<n-card title="SMTP 已接收"
				><strong>{{ stats.accepted }} 封</strong></n-card
			>
			<n-card title="发送失败"
				><strong>{{ stats.failed }} 封</strong></n-card
			>
			<n-card title="最终送达率">暂无跟踪数据</n-card>
			<n-card title="打开率">{{ trackingLabel(stats.open_state) }}</n-card>
			<n-card title="点击率">{{ trackingLabel(stats.click_state) }}</n-card>
			<n-card title="退回率">暂无跟踪数据</n-card>
		</div>
		<n-card title="发送统计" class="mb-20px">
			<p>
				SMTP 接收率：{{ acceptanceRate }}；SMTP 已接收：{{ stats.accepted }} 封；发送失败：{{
					stats.failed
				}}
				封
			</p>
			<div class="sending-chart">
				<bar-chart date-type="daily" :chart-data="sendingData" />
			</div>
		</n-card>
		<div class="rate-charts">
			<n-card v-for="rate in ratePanels" :key="rate.title" :title="rate.title">
				<div class="rate-chart"><bt-charts :options="unknownRateOptions(rate.state)" /></div>
			</n-card>
		</div>
		<n-card title="发件账号统计（按所选日期范围）" class="mb-20px">
			<n-data-table :columns="accountColumns" :data="stats.accounts" :bordered="false" />
		</n-card>
		<n-card title="发送统计（按所选日期范围）">
			<n-data-table :columns="dayColumns" :data="stats.days" :bordered="false" />
		</n-card>
	</div>
</template>

<script lang="ts" setup>
import type { NativeStats } from '../types'
import { useThemeVars } from 'naive-ui'
import BarChart from './BarChart.vue'

const props = defineProps<{ stats: NativeStats }>()
const theme = useThemeVars()
const trackingLabel = (state: NativeStats['open_state']) =>
	state === 'disabled' ? '未启用跟踪' : '暂无跟踪数据'
const outcomeColumns = [
	{ title: 'SMTP 已接收（封）', key: 'accepted' },
	{ title: '发送失败（封）', key: 'failed' },
]
const accountColumns = [{ title: '发件邮箱', key: 'email' }, ...outcomeColumns]
const dayColumns = [{ title: '日期', key: 'date' }, ...outcomeColumns]
const acceptanceRate = computed(() => {
	const attempts = props.stats.accepted + props.stats.failed
	return attempts ? `${((props.stats.accepted / attempts) * 100).toFixed(1)}%` : '暂无发送记录'
})
const sendingData = computed(() => ({
	success: {
		label: 'SMTP 已接收',
		value: props.stats.days.map(
			day => [day.date, day.accepted, 'daily'] as [string, number, string]
		),
	},
	fail: {
		label: '发送失败',
		value: props.stats.days.map(day => [day.date, day.failed, 'daily'] as [string, number, string]),
	},
}))
const ratePanels = computed(() => [
	{ title: '打开率', state: props.stats.open_state },
	{ title: '点击率', state: props.stats.click_state },
	{ title: '退回率', state: 'unavailable' as const },
])
// Missing observations are null, never a fabricated zero-percent trend.
const unknownRateOptions = (state: NativeStats['open_state']) => ({
	grid: { left: 38, right: 12, top: 30, bottom: 30 },
	xAxis: { type: 'category', data: props.stats.days.map(day => day.date) },
	yAxis: {
		type: 'value',
		name: '%',
		min: 0,
		max: 100,
		splitLine: { lineStyle: { type: 'dashed', color: theme.value.borderColor } },
	},
	series: [{ type: 'line', data: props.stats.days.map(() => null) }],
	graphic: [
		{
			type: 'text',
			left: 'center',
			top: 'middle',
			style: { text: trackingLabel(state), fill: theme.value.textColor2, fontSize: 14 },
		},
	],
})
</script>

<style lang="scss" scoped>
.sending-chart {
	height: 240px;
}
.rate-chart {
	height: 200px;
}
.rate-charts {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: 20px;
	margin-bottom: 20px;
	@media (max-width: 1000px) {
		grid-template-columns: 1fr;
	}
}
.native-metrics {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: 20px;
	margin-bottom: 20px;
	strong {
		font-size: 24px;
	}
	@media (max-width: 800px) {
		grid-template-columns: 1fr;
	}
}
</style>
