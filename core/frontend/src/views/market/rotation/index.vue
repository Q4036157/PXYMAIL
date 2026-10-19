<template>
	<div class="rotation-page" :class="{ 'results-only': resultsOnly }">
		<n-space v-if="!resultsOnly" justify="space-between" align="center">
			<div>
				<h2>账号与模板轮换发送</h2>
				<p class="hint">账号和模板分别按顺序循环，每封邮件接着上一封的模板顺序。</p>
			</div>
			<n-button @click="router.push('/market/task')">返回邮件营销</n-button>
		</n-space>
		<n-alert v-if="!resultsOnly" type="info" :show-icon="true">
			10 个账号、3 个模板时：第 10 封为邮箱 10 / 模板 1，第 11 封为邮箱 1 / 模板 2。 每账号上限 2
			封时最多发送 20 位客户；只有 19 位客户时，前 9 个账号各发 2 封，第 10 个账号发 1 封。
		</n-alert>
		<n-card
			v-if="!resultsOnly"
			title="1. 选择账号、模板与客户"
			class="section"
			:segmented="{ content: true }">
			<n-form label-placement="top" :disabled="busy">
				<div class="form-grid">
					<n-form-item label="任务名称"
						><n-input v-model:value="form.name" placeholder="例如：20 位客户轮换发送"
					/></n-form-item>
					<n-form-item label="发件人显示名"
						><n-input v-model:value="form.full_name" placeholder="填写客户看到的发件人名称"
					/></n-form-item>
					<n-form-item label="发件账号（按下方顺序轮换）">
						<div class="field-stack">
							<n-space align="center" :size="12">
								<n-button
									size="small"
									:disabled="busy || !accountOptions.length"
									@click="selectAllSenders"
									>全选可用账号</n-button
								>
								<n-button
									size="small"
									:disabled="busy || !form.senders.length"
									@click="form.senders = []"
									>清空</n-button
								>
								<n-button
									size="small"
									:disabled="busy || !accountOptions.length"
									@click="invertSenders"
									>反选</n-button
								>
								<span class="hint"
									>已选 {{ form.senders.length }} / {{ accountOptions.length }} 个</span
								>
							</n-space>
							<n-select
								v-model:value="form.senders"
								multiple
								filterable
								clearable
								:max-tag-count="3"
								:options="accountOptions"
								placeholder="选择发件账号，或点击全选可用账号" />
						</div>
					</n-form-item>
					<n-form-item label="邮件模板（按下方顺序轮换）">
						<n-select
							v-model:value="templateIds"
							multiple
							filterable
							:options="templateOptions"
							placeholder="选择 3 个模板" />
					</n-form-item>
					<n-form-item label="收件客户">
						<div class="field-stack">
							<n-radio-group v-model:value="recipientSource" :disabled="busy">
								<n-space
									><n-radio value="group">客户分组 / CSV 名单</n-radio
									><n-radio value="paste">直接粘贴邮箱</n-radio></n-space
								>
							</n-radio-group>
							<div v-if="recipientSource === 'group'" class="group-picker">
								<group-select v-model:value="groupId" :disabled="busy" />
							</div>
							<template v-else>
								<n-input
									v-model:value="recipientText"
									type="textarea"
									:autosize="{ minRows: 4, maxRows: 8 }"
									:maxlength="100000"
									placeholder="直接粘贴客户邮箱，无需上传 CSV。支持换行、逗号、分号和空格分隔。&#10;例如：client1@example.com&#10;client2@example.com, client3@example.com" />
								<p class="hint paste-summary">
									预识别 {{ pasted.emails.length }} 个邮箱；重复 {{ pasted.duplicates }} 项；无效
									{{ pasted.invalid.length }} 项。按首次出现顺序去重，最终发送名单以预览为准。
								</p>
								<p v-if="pasted.invalid.length" class="invalid-input">
									无效项：{{ pasted.invalid.slice(0, 5).join('、')
									}}{{ pasted.invalid.length > 5 ? ' 等' : '' }}
								</p>
								<n-space v-if="pasted.emails.length > total" align="center">
									<span class="hint"
										>粘贴 {{ pasted.emails.length }} 位，当前上限 {{ total }} 位。</span
									>
									<n-button size="small" :disabled="busy || !canFitPasted" @click="fitPasted"
										>按粘贴人数调整上限</n-button
									>
								</n-space>
								<p class="hint">确认创建时自动保存本次名单，不需要预先创建分组。</p>
							</template>
						</div>
					</n-form-item>
					<n-form-item label="每个账号最多发送几封"
						><n-input-number v-model:value="perSender" :min="1" :max="100"
					/></n-form-item>
					<n-form-item label="每封间隔（秒，至少 20 秒）"
						><n-input-number v-model:value="interval" :min="20" :max="3600"
					/></n-form-item>
					<n-form-item label="共同邮件主题（可在每个模板下单独修改）"
						><n-input
							v-model:value="commonSubject"
							placeholder="填写邮件主题"
							@update:value="applySubject"
					/></n-form-item>
					<n-form-item label="已发送客户的处理方式">
						<n-radio-group v-model:value="sentPolicy">
							<n-space
								><n-radio value="ask">发现后弹窗确认</n-radio
								><n-radio value="skip">始终跳过已发送客户</n-radio></n-space
							>
						</n-radio-group>
					</n-form-item>
				</div>
				<div class="selection-grid">
					<div>
						<h3>账号顺序 · {{ form.senders.length }} 个</h3>
						<div class="sender-order-list">
							<div v-for="(email, index) in form.senders" :key="email" class="order-row">
								<span class="order-number">{{ index + 1 }}</span
								><span class="order-text">{{ email }}</span>
								<n-tag v-if="previewData" size="small" type="info"
									>{{ senderCount(email) }} 封</n-tag
								>
								<n-button size="tiny" :disabled="index === 0 || busy" @click="moveSender(index, -1)"
									>上移</n-button
								>
								<n-button
									size="tiny"
									:disabled="index === form.senders.length - 1 || busy"
									@click="moveSender(index, 1)"
									>下移</n-button
								>
							</div>
						</div>
						<p v-if="!form.senders.length" class="hint">先选择发件账号。</p>
					</div>
					<div>
						<h3>模板顺序 · {{ templateIds.length }} 个</h3>
						<div v-for="(id, index) in templateIds" :key="id" class="template-row">
							<div class="order-row">
								<span class="order-number">{{ index + 1 }}</span
								><span class="order-text">{{ templateName(id) }}</span>
								<n-button
									size="tiny"
									:disabled="index === 0 || busy"
									@click="moveTemplate(index, -1)"
									>上移</n-button
								>
								<n-button
									size="tiny"
									:disabled="index === templateIds.length - 1 || busy"
									@click="moveTemplate(index, 1)"
									>下移</n-button
								>
							</div>
							<n-input v-model:value="subjects[id]" placeholder="该模板的邮件主题（必填）" />
						</div>
						<p v-if="!templateIds.length" class="hint">模板正文保持原样，包括原有链接。</p>
					</div>
				</div>
			</n-form>
			<n-space align="center" class="section">
				<n-button type="primary" :loading="busy" :disabled="!canPreview" @click="preview"
					>预览分配（不发送）</n-button
				>
				<span
					>最多 {{ total }} 位客户，每账号最多 {{ perSender || 0 }} 封<span v-if="previewData"
						>；本次实际 {{ previewData.items.length }} 位</span
					></span
				>
			</n-space>
			<p v-if="total > 1000" class="invalid-input">
				单次任务上限为 1000 位客户，请减少账号数量或每账号上限。
			</p>
			<p class="hint">
				按 CSV
				或粘贴邮箱的原顺序选取客户。已发送客户可在弹窗确认后再次纳入；退信、退订、管理员、其他任务占用及发送状态待核对的记录仍排除。原项目的“未发送待核对”成员名单可在这里预览核对；其他未确认客户会被排除。
			</p>
		</n-card>
		<n-card v-if="previewData" title="2. 检查每封邮件的分配" class="section">
			<n-alert
				v-if="previewData.config.include_sent && previewData.repeated.count"
				type="warning"
				:show-icon="true"
				class="notice">
				已确认包含
				{{
					previewData.repeated.count
				}}
				位已发送客户，原发送历史保留。本次确认仅对当前名单和分配生效。
			</n-alert>
			<n-alert v-if="!previewData.items.length" type="warning" class="notice">
				跳过已发送客户后没有可发送客户。可重新预览，在弹窗选择“包含已发送客户”。
			</n-alert>
			<n-alert v-if="unassignedCount" type="warning" :show-icon="true" class="notice">
				粘贴名单还有
				{{ unassignedCount }} 位可发送客户未分配。请提高每账号上限后重新预览，全部分配后才能创建。
			</n-alert>
			<p v-if="previewData.input" class="hint">
				本次粘贴 {{ previewData.input.total }} 项，识别
				{{ previewData.input.emails.length }} 个不同邮箱，重复
				{{ previewData.input.duplicates }} 项，无效 {{ previewData.input.invalid.length }} 项。
			</p>
			<n-alert v-if="previewData.items.length < total" type="info" :show-icon="true" class="notice">
				本次实际发送 {{ previewData.items.length }} 位客户，上限
				{{ total }} 位。可发送客户不足上限时，按实际名单创建任务。
			</n-alert>
			<p>发件分配：{{ senderDistribution }}</p>
			<p>
				可发送 {{ previewData.items.length }} 位客户。{{
					skippedText || '本次扫描未发现需排除的记录。'
				}}
			</p>
			<n-data-table
				:columns="itemColumns"
				:data="previewData.items"
				:scroll-x="1000"
				:pagination="{ pageSize: 20 }" />
			<n-space align="center" class="section">
				<n-button
					type="primary"
					:loading="busy"
					:disabled="unassignedCount > 0 || !previewData.items.length"
					@click="create"
					>确认名单并创建待发送任务</n-button
				>
				<span class="hint">创建后会保留名单，点击下方任务的“开始 / 继续”才会发送。</span>
			</n-space>
		</n-card>
		<n-card
			:title="resultsOnly ? `多账号轮换任务（${filteredJobs.length}）` : '3. 轮换任务与发送结果'"
			class="section">
			<template #header-extra
				><n-button size="small" :loading="refreshing" @click="refresh">刷新结果</n-button></template
			>
			<n-alert type="info" :show-icon="false" class="notice"
				>开始前会验证所选账号的 SMTP
				授权。验证失败或发送失败会暂停；暂停后保留原分配，已经提交的一封可能仍会完成。SMTP
				已接收不代表最终送达。</n-alert
			>
			<n-data-table
				:columns="jobColumns"
				:data="filteredJobs"
				:loading="refreshing"
				:scroll-x="1400"
				:pagination="{ pageSize: 10 }" />
		</n-card>
		<n-card v-if="detail" :title="'任务详情：' + detail.job.name" class="section">
			<n-alert v-if="detail.job.error" type="warning" class="notice">{{
				detail.job.error
			}}</n-alert>
			<n-data-table
				:columns="detailColumns"
				:data="detail.items"
				:scroll-x="1150"
				:pagination="{ pageSize: 20 }" />
		</n-card>
		<n-modal
			:show="!!repeatPreview"
			:mask-closable="false"
			:close-on-esc="!busy"
			@update:show="closeRepeat">
			<n-card
				v-if="repeatPreview"
				title="发现已发送记录，是否再次发送？"
				class="repeat-dialog"
				closable
				@close="closeRepeat(false)">
				<p>
					按原名单顺序，本次最多安排 {{ repeatPreview.items.length }} 位客户，其中
					{{ repeatPreview.repeated.count }} 位已有发送记录。
				</p>
				<p v-if="repeatPreview.repeated.same">
					{{ repeatPreview.repeated.same }} 位曾发送过相同主题及正文的邮件。
				</p>
				<p v-if="repeatPreview.repeated.other">
					{{ repeatPreview.repeated.other }} 位曾发送过其他主题或正文的邮件。
				</p>
				<p v-if="repeatPreview.repeated.unknown">
					{{ repeatPreview.repeated.unknown }} 位有历史记录，但无法核对当时的完整内容。
				</p>
				<p class="hint">
					已有记录表示 SMTP
					已接收，不能据此确认最终送达。退订、退信及其他任务占用记录已排除。此处只确认预览名单，稍后仍需创建任务并点击开始。
				</p>
				<n-collapse
					><n-collapse-item title="查看已有发送记录的客户" name="recipients">
						<div class="repeat-recipients">
							<div
								v-for="recipient in repeatPreview.repeated.recipients || []"
								:key="recipient.email">
								{{ recipient.email }} · {{ historyLabel(recipient.kind) }}
							</div>
						</div>
					</n-collapse-item></n-collapse
				>
				<template #footer
					><n-space justify="end">
						<n-button :disabled="busy" @click="closeRepeat(false)">取消预览</n-button>
						<n-button :loading="busy" @click="skipRepeated">仅发送未发送客户</n-button>
						<n-button type="primary" :disabled="busy" @click="includeRepeated"
							>包含已发送客户</n-button
						>
					</n-space></template
				>
			</n-card>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NProgress, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { formatTime } from '@/utils'
import GroupSelect from '../task/components/GroupSelect.vue'
import {
	rotationOptions,
	rotationPreview,
	rotationCreate,
	rotationList,
	rotationDetail,
	rotationAction,
	type RotationConfig,
	type RotationOptions,
	type RotationPreview,
	type RotationItem,
	type RotationJob,
} from '@/api/modules/nativeRotation'

const props = defineProps<{
	initial?: Partial<RotationConfig>
	resultsOnly?: boolean
	keyword?: string
}>()
const router = useRouter()
const options = ref<RotationOptions>({ accounts: [], templates: [], groups: [] })
const form = reactive({ name: '客户邮件轮换发送', full_name: '真模型', senders: [] as string[] })
const templateIds = ref<number[]>([])
const groupId = ref<number | null>(null)
const recipientSource = ref<'group' | 'paste'>('group')
const recipientText = ref('')
const sentPolicy = ref<'ask' | 'skip'>('ask')
const repeatPreview = ref<RotationPreview | null>(null)
const pasted = computed(() => {
	const emails: string[] = [],
		invalid: string[] = [],
		seen = new Set<string>()
	let duplicates = 0
	for (const token of recipientText.value.split(/[\s,;，；、\uFEFF]+/).filter(Boolean)) {
		const email = token.toLowerCase()
		if (email.length > 254 || !/^[^\s@<>"(),;]+@[^\s@<>"(),;]+\.[^\s@<>"(),;]+$/.test(email))
			invalid.push(token)
		else if (seen.has(email)) duplicates++
		else {
			seen.add(email)
			emails.push(email)
		}
	}
	return { emails, invalid, duplicates }
})
const perSender = ref<number | null>(2)
const interval = ref<number | null>(20)
const commonSubject = ref('')
const subjects = reactive<Record<number, string>>({})
const busy = ref(false)
const refreshing = ref(false)
const previewData = ref<RotationPreview | null>(null)
const jobs = ref<RotationJob[]>([])
const filteredJobs = computed(() => {
	const keyword = props.keyword?.trim().toLowerCase() || ''
	return keyword
		? jobs.value.filter(job => `${job.name} ${job.config}`.toLowerCase().includes(keyword))
		: jobs.value
})
const detail = ref<{ job: RotationJob; items: RotationItem[] } | null>(null)
let requestKey = ''
let timer: ReturnType<typeof setInterval> | undefined
const accountOptions = computed(() =>
	options.value.accounts.map(a => ({ value: a.email, label: a.email }))
)
const templateOptions = computed(() =>
	options.value.templates.map(t => ({ value: Number(t.id), label: t.temp_name }))
)
const total = computed(() => form.senders.length * (perSender.value || 0))
const unassignedCount = computed(() => previewData.value?.skipped['超出本次发送上限'] || 0)
const canFitPasted = computed(
	() =>
		form.senders.length > 0 &&
		pasted.value.emails.length <= 1000 &&
		Math.ceil(pasted.value.emails.length / form.senders.length) <= 100 &&
		Math.ceil(pasted.value.emails.length / form.senders.length) * form.senders.length <= 1000
)
const senderCount = (email: string) =>
	previewData.value?.items.filter(item => item.sender === email).length || 0
const senderDistribution = computed(() => {
	const counts = new Map<number, number>()
	for (const email of form.senders) {
		const count = senderCount(email)
		counts.set(count, (counts.get(count) || 0) + 1)
	}
	return [...counts.entries()]
		.sort((a, b) => b[0] - a[0])
		.map(([count, accounts]) => `${accounts} 个账号各 ${count} 封`)
		.join('；')
})
const canPreview = computed(
	() =>
		form.senders.length > 0 &&
		templateIds.value.length > 0 &&
		(recipientSource.value === 'group'
			? !!groupId.value
			: pasted.value.emails.length > 0 && pasted.value.emails.length <= 1000) &&
		total.value > 0 &&
		total.value <= 1000 &&
		(interval.value || 0) >= 20 &&
		templateIds.value.every(id => !!subjects[id]?.trim())
)
const skippedText = computed(() =>
	Object.entries(previewData.value?.skipped || {})
		.map(([reason, count]) => `${reason} ${count} 位`)
		.join('；')
)
const templateName = (id: number) =>
	options.value.templates.find(t => Number(t.id) === id)?.temp_name || `模板 ${id}`
const config = (): RotationConfig => ({
	...form,
	senders: [...form.senders],
	templates: templateIds.value.map(id => ({ id, subject: subjects[id] || '' })),
	group_id: recipientSource.value === 'group' ? groupId.value || 0 : 0,
	recipient_source: recipientSource.value,
	recipient_text: recipientSource.value === 'paste' ? recipientText.value : '',
	per_sender: perSender.value || 0,
	interval: interval.value || 0,
})
watch(
	templateIds,
	ids => {
		for (const id of ids) if (subjects[id] === undefined) subjects[id] = commonSubject.value
	},
	{ deep: true }
)
watch(
	() => JSON.stringify({ ...config(), sentPolicy: sentPolicy.value }),
	() => {
		previewData.value = null
		repeatPreview.value = null
	}
)
function applySubject(value: string) {
	for (const id of templateIds.value) subjects[id] = value
}
function selectAllSenders() {
	const selected = new Set(form.senders)
	form.senders = [
		...form.senders,
		...accountOptions.value.map(account => account.value).filter(email => !selected.has(email)),
	]
}
function invertSenders() {
	const selected = new Set(form.senders)
	form.senders = accountOptions.value
		.map(account => account.value)
		.filter(email => !selected.has(email))
}
function fitPasted() {
	if (canFitPasted.value)
		perSender.value = Math.max(1, Math.ceil(pasted.value.emails.length / form.senders.length))
}
function moveSender(index: number, offset: number) {
	const list = [...form.senders]
	;[list[index], list[index + offset]] = [list[index + offset], list[index]]
	form.senders = list
}
function moveTemplate(index: number, offset: number) {
	const list = [...templateIds.value]
	;[list[index], list[index + offset]] = [list[index + offset], list[index]]
	templateIds.value = list
}
async function preview() {
	busy.value = true
	previewData.value = null
	try {
		const result = await rotationPreview({ ...config(), include_sent: sentPolicy.value === 'ask' })
		if (sentPolicy.value === 'ask' && result.repeated.count) repeatPreview.value = result
		else acceptPreview(result)
	} catch {
		/* API displays the error. */
	} finally {
		busy.value = false
	}
}
function acceptPreview(result: RotationPreview) {
	previewData.value = result
	requestKey = `rotation-${Date.now()}-${Math.random().toString(36).slice(2)}`
}
function closeRepeat(show: boolean) {
	if (!show && !busy.value) repeatPreview.value = null
}
function includeRepeated() {
	if (!repeatPreview.value || busy.value) return
	acceptPreview(repeatPreview.value)
	repeatPreview.value = null
}
async function skipRepeated() {
	if (!repeatPreview.value || busy.value) return
	busy.value = true
	try {
		acceptPreview(await rotationPreview({ ...repeatPreview.value.config, include_sent: false }))
		repeatPreview.value = null
	} catch {
		/* API displays the error. */
	} finally {
		busy.value = false
	}
}
async function create() {
	if (!previewData.value) return
	busy.value = true
	try {
		const result = await rotationCreate(
			previewData.value.config,
			previewData.value.hash,
			requestKey
		)
		previewData.value = null
		detail.value = await rotationDetail(result.id)
		await refresh()
	} catch {
		/* API displays the error. */
	} finally {
		busy.value = false
	}
}
async function refresh() {
	if (refreshing.value || (busy.value && !detail.value)) return
	refreshing.value = true
	try {
		jobs.value = await rotationList()
		if (detail.value) detail.value = await rotationDetail(detail.value.job.id)
	} catch {
		/* API displays the error. */
	} finally {
		refreshing.value = false
	}
}
async function action(job: RotationJob, value: string) {
	busy.value = true
	try {
		await rotationAction(job.id, value)
		detail.value = await rotationDetail(job.id)
		await refresh()
	} catch {
		/* API displays the error. */
	} finally {
		busy.value = false
	}
}
async function showDetail(job: RotationJob) {
	try {
		detail.value = await rotationDetail(job.id)
	} catch {
		/* API displays the error. */
	}
}
const statusLabels: Record<string, string> = {
	paused: '待发送 / 已暂停',
	running: '发送中',
	completed: '已完成',
	cancelled: '已停止',
}
const statusLabel = (value: string) => statusLabels[value] || value
const historyLabels: Record<string, string> = {
	same: '相同主题及正文',
	other: '其他内容',
	unknown: '历史内容待核对',
	confirmed: '已确认再次发送',
}
const historyLabel = (kind?: string) => historyLabels[kind || ''] || '无发送记录'
const itemColumns: DataTableColumns<RotationItem> = [
	{ key: 'sequence', title: '顺序', width: 65 },
	{ key: 'recipient', title: '客户邮箱', minWidth: 200 },
	{
		key: 'previous_sent',
		title: '本次前历史',
		minWidth: 150,
		render: row =>
			row.previous_sent
				? h(
						NTag,
						{ type: 'warning', size: 'small' },
						{ default: () => historyLabel(row.previous_sent) }
					)
				: row.task_id
					? '—'
					: '未发送',
	},
	{ key: 'sender', title: '发件账号', minWidth: 200 },
	{ key: 'template_name', title: '模板', minWidth: 190 },
	{ key: 'subject', title: '邮件主题', minWidth: 260 },
]
const detailColumns: DataTableColumns<RotationItem> = [
	...itemColumns,
	{
		key: 'state',
		title: '结果',
		width: 130,
		render: row =>
			row.is_sent === 1
				? 'SMTP 已接收'
				: row.is_sent === 3
					? '发送失败'
					: row.is_sent === 2
						? 'SMTP 处理中 / 待核对'
						: row.state === 'dispatched'
							? '已排队'
							: '待发送',
	},
]
const jobColumns: DataTableColumns<RotationJob> = [
	{ key: 'created_at', title: '创建时间', width: 175, render: row => formatTime(row.created_at) },
	{ key: 'name', title: '任务', minWidth: 180 },
	{ key: 'total', title: '总数', width: 65 },
	{ key: 'accepted', title: 'SMTP 已接收', width: 120 },
	{ key: 'failed', title: '失败', width: 65 },
	{
		key: 'pending',
		title: '待完成',
		width: 80,
		render: row => Math.max(0, row.total - row.accepted - row.failed),
	},
	{
		key: 'progress',
		title: '进度',
		width: 140,
		render: row =>
			h(NProgress, {
				type: 'line',
				percentage:
					row.total > 0
						? Math.min(100, Math.round(((row.accepted + row.failed) / row.total) * 100))
						: 0,
				status: row.failed > 0 ? 'warning' : 'success',
			}),
	},
	{
		key: 'status',
		title: '状态',
		width: 150,
		render: row =>
			h(
				NTag,
				{
					type:
						row.status === 'running' ? 'info' : row.status === 'completed' ? 'success' : 'default',
				},
				{ default: () => statusLabel(row.status) }
			),
	},
	{ key: 'error', title: '提示', minWidth: 240 },
	{
		key: 'actions',
		title: '操作',
		width: 250,
		render: row =>
			h(
				NSpace,
				{ size: 10 },
				{
					default: () => [
						h(
							NButton,
							{ text: true, type: 'primary', disabled: busy.value, onClick: () => showDetail(row) },
							{ default: () => '详情' }
						),
						row.status === 'paused'
							? h(
									NButton,
									{
										text: true,
										type: 'primary',
										disabled: busy.value,
										onClick: () => action(row, 'start'),
									},
									{ default: () => '开始 / 继续' }
								)
							: null,
						row.status === 'running'
							? h(
									NButton,
									{ text: true, disabled: busy.value, onClick: () => action(row, 'pause') },
									{ default: () => '暂停' }
								)
							: null,
						['paused', 'running'].includes(row.status)
							? h(
									NButton,
									{
										text: true,
										type: 'error',
										disabled: busy.value,
										onClick: () => action(row, 'stop'),
									},
									{ default: () => '停止' }
								)
							: null,
					],
				}
			),
	},
]
onMounted(async () => {
	try {
		if (!props.resultsOnly) options.value = await rotationOptions()
		const initial = props.initial
		if (initial) {
			form.senders = (initial.senders || []).filter(email =>
				options.value.accounts.some(a => a.email === email)
			)
			if (initial.full_name) form.full_name = initial.full_name
			if (initial.group_id) groupId.value = initial.group_id
			if (initial.recipient_source === 'paste') recipientSource.value = 'paste'
			if (initial.recipient_text) recipientText.value = initial.recipient_text
			templateIds.value = (initial.templates || [])
				.filter(t => options.value.templates.some(option => Number(option.id) === t.id))
				.map(t => t.id)
			for (const template of initial.templates || []) subjects[template.id] = template.subject
			commonSubject.value = initial.templates?.[0]?.subject || ''
		}
		await refresh()
		timer = setInterval(() => {
			if (!busy.value) void refresh()
		}, 5000)
	} catch {
		/* API displays the error. */
	}
})
onUnmounted(() => {
	if (timer) clearInterval(timer)
})
</script>

<style scoped>
:global(body:has(.rotation-page)) {
	min-width: 0;
}
.rotation-page {
	padding: 24px;
	max-width: 1600px;
	min-width: 0;
	margin: 0 auto;
}
.rotation-page.results-only {
	padding: 0;
	max-width: none;
	margin: 0 0 20px;
}
h2 {
	font-size: 22px;
	margin: 0 0 6px;
}
h3 {
	font-size: 15px;
	margin-bottom: 12px;
}
.hint {
	color: var(--bt-text-color-secondary, #68777a);
	font-size: 13px;
	line-height: 1.7;
}
.section {
	margin-top: 20px;
}
.notice {
	margin-bottom: 16px;
}
.form-grid,
.selection-grid {
	display: grid;
	align-items: start;
	grid-template-columns: repeat(2, minmax(0, 1fr));
	gap: 0 28px;
}
.selection-grid {
	gap: 24px;
}
.group-picker {
	display: flex;
	align-items: center;
	width: 100%;
}
.field-stack {
	width: 100%;
	min-width: 0;
	display: flex;
	flex-direction: column;
	gap: 10px;
}
.sender-order-list {
	max-height: 360px;
	overflow-y: auto;
}
.invalid-input {
	color: #b94040;
	font-size: 13px;
	overflow-wrap: anywhere;
}
.repeat-dialog {
	width: min(660px, calc(100vw - 32px));
	max-height: calc(100vh - 48px);
	overflow-y: auto;
}
.repeat-dialog p {
	margin-bottom: 12px;
}
.repeat-recipients {
	max-height: 200px;
	overflow-y: auto;
	overflow-wrap: anywhere;
}
.order-row {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 8px 0;
	min-width: 0;
}
.order-number {
	flex-shrink: 0;
	width: 26px;
	color: #238353;
	font-weight: 600;
}
.order-text {
	flex: 1;
	min-width: 0;
	overflow-wrap: anywhere;
}
.template-row {
	margin-bottom: 12px;
}
@media (max-width: 850px) {
	.rotation-page {
		padding: 16px;
	}
	.form-grid,
	.selection-grid {
		grid-template-columns: 1fr;
	}
}
</style>
