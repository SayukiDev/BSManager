<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NFlex,
  NForm,
  NFormItem,
  NInputNumber,
  NRadioButton,
  NRadioGroup,
  NSelect,
  NSwitch,
  NText,
} from 'naive-ui'
import type { settings as cfg } from '../../wailsjs/go/models'
import * as api from '../api/app'
import { getMessage } from '../plugins/discrete'
import { useAppStore } from '../stores/app'
import { useThemeStore, type ThemeMode } from '../stores/theme'

const app = useAppStore()
const theme = useThemeStore()

/** フォームで編集する項目のみ。devices / setuped は保存時に現在値を引き継ぐ */
interface Form {
  followSteamVR: boolean
  checkTrackerConnected: boolean
  minTrackerCount: number | null
  checkInterval: number | null
  shutdownWaiting: number | null
  logLevel: string
}

function fromSettings(s: cfg.Content | null): Form {
  return {
    followSteamVR: s?.followSteamVR ?? true,
    checkTrackerConnected: s?.checkTrackerConnected ?? true,
    minTrackerCount: s?.minTrackerCount ?? 2,
    checkInterval: s?.checkInterval ?? 5,
    shutdownWaiting: s?.shutdownWaiting ?? 60,
    logLevel: s?.logLevel ?? 'info',
  }
}

const form = reactive<Form>(fromSettings(app.settings))
const saving = ref(false)
const logPath = ref('')

const dirty = computed(() => JSON.stringify(form) !== JSON.stringify(fromSettings(app.settings)))

// 他所（settings:changed）で設定が変わったとき、ローカル編集が無ければフォームを追従させる
watch(
  () => app.settings,
  (next, prev) => {
    if (JSON.stringify(form) === JSON.stringify(fromSettings(prev ?? null))) {
      Object.assign(form, fromSettings(next))
    }
  },
)

const logLevelOptions = ['debug', 'info', 'warn', 'error', 'fatal'].map((v) => ({ label: v, value: v }))
const themeOptions: { label: string; value: ThemeMode }[] = [
  { label: 'システム', value: 'system' },
  { label: 'ライト', value: 'light' },
  { label: 'ダーク', value: 'dark' },
]

function reset() {
  Object.assign(form, fromSettings(app.settings))
}

async function save() {
  if (!app.settings) return
  saving.value = true
  try {
    await app.saveSettings({
      ...app.settings,
      followSteamVR: form.followSteamVR,
      checkTrackerConnected: form.checkTrackerConnected,
      minTrackerCount: form.minTrackerCount ?? 0,
      checkInterval: form.checkInterval ?? 1,
      shutdownWaiting: form.shutdownWaiting ?? 0,
      logLevel: form.logLevel,
    })
    getMessage().success('設定を保存しました')
  } catch {
    /* ストア側でトースト済み */
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    logPath.value = await api.GetLogPath()
  } catch {
    /* ignore */
  }
})
</script>

<template>
  <NForm label-placement="left" label-width="auto" :show-feedback="false">
    <NFlex vertical :size="16">
      <NCard title="動作">
        <NFlex vertical :size="12">
          <NFormItem label="SteamVR に追従">
            <NSwitch v-model:value="form.followSteamVR" />
          </NFormItem>
          <NAlert v-if="form.followSteamVR" type="info" :show-icon="false">
            有効時はホームの電源スイッチは無効になり、SteamVR の起動状態に合わせて自動で切り替わります
          </NAlert>
          <NFormItem label="トラッカー接続を確認">
            <NSwitch v-model:value="form.checkTrackerConnected" />
          </NFormItem>
          <NFormItem label="最小トラッカー数">
            <NInputNumber
              v-model:value="form.minTrackerCount"
              :min="0"
              :disabled="!form.checkTrackerConnected"
              style="width: 160px"
            >
              <template #suffix>台</template>
            </NInputNumber>
          </NFormItem>
          <NFormItem label="チェック間隔">
            <NInputNumber v-model:value="form.checkInterval" :min="1" style="width: 160px">
              <template #suffix>秒</template>
            </NInputNumber>
          </NFormItem>
          <NFormItem label="電源オフまでの遅延">
            <NInputNumber v-model:value="form.shutdownWaiting" :min="0" style="width: 160px">
              <template #suffix>秒</template>
            </NInputNumber>
          </NFormItem>
        </NFlex>
      </NCard>

      <NCard title="ログ">
        <NFlex vertical :size="12">
          <NFormItem label="ログレベル">
            <NSelect v-model:value="form.logLevel" :options="logLevelOptions" style="width: 160px" />
          </NFormItem>
          <NText :depth="3" style="font-size: 12px">ログファイル: {{ logPath || '取得できません' }}</NText>
        </NFlex>
      </NCard>

      <NCard title="外観">
        <NFormItem label="テーマ">
          <NRadioGroup v-model:value="theme.mode">
            <NRadioButton v-for="o in themeOptions" :key="o.value" :value="o.value" :label="o.label" />
          </NRadioGroup>
        </NFormItem>
      </NCard>

      <NFlex justify="end" :size="8">
        <NButton :disabled="!dirty || saving" @click="reset">元に戻す</NButton>
        <NButton type="primary" :disabled="!dirty" :loading="saving" @click="save">保存</NButton>
      </NFlex>
    </NFlex>
  </NForm>
</template>
