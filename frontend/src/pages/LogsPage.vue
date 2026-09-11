<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { NButton, NCard, NFlex, NLog, NSelect, NSwitch, NText, type LogInst } from 'naive-ui'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import * as api from '../api/app'
import { getMessage } from '../plugins/discrete'

const AUTO_REFRESH_MS = 2000

const lineOptions = [100, 200, 500, 1000].map((n) => ({ label: `${n} 行`, value: n }))
const lineCount = ref(200)
const lines = ref<string[]>([])
const logPath = ref('')
const autoRefresh = ref(false)
const loading = ref(false)
const logRef = ref<LogInst | null>(null)
let timer: number | null = null

async function refresh() {
  loading.value = true
  try {
    lines.value = (await api.ReadLogs(lineCount.value)) ?? []
    await nextTick()
    logRef.value?.scrollTo({ position: 'bottom', silent: true })
  } catch (e) {
    getMessage().error(api.toError(e))
  } finally {
    loading.value = false
  }
}

function openLogFolder() {
  const dir = logPath.value.replace(/[\\/][^\\/]*$/, '')
  if (!dir) return
  try {
    BrowserOpenURL('file:///' + dir.replace(/\\/g, '/'))
  } catch (e) {
    getMessage().error(api.toError(e))
  }
}

watch(autoRefresh, (on) => {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
  if (on) timer = window.setInterval(() => void refresh(), AUTO_REFRESH_MS)
})
watch(lineCount, () => void refresh())

onMounted(async () => {
  try {
    logPath.value = await api.GetLogPath()
  } catch {
    /* ignore */
  }
  await refresh()
})

onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<template>
  <NFlex vertical :size="16">
    <NCard>
      <template #header>
        <NFlex align="center" :size="12" :wrap="false">
          <NSelect v-model:value="lineCount" :options="lineOptions" size="small" style="width: 120px" />
          <NButton size="small" :loading="loading" @click="refresh">更新</NButton>
          <NSwitch v-model:value="autoRefresh" size="small" />
          <NText :depth="3" style="font-size: 13px">自動更新（{{ AUTO_REFRESH_MS / 1000 }} 秒）</NText>
        </NFlex>
      </template>
      <template #header-extra>
        <NButton text size="small" :disabled="!logPath" @click="openLogFolder">ログフォルダを開く</NButton>
      </template>
      <NFlex vertical :size="8">
        <NText :depth="3" code style="font-size: 12px">{{ logPath || 'ログファイルのパスを取得できません' }}</NText>
        <NLog ref="logRef" :lines="lines" :rows="24" style="font-size: 12px" />
      </NFlex>
    </NCard>
  </NFlex>
</template>
