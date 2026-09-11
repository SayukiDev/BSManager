<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { NAlert, NButton, NCard, NFlex, NResult, NSpin, NStep, NSteps, NText } from 'naive-ui'
import TitleBar from '../components/TitleBar.vue'
import ScanResultTable from '../components/ScanResultTable.vue'
import { applyWindowSize, SETUP_SIZE } from '../composables/useWindow'
import { useAppStore } from '../stores/app'
import { useScanStore } from '../stores/scan'

const emit = defineEmits<{ done: [] }>()

const app = useAppStore()
const scan = useScanStore()
const step = ref(1)

onMounted(() => {
  void applyWindowSize(SETUP_SIZE)
  scan.bindEvents()
  void scan.sync()
})

onBeforeUnmount(() => {
  if (scan.scanning) void scan.stop()
})

async function confirm() {
  try {
    await scan.confirm()
    await app.refreshAll()
    step.value = 2
  } catch {
    /* ストア側でトースト済み */
  }
}
</script>

<template>
  <div style="display: flex; flex-direction: column; height: 100vh">
    <TitleBar />
    <NFlex justify="center" style="flex: 1; overflow: auto; padding: 24px">
      <NCard title="初期セットアップ" style="max-width: 640px; width: 100%; align-self: flex-start">
        <NFlex vertical :size="16">
          <NSteps :current="step" size="small">
            <NStep title="検出と選択" description="ベースステーションを検出して登録" />
            <NStep title="完了" />
          </NSteps>

          <template v-if="step === 1">
            <NText :depth="3">
              ベースステーションの電源を入れて検出されるのを待ち、登録するものにチェックを入れて確定してください。
            </NText>
            <NFlex :size="8">
              <NButton type="primary" :disabled="scan.scanning" @click="scan.start()">スキャン開始</NButton>
              <NButton :disabled="!scan.scanning" @click="scan.stop()">停止</NButton>
            </NFlex>
            <NAlert v-if="scan.scanning" type="info" :show-icon="false">
              <NFlex align="center" :size="8" :wrap="false">
                <NSpin :size="16" />
                <span>検出中… 名前順に一覧へ追加されます</span>
              </NFlex>
            </NAlert>
            <ScanResultTable v-model:checked-keys="scan.checked" :rows="scan.results" hide-state />
            <NFlex justify="end">
              <NButton type="primary" :disabled="!scan.checked.length" :loading="scan.confirming" @click="confirm">
                確定して登録
              </NButton>
            </NFlex>
          </template>

          <NResult
            v-else
            status="success"
            title="セットアップ完了"
            description="ベースステーションを登録しました。SteamVR の起動に合わせて電源が制御されます。"
          >
            <template #footer>
              <NButton type="primary" @click="emit('done')">ホームへ</NButton>
            </template>
          </NResult>
        </NFlex>
      </NCard>
    </NFlex>
  </div>
</template>
