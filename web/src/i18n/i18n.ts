import type { App } from 'vue'
import { createI18n } from 'vue-i18n'

import { useThemeStore } from '@/store'
import en from './en.json'
import zh_CN from './zh_CN.json'

let i18n: ReturnType<typeof createI18n>

export function setupI18n(app: App) {
  const themeStore = useThemeStore()
  i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: themeStore.locale,
    missingWarn: false, // TODO 完成 i18n 之后需要去除
    fallbackWarn: false, // TODO 完成 i18n 之后需要去除
    fallbackLocale: 'zh_CN',
    messages: {
      en: en,
      zh_CN: zh_CN
    }
  })
  app.use(i18n)
}
