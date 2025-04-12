import type { RouteType } from '~/types/router'
import { $gettext } from '@/utils/gettext'

const Layout = () => import('@/layout/IndexView.vue')

export default {
  name: 'ssh',
  path: '/ssh',
  component: Layout,
  meta: {
    order: 70
  },
  children: [
    {
      name: 'ssh-index',
      path: '',
      component: () => import('./IndexView.vue'),
      meta: {
        title: $gettext('Terminal'),
        icon: 'mdi:console',
        role: ['admin'],
        requireAuth: true
      }
    }
  ]
} as RouteType
