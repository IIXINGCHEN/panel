import '@/styles/index.scss'
import '@/styles/reset.css'
import '@vue-js-cron/naive-ui/dist/naive-ui.css'
import 'virtual:uno.css'

import { createApp } from 'vue'
import App from './App.vue'

import { setupI18n } from '@/i18n/i18n'
import { setupRouter } from '@/router'
import { setupStore, useThemeStore } from '@/store'
import { createGettext, setupNaiveDiscreteApi } from '@/utils'

import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor'

import dashboard from '@/api/panel/dashboard'
import CronNaivePlugin from '@vue-js-cron/naive-ui'

async function setupApp() {
  const app = createApp(App)
  app.use(VueMonacoEditorPlugin, {
    paths: {
      vs: window.location.origin + '/assets/vs'
    },
    'vs/nls': {
      availableLanguages: { '*': 'zh-cn' }
    }
  })
  app.use(CronNaivePlugin)
  await setupStore(app)
  await setupNaiveDiscreteApi()
  await setupPanel().then(() => {
    app.use(createGettext)
    setupI18n(app)
  })
  await setupRouter(app)
  app.mount('#app')
}

const setupPanel = async () => {
  const themeStore = useThemeStore()
  useRequest(dashboard.panel, {
    initialData: {
      name: import.meta.env.VITE_APP_TITLE,
      locale: 'zh_CN'
    }
  }).onSuccess(({ data }: { data: any }) => {
    themeStore.setLocale(data.locale)
    themeStore.setName(data.name)
  })
}

setupApp()
