import { useGlobalStore, useUserStore } from '@/store'
import { setLanguage } from '@/i18n'
import { clearPendingRequests } from '@/api'
import { routes } from '@/router/router'
import router from '@/router/router'
import loadingBar from '@/config/loadingBar'
import { ensurePlatformLogin } from '@/utils/platformLogin'

// Route white list
const whitePathList = ['/login']

router.beforeEach(async (to, from, next) => {
	loadingBar.start()

	clearPendingRequests()
 try {
  await ensurePlatformLogin()
 } catch (error) {
  loadingBar.finish()
  const panel = document.createElement('div')
  panel.style.cssText = 'padding:32px;font-size:16px;line-height:1.8'
  panel.textContent = error instanceof Error ? error.message : '个人邮箱空间暂不可用，请刷新重试'
  const retry = document.createElement('button')
  retry.textContent = '重新进入'
  retry.onclick = () => window.location.replace('/apps/launch/mail')
  panel.appendChild(document.createElement('br'))
  panel.appendChild(retry)
  document.getElementById('mail-loading')?.remove()
  document.getElementById('root')?.replaceChildren(panel)
  next(false)
  return
 }

	const globalStore = useGlobalStore()

	// Set the language
	try {
		await globalStore.getLang()
		setLanguage(globalStore.lang)
	} catch {
		setLanguage(globalStore.lang)
	}

	// Check if the visited route exists in the registered routes
	const routeExists = routes.some(route => route.path === to.path)

	// If the route does not exist, go directly
	if (!routeExists) {
		next()
		return
	}

	const userStore = useUserStore()

	// User is logged in
	if (userStore.isLogin) {
		// If the visited route is in the white list, jump to the home page
		if (whitePathList.includes(to.path)) {
			next('/')
		} else {
			next()
		}
	} else if (whitePathList.includes(to.path)) {
		// If the visited route is in the white list, go directly
		next()
	} else {
		next('/login')
	}
})

router.afterEach(() => {
	loadingBar.finish()
	document.getElementById('mail-loading')?.remove()
})

export default router
