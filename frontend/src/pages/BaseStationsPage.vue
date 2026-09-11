<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NEmpty,
  NFlex,
  NPopconfirm,
  NSpin,
  NSwitch,
  NText,
  NTooltip,
  type DataTableColumns,
} from 'naive-ui'
import type { settings as cfg } from '../../wailsjs/go/models'
import ScanResultTable from '../components/ScanResultTable.vue'
import PowerStateTag from '../components/PowerStateTag.vue'
import * as api from '../api/app'
import { getMessage } from '../plugins/discrete'
import { useAppStore } from '../stores/app'
import { useScanStore } from '../stores/scan'

const app = useAppStore()
const scan = useScanStore()

const identifying = ref<string | null>(null)
const powering = ref<{ addr: string; on: boolean } | null>(null)
const removing = ref<string | null>(null)
const applying = ref(false)

const enableDraft = reactive<Record<string, boolean>>({})

function draftEnable(d: cfg.BaseStation): boolean {
  return enableDraft[d.addr] ?? d.enable
}

const enableChanges = computed(() => app.devices.filter((d) => draftEnable(d) !== d.enable))
const dirty = computed(() => enableChanges.value.length > 0)

function resetDraft() {
  for (const k of Object.keys(enableDraft)) delete enableDraft[k]
}

watch(
  () => app.devices,
  (next) => {
    const addrs = new Set(next.map((d) => d.addr))
    for (const k of Object.keys(enableDraft)) if (!addrs.has(k)) delete enableDraft[k]
  },
)

onMounted(() => {
  scan.bindEvents()
  void scan.sync()
  if (app.enabledDevices.length) void app.refreshPowerState()
})

onBeforeUnmount(() => {
  if (scan.scanning) void scan.stop()
})

async function confirm() {
  try {
    await scan.confirm(app.devices)
    getMessage().success('デバイスを登録しました')
  } catch {
    /* ストア側でトースト済み */
  }
}

async function apply() {
  applying.value = true
  try {
    await app.setDevices(app.devices.map((d) => ({ ...d, enable: draftEnable(d) }) as cfg.BaseStation))
    resetDraft()
    getMessage().success('適用しました')
  } catch {
    /* ストア側でトースト済み */
  } finally {
    applying.value = false
  }
}

async function setPower(addr: string, on: boolean) {
  powering.value = { addr, on }
  try {
    await app.setPowerSome(on, [addr])
    getMessage().success(`${addr} を${on ? 'ON' : 'OFF'}にしました`)
    window.setTimeout(() => void app.refreshPowerState(), 1500)
  } catch {
    /* ストア側でトースト済み */
  } finally {
    powering.value = null
  }
}

async function identify(addr: string) {
  identifying.value = addr
  try {
    await api.Identify(addr)
    getMessage().success(`${addr} に識別信号を送りました`)
  } catch (e) {
    getMessage().error(api.toError(e))
  } finally {
    identifying.value = null
  }
}

async function remove(addr: string) {
  removing.value = addr
  try {
    await app.setDevices(app.devices.filter((d) => d.addr !== addr))
  } catch {
    /* ストア側でトースト済み */
  } finally {
    removing.value = null
  }
}

const columns = computed<DataTableColumns<cfg.BaseStation>>(() => [
  {
    title: 'デバイス',
    key: 'device',
    render: (row) =>
      h(NFlex, { vertical: true, size: 0 }, () => [
        h(NText, null, () => row.name || '(名前なし)'),
        h(NText, { depth: 3, style: 'font-size: 12px' }, () => row.addr),
      ]),
  },
  {
    title: '制御対象',
    key: 'enable',
    width: 110,
    render: (row) =>
      h(NSwitch, {
        size: 'small',
        value: draftEnable(row),
        disabled: applying.value,
        'onUpdate:value': (v: boolean) => {
          enableDraft[row.addr] = v
        },
      }),
  },
  {
    title: '状態',
    key: 'state',
    width: 110,
    render: (row) =>
      row.enable
        ? h(PowerStateTag, { state: app.powerStates[row.addr] })
        : h(NText, { depth: 3, style: 'font-size: 12px' }, () => '制御対象外'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render: (row) =>
      h(NFlex, { size: 8, wrap: false }, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            secondary: true,
            disabled: !row.enable || powering.value !== null,
            loading: powering.value?.addr === row.addr && powering.value.on,
            onClick: () => setPower(row.addr, true),
          },
          () => 'ON',
        ),
        h(
          NButton,
          {
            size: 'small',
            secondary: true,
            disabled: !row.enable || powering.value !== null,
            loading: powering.value?.addr === row.addr && !powering.value.on,
            onClick: () => setPower(row.addr, false),
          },
          () => 'OFF',
        ),
        h(
          NButton,
          { size: 'small', disabled: !row.enable, loading: identifying.value === row.addr, onClick: () => identify(row.addr) },
          () => '識別',
        ),
        h(
          NTooltip,
          { disabled: app.devices.length > 1 },
          {
            trigger: () =>
              h('span', null, [
                h(
                  NPopconfirm,
                  { positiveText: '削除', negativeText: 'キャンセル', onPositiveClick: () => remove(row.addr) },
                  {
                    trigger: () =>
                      h(
                        NButton,
                        {
                          size: 'small',
                          type: 'error',
                          ghost: true,
                          disabled: app.devices.length <= 1,
                          loading: removing.value === row.addr,
                        },
                        () => '削除',
                      ),
                    default: () => `${row.addr} を登録から削除しますか？`,
                  },
                ),
              ]),
            default: () => '最低 1 台のデバイスが必要です',
          },
        ),
      ]),
  },
])
</script>

<template>
  <NFlex vertical :size="16">
    <NCard title="スキャン">
      <template #header-extra>
        <NFlex :size="8">
          <NButton type="primary" :disabled="scan.scanning" @click="scan.start()">スキャン開始</NButton>
          <NButton :disabled="!scan.scanning" @click="scan.stop()">停止</NButton>
        </NFlex>
      </template>
      <NFlex vertical :size="12">
        <NAlert v-if="scan.scanning" type="info" :show-icon="false">
          <NFlex align="center" :size="8" :wrap="false">
            <NSpin :size="16" />
            <span>検出中… 名前順に一覧へ追加されます。登録するデバイスにチェックを入れて確定してください。</span>
          </NFlex>
        </NAlert>
        <ScanResultTable v-model:checked-keys="scan.checked" :rows="scan.results" :registered="app.deviceAddrs" />
        <NFlex justify="end">
          <NButton type="primary" :disabled="!scan.checked.length" :loading="scan.confirming" @click="confirm">
            確定して登録
          </NButton>
        </NFlex>
      </NFlex>
    </NCard>

    <NCard title="登録済みデバイス">
      <template #header-extra>
        <NFlex :size="8">
          <NButton
            size="small"
            :loading="app.powerStatesLoading"
            :disabled="!app.enabledDevices.length || applying"
            @click="app.refreshPowerState()"
          >
            状態を更新
          </NButton>
          <NButton size="small" :disabled="!dirty || applying" @click="resetDraft">元に戻す</NButton>
          <NButton size="small" type="primary" :disabled="!dirty" :loading="applying" @click="apply">適用</NButton>
        </NFlex>
      </template>
      <NEmpty v-if="!app.devices.length" description="登録されているデバイスはありません" />
      <NDataTable v-else size="small" :columns="columns" :data="app.devices" :row-key="(row: cfg.BaseStation) => row.addr" :bordered="false" />
    </NCard>
  </NFlex>
</template>
