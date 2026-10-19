import { instance } from '@/api'

export interface RotationConfig {
	name: string
	senders: string[]
	templates: { id: number; subject: string }[]
	group_id: number
	recipient_source?: 'group' | 'paste'
	recipient_text?: string
	include_sent?: boolean
	accepted_history?: Record<string, string>
	per_sender: number
	interval: number
	full_name: string
}
export interface RotationItem {
	sequence: number
	recipient: string
	sender: string
	template_id: number
	template_name: string
	subject: string
	task_id: number
	state: string
	is_sent: number
	previous_sent?: 'same' | 'other' | 'unknown' | 'confirmed'
}
export interface RotationPreview {
	config: RotationConfig
	items: RotationItem[]
	skipped: Record<string, number>
	hash: string
	input?: { total: number; emails: string[]; duplicates: number; invalid: string[] }
	repeated: {
		count: number
		same: number
		other: number
		unknown: number
		planned_total: number
		recipients: { email: string; kind: 'same' | 'other' | 'unknown' }[] | null
	}
}
export interface RotationJob {
	id: number
	name: string
	status: string
	total: number
	accepted: number
	failed: number
	created_at: number
	error: string
	config: string
}
export interface RotationOptions {
	accounts: { id: number; email: string }[]
	templates: { id: number; temp_name: string }[]
	groups: { id: number; name: string; count: number }[]
}
export const rotationOptions = () =>
	instance.get<unknown, RotationOptions>('/native_rotation/options')
export const rotationPreview = (config: RotationConfig) =>
	instance.post<unknown, RotationPreview>('/native_rotation/preview', { config })
export const rotationCreate = (config: RotationConfig, preview_hash: string, request_key: string) =>
	instance.post<unknown, { id: number }>(
		'/native_rotation/create',
		{ config, preview_hash, request_key },
		{ fetchOptions: { successMessage: true } }
	)
export const rotationList = () => instance.get<unknown, RotationJob[]>('/native_rotation/list')
export const rotationDetail = (id: number) =>
	instance.get<unknown, { job: RotationJob; items: RotationItem[] }>('/native_rotation/detail', {
		params: { id },
	})
export const rotationAction = (id: number, action: string) =>
	instance.post(
		'/native_rotation/action',
		{ id, action },
		{ fetchOptions: { successMessage: true } }
	)
