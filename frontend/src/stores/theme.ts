import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { darkTheme, useOsTheme } from 'naive-ui'

export type ThemeMode = 'system' | 'light' | 'dark'

const STORAGE_KEY = 'bsm.theme'

function loadMode(): ThemeMode {
  try {
    const v = localStorage.getItem(STORAGE_KEY)
    if (v === 'light' || v === 'dark' || v === 'system') return v
  } catch {
    /* localStorage が使えない環境では system 扱い */
  }
  return 'system'
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(loadMode())
  const osTheme = useOsTheme()

  const isDark = computed(
    () => mode.value === 'dark' || (mode.value === 'system' && osTheme.value === 'dark'),
  )
  const theme = computed(() => (isDark.value ? darkTheme : null))

  watch(mode, (v) => {
    try {
      localStorage.setItem(STORAGE_KEY, v)
    } catch {
      /* ignore */
    }
  })

  return { mode, isDark, theme }
})
