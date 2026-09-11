import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type PageKey = 'home' | 'stations' | 'logs' | 'settings'

const SIDER_KEY = 'bsm.sider'

function loadCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDER_KEY) === '1'
  } catch {
    return false
  }
}

export const useUiStore = defineStore('ui', () => {
  const page = ref<PageKey>('home')
  const siderCollapsed = ref(loadCollapsed())

  watch(siderCollapsed, (v) => {
    try {
      localStorage.setItem(SIDER_KEY, v ? '1' : '0')
    } catch {
      /* ignore */
    }
  })

  return { page, siderCollapsed }
})
