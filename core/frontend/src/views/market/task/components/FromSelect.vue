<template>
	<div class="flex-1">
		<n-select
			v-model:value="domain"
			:loading="loading"
			:options="domainOptions"
			@update:value="handleUpdateDomain">
		</n-select>
	</div>
	<div class="flex-1 ml-10px">
		<n-select
			v-model:value="sender"
			:loading="loading"
			:options="senderOptions"
			:filterable="true"
			@update:value="handleUpdateSender">
		</n-select>
	</div>
</template>

<script lang="ts" setup>
import { SelectOption } from 'naive-ui'
import { isArray, isObject } from '@/utils'
import { getDomainAll } from '@/api/modules/domain'
import { getMailboxList } from '@/api/modules/mailbox'
import { getSmtpList } from '@/api/modules/smtp'
import type { MailDomain } from '@/views/domain/interface'
import type { MailBox } from '@/views/mailbox/interface'
import type { SmtpService } from '@/views/smtp/types/base'
import { buildSenderOptions } from './sender-options'

const sender = defineModel<string | null>('value')

const name = defineModel<string>('name')

const loading = ref<boolean>(false)

const domain = defineModel<string | null>('domain')

const domainOptions = ref<SelectOption[]>([])

const mailboxList = ref<MailBox[]>([])
const smtpList = ref<SmtpService[]>([])

const senderOptions = computed(() =>
	buildSenderOptions(mailboxList.value, smtpList.value, domain.value)
)

const handleUpdateDomain = async () => {
	await nextTick()
	if (senderOptions.value.length > 0) {
		sender.value = senderOptions.value[0].value
		handleUpdateSender(sender.value, senderOptions.value[0])
	} else {
		sender.value = null
		name.value = ''
	}
}

const handleUpdateSender = (val: string, option: SelectOption) => {
	const { data } = option
	if (isObject<MailBox>(data) && 'full_name' in data) {
		name.value = data.full_name
	} else if (isObject<SmtpService>(data)) {
		name.value = data.smtp_name || data.remark || ''
	}
}

// 获取域名列表
const getDomainOptions = async () => {
	const domainRes = await getDomainAll()
	if (isArray<MailDomain>(domainRes)) {
		const domains = new Map(domainRes.map(item => [item.domain.toLowerCase(), item]))
		for (const smtp of smtpList.value) {
			const email = smtp.sender_email || smtp.auth_user
			const smtpDomain = email.split('@')[1]?.toLowerCase()
			if (smtp.active && smtpDomain && !domains.has(smtpDomain)) {
				domains.set(smtpDomain, { domain: smtpDomain, default: false } as MailDomain)
			}
		}
		domainOptions.value = [...domains.values()].map(item => ({
			label: item.domain,
			value: item.domain,
		}))
		if (!domain.value && domainRes.length > 0) {
			const domainData = domainRes.filter(item => item.default)
			if (domainData.length > 0) {
				domain.value = domainData[0].domain
			} else {
				domain.value = domainRes[0].domain
			}
		}
	}
}

// 获取邮箱列表
const getSenderOptions = async () => {
	const [mailboxRes, smtpRes] = await Promise.all([
		getMailboxList({ page: 1, page_size: 1000, domain: '' }),
		getSmtpList(),
	])
	if (isObject<{ list: MailBox[] }>(mailboxRes)) mailboxList.value = mailboxRes.list
	if (isObject<{ list: SmtpService[] }>(smtpRes)) smtpList.value = smtpRes.list
}

const initData = async () => {
	try {
		loading.value = true
		await getSenderOptions()
		await getDomainOptions()
		await nextTick()
		if (sender.value === null) {
			handleUpdateDomain()
		} else if (sender.value) {
			domain.value = sender.value.split('@')[1]
		}
	} finally {
		loading.value = false
	}
}

initData()
</script>

<style lang="scss" scoped></style>
