import { useUserStore, useGlobalStore } from '@/store'
import { webBasePath } from '@/utils/base'

let bootstrap: Promise<void> | undefined

export function ensurePlatformLogin(): Promise<void> {
 if (!webBasePath) return Promise.resolve()
 if (bootstrap) return bootstrap
 bootstrap = (async () => {
  const fragment = new URLSearchParams(window.location.hash.slice(1))
  const platformToken = fragment.get('host_token') || ''
  if (platformToken) window.history.replaceState(null, '', window.location.pathname + window.location.search)
  const response = await fetch(`${webBasePath}/api/platform/session`, {
   method: 'POST', credentials: 'same-origin',
   signal: AbortSignal.timeout(90000),
   headers: { 'Content-Type': 'application/json' },
   body: JSON.stringify({ platformToken }),
  })
  if (response.status === 401 || response.status === 403 || response.redirected) {
   useUserStore().resetLoginInfo()
   window.location.replace('/apps/launch/mail')
   throw new Error('平台会话已过期，正在重新进入')
  }
  const result = await response.json()
  if (!response.ok || result.code !== 0 || !result.data?.token) {
   throw new Error(result.msg || '个人邮箱空间暂不可用，请刷新重试')
  }
  useUserStore().setLoginInfo({ token: result.data.token, refresh_token: '', ttl: result.data.ttl })
  if (result.data.accountInfo?.lang) useGlobalStore().lang = result.data.accountInfo.lang
 })()
 return bootstrap
}
