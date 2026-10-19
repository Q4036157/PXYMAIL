import { instance } from '@/api'

export interface InboxAccount {
	id: number
	email: string
	remark: string
	enabled: boolean
	status: string
	last_error: string
	last_attempt: number
	last_success: number
	next_poll: number
	mailbox_total: number
	cached_count: number
}
export interface InboxCheckResult {
	account_id: number
	checked_at: number
	new_count: number
	cached_count: number
	reply_count: number
}
export interface InboxMail {
	id: number
	account_id: number
	account: string
	from_address: string
	from_name: string
	subject: string
	received_at: number
	kind: string
	processed: boolean
	seen: boolean
	body_cached: boolean
	body_text?: string
	body_html?: string
	attachments?: string
	truncated: boolean
}
export const inboxAccounts = () =>
	instance.get<InboxAccount[], InboxAccount[]>('/native_inbox/accounts')
export const inboxList = (params: Record<string, unknown>) =>
	instance.get<unknown, { list: InboxMail[]; total: number }>('/native_inbox/list', { params })
export const inboxDetail = (id: number) =>
	instance.get<InboxMail, InboxMail>('/native_inbox/detail', { params: { id } })
export const inboxSync = (account_id = 0) =>
	instance.post('/native_inbox/sync', { account_id }, { fetchOptions: { successMessage: true } })
export const inboxCheckNow = (account_id: number) =>
	instance.post<unknown, InboxCheckResult>('/native_inbox/check_now', { account_id })
export const inboxToggle = (account_id: number, enabled: boolean) =>
	instance.post('/native_inbox/update', { account_id, enabled })
export const inboxMark = (id: number, processed: boolean) =>
	instance.post('/native_inbox/mark', { id, processed })
