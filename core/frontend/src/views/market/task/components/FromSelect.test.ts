import { describe, expect, it } from 'vitest'
import { buildSenderOptions } from './sender-options'

const mailbox = (username: string, domain: string) =>
	({ username, domain, full_name: username }) as any
const smtp = (overrides: Record<string, unknown>) =>
	({
		active: 1,
		auth_user: 'smtp-user',
		sender_email: '',
		smtp_name: 'Brevo',
		remark: 'Brevo SMTP',
		...overrides,
	}) as any

describe('FromSelect sender options', () => {
	it('uses sender_email for active SMTP and keeps auth_user separate', () => {
		const options = buildSenderOptions(
			[],
			[smtp({ auth_user: 'login', sender_email: 'support@zzcoin.cc' })],
			'zzcoin.cc'
		)
		expect(options.map(option => option.value)).toEqual(['support@zzcoin.cc'])
	})

	it('filters inactive SMTP and invalid effective sender addresses', () => {
		const options = buildSenderOptions(
			[],
			[smtp({ active: 0, sender_email: 'inactive@example.com' }), smtp({ auth_user: 'apikey' })],
			'example.com'
		)
		expect(options).toHaveLength(0)
	})

	it('keeps mailbox domain filtering, merges SMTP without a local domain, and deduplicates', () => {
		const options = buildSenderOptions(
			[mailbox('local@example.com', 'example.com'), mailbox('other@other.com', 'other.com')],
			[smtp({ auth_user: 'local@example.com' }), smtp({ sender_email: 'support@zzcoin.cc' })],
			'example.com'
		)
		expect(options.map(option => option.value)).toEqual(['local@example.com'])
	})

	it('only shows SMTP senders for the selected domain', () => {
		const options = buildSenderOptions(
			[],
			[smtp({ sender_email: 'support@zzcoin.cc' })],
			'other.com'
		)
		expect(options).toHaveLength(0)
	})
})
