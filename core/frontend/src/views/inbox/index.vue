<template>
	<div class="inbox-page">
		<div class="toolbar">
			<div>
				<h1>收件箱 / 客户回复</h1>
				<p>统一查看已配置的 126、163 邮箱</p>
			</div>
			<n-space
				><n-button @click="router.push('/smtp')">管理邮箱（添加 / 删除）</n-button
				><n-button :loading="loading" @click="refresh">刷新列表</n-button
				><n-button :loading="queueing" @click="queue(0)">安排全部检查</n-button></n-space
			>
		</div>
		<n-alert type="info" :show-icon="true" class="notice">
			自动轮询每个账号至少间隔 15
			分钟。录制演示时可选择一个邮箱，点击“立即检查当前邮箱”，检查完成后自动刷新；同一邮箱手动检查冷却
			60 秒，账号连接按顺序进行，可能需要等待当前连接完成。 每次同步最近 100 封邮件，并缓存最新 10
			封正文；其他正文按需读取。
			查看及“已处理”标记不会改变原邮箱的已读状态，也不会移动或删除原邮件。
		</n-alert>
		<n-alert v-if="checkMessage" :type="checkState" :show-icon="true" class="notice">{{
			checkMessage
		}}</n-alert>
		<n-collapse class="account-panel" :default-expanded-names="['accounts']">
			<n-collapse-item
				:title="`收件账号 · ${accounts.filter(a => a.last_success > 0).length} / ${accounts.length} 已同步`"
				name="accounts">
				<div class="accounts-grid">
					<article v-for="account in accounts" :key="account.id" class="account-card">
						<div class="account-title">
							<strong>{{ account.email }}</strong
							><n-switch
								:value="account.enabled"
								:disabled="checkingId === account.id"
								size="small"
								@update:value="value => toggle(account.id, value)" />
						</div>
						<div :class="{ 'account-error': !!account.last_error }">{{ account.status }}</div>
						<div v-if="account.last_error" class="account-error">{{ account.last_error }}</div>
						<div class="muted">最近同步：{{ date(account.last_success) }}</div>
						<div v-if="account.enabled && account.next_poll" class="muted">
							下次检查：{{ date(account.next_poll) }}
						</div>
						<div class="account-footer">
							<span class="muted"
								>缓存 {{ account.cached_count }} 封 / 原收件箱 {{ account.mailbox_total }} 封</span
							><n-button
								text
								type="primary"
								:loading="checkingId === account.id"
								:disabled="cannotCheck(account)"
								@click="checkNow(account.id)"
								>立即检查</n-button
							>
						</div>
						<div v-if="cooldown(account) > 0" class="muted">
							{{ cooldown(account) }} 秒后可再次手动检查
						</div>
					</article>
				</div>
				<n-empty
					v-if="!accounts.length && !loading"
					description="尚无可收件的账号。请在 SMTP 页面保存 126 或 163 邮箱及授权码。" />
			</n-collapse-item>
		</n-collapse>
		<section class="mail-panel">
			<n-tabs v-model:value="kind" type="line" @update:value="filterChanged">
				<n-tab name="all">全部邮件</n-tab><n-tab name="customer_reply">客户回复</n-tab
				><n-tab name="bounce">退信</n-tab><n-tab name="other">其他邮件</n-tab>
			</n-tabs>
			<div class="filters">
				<n-select
					v-model:value="accountId"
					:options="accountOptions"
					@update:value="filterChanged" /><n-input
					v-model:value="keyword"
					placeholder="搜索主题、发件人"
					clearable
					@keyup.enter="filterChanged" /><n-button @click="filterChanged">搜索</n-button
				><n-button
					type="primary"
					:loading="!!checkingId"
					:disabled="cannotCheck(currentAccount)"
					@click="checkNow(accountId)"
					>立即检查当前邮箱</n-button
				>
			</div>
			<p v-if="!accountId" class="muted">
				立即检查前，请先在上方选择一个收件账号；也可以点击账号卡片上的“立即检查”。
			</p>
			<p v-else-if="currentAccount && cooldown(currentAccount) > 0" class="muted">
				当前邮箱：{{ cooldown(currentAccount) }} 秒后可再次手动检查。
			</p>
			<p class="muted">
				“客户回复”按已有发送记录和客户名单中的回复标识识别；退信分类用于查看，不会自动修改拒发名单。
			</p>
			<n-data-table
				:columns="columns"
				:data="messages"
				:loading="loading"
				:row-key="(row: InboxMail) => row.id"
				:row-props="rowProps"
				:scroll-x="1050" />
			<div class="pagination">
				<span class="muted">已缓存 {{ total }} 封</span
				><n-pagination
					v-model:page="page"
					:page-count="Math.max(1, Math.ceil(total / 20))"
					@update:page="loadMessages" />
			</div>
		</section>
		<n-drawer v-model:show="showMail" :width="drawerWidth" placement="right">
			<n-drawer-content title="邮件内容" closable>
				<n-spin :show="detailLoading">
					<n-alert v-if="detailError" type="error">{{ detailError }}</n-alert>
					<template v-if="selected">
						<h2>{{ selected.subject || '（无主题）' }}</h2>
						<dl>
							<dt>发件人</dt>
							<dd>{{ selected.from_name }} &lt;{{ selected.from_address }}&gt;</dd>
							<dt>收件账号</dt>
							<dd>{{ selected.account }}</dd>
							<dt>时间</dt>
							<dd>{{ date(selected.received_at) }}</dd>
						</dl>
						<n-space align="center" class="detail-actions"
							><n-checkbox :checked="selected.processed" @update:checked="mark"
								>已处理（仅本项目）</n-checkbox
							><n-button tag="a" :href="originalMailbox" target="_blank" rel="noopener noreferrer"
								>打开原邮箱</n-button
							></n-space
						>
						<n-alert v-if="selected.truncated" type="warning"
							>正文预览限 1 MB，此邮件较大，请打开原邮箱查看完整内容。</n-alert
						>
						<n-tabs v-model:value="bodyTab" type="line">
							<n-tab-pane name="text" tab="纯文本">
								<pre class="body-text">{{
									selected.body_text || '此邮件没有纯文本正文，请选择 HTML 预览。'
								}}</pre>
							</n-tab-pane>
							<n-tab-pane v-if="selected.body_html" name="html" tab="HTML 预览"
								><p class="muted">预览已禁用脚本、远程图片和链接跳转。</p>
								<iframe
									title="邮件 HTML 预览"
									sandbox=""
									referrerpolicy="no-referrer"
									:srcdoc="safeHTML"
							/></n-tab-pane>
						</n-tabs>
						<div v-if="attachments.length" class="attachments">
							<strong>附件</strong>
							<div v-for="(name, index) in attachments" :key="index">{{ name }}</div>
							<p class="muted">请打开原邮箱下载附件。</p>
						</div>
					</template>
				</n-spin>
			</n-drawer-content>
		</n-drawer>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, onUnmounted } from 'vue'
import { NTag, type DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import {
	inboxAccounts,
	inboxList,
	inboxDetail,
	inboxSync,
	inboxCheckNow,
	inboxToggle,
	inboxMark,
	type InboxAccount,
	type InboxMail,
} from '@/api/modules/nativeInbox'

const router = useRouter()
const accounts = ref<InboxAccount[]>([]),
	messages = ref<InboxMail[]>([])
const accountId = ref(0),
	kind = ref('all'),
	keyword = ref(''),
	page = ref(1),
	total = ref(0)
const loading = ref(false),
	queueing = ref(false),
	showMail = ref(false),
	detailLoading = ref(false)
const checkingId = ref(0),
	checkMessage = ref(''),
	checkState = ref<'info' | 'success' | 'error'>('info')
const clockNow = ref(Math.floor(Date.now() / 1000))
const currentAccount = computed(() =>
	accounts.value.find(account => account.id === accountId.value)
)
const cooldown = (account: InboxAccount) =>
	Math.max(0, (account.last_attempt || 0) + 60 - clockNow.value)
const cannotCheck = (account?: InboxAccount) =>
	!account ||
	!account.enabled ||
	!!checkingId.value ||
	cooldown(account) > 0 ||
	(['正在同步', '正在手动检查'].includes(account.status) &&
		clockNow.value - account.last_attempt < 120)
const selected = ref<InboxMail | null>(null),
	detailError = ref(''),
	bodyTab = ref('text')
const width = ref(window.innerWidth)
const drawerWidth = computed(() => Math.min(900, width.value))
const accountOptions = computed(() => [
	{ label: '全部收件账号', value: 0 },
	...accounts.value.map(a => ({ label: a.email, value: a.id })),
])
const date = (seconds: number) =>
	seconds > 0 ? new Date(seconds * 1000).toLocaleString('zh-CN', { hour12: false }) : '尚未同步'
const labels: Record<string, string> = { customer_reply: '客户回复', bounce: '退信', other: '其他' }
const columns: DataTableColumns<InboxMail> = [
	{ title: '时间', key: 'received_at', width: 175, render: row => date(row.received_at) },
	{ title: '收件账号', key: 'account', width: 210 },
	{
		title: '发件人',
		key: 'from_address',
		width: 220,
		ellipsis: { tooltip: true },
		render: row => row.from_name || row.from_address,
	},
	{
		title: '主题（点击查看）',
		key: 'subject',
		minWidth: 250,
		ellipsis: { tooltip: true },
		render: row => row.subject || '（无主题）',
	},
	{
		title: '分类',
		key: 'kind',
		width: 100,
		render: row =>
			h(
				NTag,
				{
					size: 'small',
					type:
						row.kind === 'bounce'
							? 'warning'
							: row.kind === 'customer_reply'
								? 'success'
								: 'default',
				},
				{ default: () => labels[row.kind] || '其他' }
			),
	},
	{
		title: '处理',
		key: 'processed',
		width: 85,
		render: row => (row.processed ? '已处理' : '待处理'),
	},
]
const rowProps = (row: InboxMail) => ({ style: 'cursor:pointer', onClick: () => openMail(row) })
const attachments = computed<string[]>(() => {
	try {
		return JSON.parse(selected.value?.attachments || '[]')
	} catch {
		return []
	}
})
const originalMailbox = computed(() =>
	selected.value?.account.endsWith('@126.com') ? 'https://mail.126.com/' : 'https://mail.163.com/'
)
// A sandbox plus a deny-by-default CSP isolates arbitrary sender HTML.
const safeHTML = computed(() => {
	const doc = new DOMParser().parseFromString(selected.value?.body_html || '', 'text/html')
	doc
		.querySelectorAll('script,iframe,object,embed,form,meta,link,base')
		.forEach(node => node.remove())
	doc.querySelectorAll('*').forEach(node => {
		for (const name of node.getAttributeNames()) {
			if (
				/^on/i.test(name) ||
				['href', 'xlink:href', 'action', 'formaction', 'srcdoc', 'target'].includes(name)
			)
				node.removeAttribute(name)
		}
	})
	return `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data:; style-src 'unsafe-inline'; form-action 'none'; base-uri 'none'"><style>body{font:15px sans-serif;overflow-wrap:anywhere}img{max-width:100%}</style>${doc.documentElement.outerHTML}`
})
async function loadMessages() {
	const result = await inboxList({
		account_id: accountId.value,
		kind: kind.value,
		keyword: keyword.value.trim(),
		page: page.value,
		page_size: 20,
	})
	messages.value = result.list
	total.value = result.total
}
async function refresh() {
	if (loading.value) return
	loading.value = true
	try {
		accounts.value = await inboxAccounts()
		await loadMessages()
	} finally {
		loading.value = false
	}
}
async function filterChanged() {
	page.value = 1
	await refresh()
}
async function queue(id: number) {
	queueing.value = true
	try {
		await inboxSync(id)
		await refresh()
	} finally {
		queueing.value = false
	}
}
async function checkNow(id: number) {
	const account = accounts.value.find(item => item.id === id)
	if (cannotCheck(account)) return
	checkingId.value = id
	accountId.value = id
	page.value = 1
	checkState.value = 'info'
	checkMessage.value = `正在检查 ${account!.email}，完成后自动刷新列表，请等待当前连接完成…`
	try {
		const result = await inboxCheckNow(id)
		checkState.value = 'success'
		checkMessage.value = `检查完成（${date(result.checked_at)}）：新增缓存 ${result.new_count} 封；当前缓存 ${result.cached_count} 封，客户回复 ${result.reply_count} 封。`
	} catch (error) {
		const failure = error as { msg?: string; message?: string }
		checkState.value = 'error'
		checkMessage.value = failure?.msg || failure?.message || '检查未完成，请查看账号状态后重试。'
	} finally {
		checkingId.value = 0
		clockNow.value = Math.floor(Date.now() / 1000)
		await refresh()
	}
}
async function toggle(id: number, enabled: boolean) {
	await inboxToggle(id, enabled)
	await refresh()
}
let detailSequence = 0
async function openMail(row: InboxMail) {
	const sequence = ++detailSequence
	selected.value = null
	detailError.value = ''
	detailLoading.value = true
	showMail.value = true
	try {
		const result = await inboxDetail(row.id)
		if (sequence === detailSequence) {
			selected.value = result
			bodyTab.value = result.body_text ? 'text' : 'html'
		}
	} catch (error) {
		const failure = error as { msg?: string; message?: string }
		if (sequence === detailSequence)
			detailError.value = failure?.msg || failure?.message || '无法读取正文，请检查收件账号状态。'
	} finally {
		if (sequence === detailSequence) detailLoading.value = false
	}
}
async function mark(processed: boolean) {
	if (!selected.value) return
	await inboxMark(selected.value.id, processed)
	selected.value.processed = processed
	await loadMessages()
}
let timer: ReturnType<typeof setInterval>
let clockTimer: ReturnType<typeof setInterval>
const resize = () => {
	width.value = window.innerWidth
}
onMounted(() => {
	document.body.classList.add('mail-inbox-responsive')
	void refresh().catch(() => {})
	timer = setInterval(() => {
		if (!document.hidden) void refresh().catch(() => {})
	}, 10000)
	clockTimer = setInterval(() => {
		clockNow.value = Math.floor(Date.now() / 1000)
	}, 1000)
	window.addEventListener('resize', resize)
})
onUnmounted(() => {
	document.body.classList.remove('mail-inbox-responsive')
	clearInterval(timer)
	clearInterval(clockTimer)
	window.removeEventListener('resize', resize)
	detailSequence++
})
</script>

<style>
body.mail-inbox-responsive {
	min-width: 0;
}
</style>

<style scoped>
.inbox-page {
	padding: 24px;
	max-width: 1700px;
	margin: auto;
}
.toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16px;
	margin-bottom: 20px;
}
h1 {
	font-size: 22px;
	margin: 0 0 4px;
}
h2 {
	font-size: 20px;
	overflow-wrap: anywhere;
}
.toolbar p,
.muted {
	color: #72808b;
	font-size: 13px;
}
.toolbar p {
	margin: 0;
}
.notice {
	margin-bottom: 20px;
}
.account-panel,
.mail-panel {
	background: var(--n-color, #fff);
	border-radius: 8px;
	padding: 20px;
	margin-bottom: 20px;
}
.accounts-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
	gap: 12px;
}
.account-card {
	border: 1px solid #e6ece9;
	border-radius: 6px;
	padding: 12px;
	font-size: 13px;
}
.account-title {
	display: flex;
	justify-content: space-between;
	gap: 8px;
	margin-bottom: 8px;
}
.account-title strong {
	overflow-wrap: anywhere;
}
.account-error {
	color: #b53b35;
	overflow-wrap: anywhere;
}
.account-footer {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 8px;
	margin-top: 8px;
}
.filters {
	display: flex;
	gap: 12px;
	margin: 16px 0;
}
.filters .n-select {
	width: 290px;
	flex-shrink: 0;
}
.filters .n-input {
	max-width: 420px;
}
.pagination {
	display: flex;
	justify-content: space-between;
	margin-top: 16px;
	gap: 12px;
}
dl {
	display: grid;
	grid-template-columns: 80px 1fr;
	gap: 8px;
}
dt {
	color: #72808b;
}
dd {
	margin: 0;
	overflow-wrap: anywhere;
}
.detail-actions {
	margin: 20px 0;
}
.body-text {
	white-space: pre-wrap;
	overflow-wrap: anywhere;
	font-family: inherit;
	line-height: 1.7;
}
iframe {
	width: 100%;
	height: 65vh;
	border: 1px solid #e6ece9;
	border-radius: 4px;
	background: #fff;
}
.attachments {
	padding: 12px;
	background: #f5f8f6;
	border-radius: 6px;
}
.attachments div {
	overflow-wrap: anywhere;
}
@media (max-width: 700px) {
	.inbox-page {
		padding: 12px;
	}
	.toolbar {
		flex-direction: column;
		align-items: flex-start;
	}
	.filters {
		flex-wrap: wrap;
	}
	.filters .n-select,
	.filters .n-input {
		width: 100%;
		max-width: none;
	}
	.account-panel,
	.mail-panel {
		padding: 12px;
	}
	.accounts-grid {
		grid-template-columns: 1fr;
	}
	.pagination {
		flex-wrap: wrap;
	}
}
</style>
