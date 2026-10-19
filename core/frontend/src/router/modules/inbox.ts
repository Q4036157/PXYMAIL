import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
 path: '/inbox', component: Layout,
 meta: {key: 'inbox', title: '收件箱 / 客户回复'},
 children: [{path: '/inbox', name: 'Inbox', component: () => import('@/views/inbox/index.vue')}],
}
export default route
