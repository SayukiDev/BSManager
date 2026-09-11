<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NAlert, NButton, NCard, NFlex, NH2, NIcon, NStatistic, NSwitch, NTag, NText, NTooltip } from 'naive-ui'
import { BluetoothOutline, RadioOutline, SyncOutline } from '@vicons/ionicons5'
import InfoCard from '../components/InfoCard.vue'
import PowerStateTag from '../components/PowerStateTag.vue'
import { useAppStore } from '../stores/app'
import { useUiStore } from '../stores/ui'

const app = useAppStore()
const ui = useUiStore()

const status = computed(() => app.status)
const settings = computed(() => app.settings)
const powerBusy = ref(false)
const syncBusy = ref(false)

async function togglePower(on: boolean) {
  powerBusy.value = true
  try {
    await app.setPower(on)
  } catch {
    /* ストア側でトースト済み */
  } finally {
    powerBusy.value = false
  }
}

async function syncPower() {
  syncBusy.value = true
  try {
    await app.syncPowerState()
  } catch {
    /* ストア側でトースト済み */
  } finally {
    syncBusy.value = false
  }
}

onMounted(() => {
  if (app.enabledDevices.length) void app.refreshPowerState()
})
</script>

<template>
  <NFlex vertical :size="16">
    <NCard>
      <NFlex align="center" :size="24" :wrap="false">
        <NIcon :size="120" :depth="3" :component="RadioOutline" style="flex-shrink: 0; margin: 0 24px" />
        <NFlex vertical :size="6" style="flex: 1; min-width: 0">
          <NH2 style="margin: 0">ベースステーション</NH2>
          <NText :depth="3">
            電源 {{ status?.powerOn ? 'ON' : 'OFF' }}
            <template v-if="app.powerChanging">（ベースステーションに反映中…）</template>
          </NText>
          <NText :depth="3">SteamVR {{ status?.steamVRRunning ? '起動中' : '停止中' }}</NText>
          <NText v-if="settings?.checkTrackerConnected && status?.steamVRRunning" :depth="3">
            トラッカー {{ status?.trackerCount ?? 0 }} 台
          </NText>
          <NText :depth="3">マネージャー {{ status?.managerRunning ? '動作中' : '停止中' }}</NText>
        </NFlex>
        <NTooltip :disabled="!app.followSteamVR" trigger="hover">
          <template #trigger>
            <!-- disabled 要素は hover を拾わないので span で包む -->
            <span style="margin-right: 24px">
              <NSwitch
                size="large"
                :value="status?.powerOn ?? false"
                :disabled="app.followSteamVR || !app.enabledDevices.length || app.powerChanging"
                :loading="powerBusy || app.powerChanging"
                @update:value="togglePower"
              >
                <template #checked>電源 ON</template>
                <template #unchecked>電源 OFF</template>
              </NSwitch>
            </span>
          </template>
          SteamVR 追従が有効なため手動切替はできません（設定で変更できます）
        </NTooltip>
      </NFlex>
    </NCard>

    <NAlert v-if="!app.devices.length" type="warning" title="ベースステーションが未登録です">
      <NButton text type="primary" @click="ui.page = 'stations'">ベースステーションページで登録する</NButton>
    </NAlert>

    <InfoCard title="登録済みベースステーション" :icon="BluetoothOutline">
      <NFlex v-for="d in app.devices" :key="d.addr" align="center" :size="8">
        <NText :depth="3">{{ d.name || '(名前なし)' }}</NText>
        <NText :depth="3" code>{{ d.addr }}</NText>
        <PowerStateTag v-if="d.enable" :state="app.powerStates[d.addr]" />
        <NTag v-else size="small" :bordered="false">制御対象外</NTag>
      </NFlex>
      <NText v-if="!app.devices.length" :depth="3">登録されているデバイスはありません</NText>
      <template #aside>
        <NStatistic label="制御対象" :value="app.enabledDevices.length">
          <template #suffix>/ {{ app.devices.length }} 台</template>
        </NStatistic>
        <NButton
          text
          size="small"
          :loading="app.powerStatesLoading"
          :disabled="!app.enabledDevices.length || app.powerChanging"
          @click="app.refreshPowerState()"
        >
          状態を更新
        </NButton>
        <NText v-if="app.powerChanging" :depth="3" style="font-size: 12px">反映中…</NText>
        <NTooltip v-if="app.hasPowerMismatch" trigger="hover">
          <template #trigger>
            <NButton
              text
              size="small"
              type="warning"
              :loading="syncBusy || app.powerChanging"
              :disabled="app.powerStatesLoading || app.powerChanging"
              style="margin-top: 8px"
              @click="syncPower"
            >
              作動状態を同期
            </NButton>
          </template>
          マネージャーは電源 {{ status?.powerOn ? 'ON' : 'OFF' }} ですが、{{ app.powerMismatchDevices.length }} 台の状態が一致していません
        </NTooltip>
      </template>
    </InfoCard>

    <InfoCard title="動作モード" :icon="SyncOutline">
      <NText :depth="3">
        {{ app.followSteamVR ? 'SteamVR の起動状態に追従して電源を制御します' : '手動で電源を切り替えます' }}
      </NText>
      <NText :depth="3">
        トラッカー接続確認:
        {{ settings?.checkTrackerConnected ? `有効（最小 ${settings?.minTrackerCount ?? 0} 台）` : '無効' }}
      </NText>
      <NText :depth="3">電源オフまでの遅延: {{ settings?.shutdownWaiting ?? 0 }} 秒</NText>
      <template #aside>
        <NTag :type="app.followSteamVR ? 'primary' : 'default'" :bordered="false">
          {{ app.followSteamVR ? 'SteamVR 追従' : '手動' }}
        </NTag>
        <NStatistic label="チェック間隔" :value="settings?.checkInterval ?? 0">
          <template #suffix>秒</template>
        </NStatistic>
        <NButton text size="small" @click="ui.page = 'settings'">設定を開く</NButton>
      </template>
    </InfoCard>
  </NFlex>
</template>
