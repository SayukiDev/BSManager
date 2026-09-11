<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { NConfigProvider, NFlex, NGlobalStyle, NSpin, dateJaJP, jaJP } from 'naive-ui'
import MainLayout from './layouts/MainLayout.vue'
import SetupView from './views/SetupView.vue'
import { applyWindowSize, MAIN_SIZE } from './composables/useWindow'
import { useAppStore } from './stores/app'
import { useThemeStore } from './stores/theme'
import { themeOverrides } from './theme/overrides'

const POLL_INTERVAL_MS = 3000

const theme = useThemeStore()
const app = useAppStore()

/** null = 初期ロード中。初回ロード後に一度だけ決めるので、セットアップ完了画面が途中で消えない */
const showSetup = ref<boolean | null>(null)

onMounted(async () => {
  app.bindEvents()
  await app.refreshAll()
  showSetup.value = !app.isSetuped
  if (!showSetup.value) app.startPolling(POLL_INTERVAL_MS)
})

onBeforeUnmount(() => {
  app.stopPolling()
  app.unbindEvents()
})

function onSetupDone() {
  showSetup.value = false
  void applyWindowSize(MAIN_SIZE)
  app.startPolling(POLL_INTERVAL_MS)
}
</script>

<template>
  <NConfigProvider
    :theme="theme.theme"
    :theme-overrides="themeOverrides"
    :locale="jaJP"
    :date-locale="dateJaJP"
  >
    <NGlobalStyle />
    <NFlex v-if="showSetup === null" justify="center" align="center" style="height: 100vh">
      <NSpin size="large" />
    </NFlex>
    <SetupView v-else-if="showSetup" @done="onSetupDone" />
    <MainLayout v-else />
  </NConfigProvider>
</template>
