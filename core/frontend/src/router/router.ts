import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { is, isDev, webBasePath } from '@/utils'

// Routes reflect list
const routesReflectList = [
	'Overview',
	'Email Marketing',
	'收件箱 / 客户回复',
	'template',
	'Send API',
	'Contacts',
	'Sequences',
	'Leads',
	'Enrichment',
	'MailDomain',
	'MailBoxes',
	'SMTP',
	'Logs',
	'Settings',
	'Automation',
	'Video Outreach',
]

// Import routes from modules
const modules = import.meta.webpackContext('./modules', {
	// Whether to search for subdirectories
	recursive: false,
	// Context keys include the leading "./" (for example "./overview.ts").
	regExp: /^\.\/[^/.]+\.ts$/,
})

// Module routes
export let menuList: RouteRecordRaw[] = []

// Iterate through the module list to generate module routes
for (const path of modules.keys()) {
	const mod = modules(path)
	if (is<{ default: RouteRecordRaw }>(mod, 'Module')) {
		menuList.push(mod.default)
	}
}

// Sort module routes
menuList.sort((a, b) => {
	const rank = (route: RouteRecordRaw) => {
		const index = routesReflectList.findIndex(item => item === route.meta?.title)
		return index < 0 ? routesReflectList.length : index
	}
	return rank(a) - rank(b)
})

const otherArray: RouteRecordRaw[] = []

if (isDev) {
	otherArray.push({
		path: '/test',
		name: 'Test',
		component: () => import('@/views/test/index.vue'),
	})
}

export const routes: RouteRecordRaw[] = [
	{
		path: '/login',
		name: 'Login',
		component: () => import('@/views/login/index.vue'),
	},
	{
		path: '/',
		redirect: '/overview',
	},
	...menuList,
	...otherArray,
]

const router = createRouter({
	history: createWebHistory(webBasePath + '/'),
	routes,
	strict: false,
	scrollBehavior: () => ({ left: 0, top: 0 }),
})

export default router
