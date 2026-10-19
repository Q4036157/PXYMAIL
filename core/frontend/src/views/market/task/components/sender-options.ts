import type { SelectOption } from 'naive-ui'
import type { MailBox } from '@/views/mailbox/interface'
import type { SmtpService } from '@/views/smtp/types/base'

export const buildSenderOptions = (
	mailboxes: MailBox[],
	smtps: SmtpService[],
	selectedDomain: string | null
) => {
	const localOptions = mailboxes
		.filter(item => item.domain.toLowerCase() === (selectedDomain || '').toLowerCase())
		.map(item => ({ label: item.username, value: item.username, data: item }))
	const smtpOptions = smtps
		.filter(item => item.active)
		.map(item => {
			const email = item.sender_email || item.auth_user
			const domain = email.split('@')[1]?.toLowerCase()
			if (!domain || domain !== (selectedDomain || '').toLowerCase()) return null
			if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return null
			return { label: `${email}（SMTP）`, value: email, data: item }
		})
	const seen = new Set<string>()
	return [...localOptions, ...smtpOptions]
		.filter((option): option is SelectOption => Boolean(option))
		.filter(option => {
			const value = String(option.value)
			if (seen.has(value)) return false
			seen.add(value)
			return true
		})
}
