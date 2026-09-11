import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime'
import type { service as svc, settings as cfg } from '../../wailsjs/go/models'
import * as api from '../api/app'
import { getMessage } from '../plugins/discrete'

const EVENT_POWER_CHANGING = 'power:changing'
const EVENT_POWER = 'power:changed'
/** power:changing の後に power:changed が来ない（BLE 反映失敗など）場合に切替中表示を解除するまでの時間 */
const POWER_CHANGING_TIMEOUT_MS = 30_000
const EVENT_SETTINGS = 'settings:changed'
const EVENT_ERROR = 'app:error'
const EVENT_READY = 'app:ready'

export const useAppStore = defineStore('app', () => {
  const status = ref<svc.Status | null>(null)
  const settings = ref<cfg.Content | null>(null)
  const powerStates = ref<Record<string, string>>({})
  const powerStatesLoading = ref(false)
  /** マネージャーが実機へ電源状態を反映している最中か（power:changing 〜 power:changed の間） */
  const powerChanging = ref(false)
  const loaded = ref(false)

  const devices = computed<cfg.BaseStation[]>(() => settings.value?.devices ?? [])
  const deviceAddrs = computed(() => devices.value.map((d) => d.addr))
  const enabledDevices = computed(() => devices.value.filter((d) => d.enable))
  const followSteamVR = computed(() => settings.value?.followSteamVR ?? true)
  const isSetuped = computed(() => settings.value?.setuped ?? false)

  /** GetPowerState の文字列を ON/OFF に正規化する。判定不能なら null */
  function deviceIsOn(state: string | undefined): boolean | null {
    switch ((state ?? '').toUpperCase()) {
      case 'ON':
      case 'BOOTING':
        return true
      case 'STANDBY':
      case 'SLEEP':
        return false
      default:
        return null
    }
  }

  /** 実機の作動状態がマネージャーの状態と食い違っているデバイス一覧 */
  const powerMismatchDevices = computed(() => {
    const managerOn = status.value?.powerOn
    // 反映中は実機側が遷移途中なので差異として扱わない
    if (managerOn === undefined || powerChanging.value) return []
    return enabledDevices.value.filter((d) => {
      const on = deviceIsOn(powerStates.value[d.addr])
      return on !== null && on !== managerOn
    })
  })
  const hasPowerMismatch = computed(() => powerMismatchDevices.value.length > 0)

  let statusInFlight = false
  async function refreshStatus() {
    if (statusInFlight) return
    statusInFlight = true
    try {
      status.value = await api.GetStatus()
    } catch (e) {
      console.error('GetStatus failed', e)
    } finally {
      statusInFlight = false
    }
  }

  async function refreshSettings() {
    try {
      settings.value = await api.GetSettings()
    } catch (e) {
      console.error('GetSettings failed', e)
    }
  }

  async function refreshPowerState() {
    if (!enabledDevices.value.length) {
      powerStates.value = {}
      return
    }
    powerStatesLoading.value = true
    try {
      powerStates.value = (await api.GetPowerState()) ?? {}
    } catch (e) {
      getMessage().error(api.toError(e))
    } finally {
      powerStatesLoading.value = false
    }
  }

  async function refreshAll() {
    await Promise.all([refreshSettings(), refreshStatus()])
    loaded.value = true
  }

  let timer: number | null = null
  function startPolling(ms = 3000) {
    stopPolling()
    timer = window.setInterval(() => {
      if (!document.hidden) void refreshStatus()
    }, ms)
  }
  function stopPolling() {
    if (timer !== null) {
      window.clearInterval(timer)
      timer = null
    }
  }

  let changingTimer: number | null = null
  function beginPowerChanging(on: boolean) {
    powerChanging.value = true
    if (status.value) status.value.powerOn = on
    if (changingTimer !== null) window.clearTimeout(changingTimer)
    changingTimer = window.setTimeout(() => {
      changingTimer = null
      if (!powerChanging.value) return
      powerChanging.value = false
      void refreshStatus()
      void refreshPowerState()
    }, POWER_CHANGING_TIMEOUT_MS)
  }
  function endPowerChanging() {
    if (changingTimer !== null) {
      window.clearTimeout(changingTimer)
      changingTimer = null
    }
    powerChanging.value = false
  }

  let bound = false
  function bindEvents() {
    if (bound) return
    bound = true
    EventsOn(EVENT_POWER_CHANGING, (on: boolean) => {
      beginPowerChanging(on)
    })
    EventsOn(EVENT_POWER, (on: boolean) => {
      endPowerChanging()
      if (status.value) status.value.powerOn = on
      window.setTimeout(() => void refreshPowerState(), 1500)
    })
    EventsOn(EVENT_SETTINGS, (c: cfg.Content) => {
      settings.value = c
      void refreshStatus()
    })
    EventsOn(EVENT_ERROR, (msg: string) => {
      getMessage().error(msg)
    })
    EventsOn(EVENT_READY, () => {
      void refreshAll()
    })
  }
  function unbindEvents() {
    if (!bound) return
    EventsOff(EVENT_POWER_CHANGING, EVENT_POWER, EVENT_SETTINGS, EVENT_ERROR, EVENT_READY)
    endPowerChanging()
    bound = false
  }

  /** 状態を変えるバインディング呼び出しの共通処理: エラーはトーストして再 throw、終了後に状態を更新 */
  async function mutate<T>(fn: () => Promise<T>): Promise<T> {
    try {
      return await fn()
    } catch (e) {
      getMessage().error(api.toError(e))
      throw e
    } finally {
      void refreshStatus()
    }
  }

  async function setPower(on: boolean) {
    await mutate(() => api.SetPower(on))
  }

  /** マネージャーの状態を実機に再適用して作動状態を揃える */
  async function syncPowerState() {
    const on = status.value?.powerOn
    if (on === undefined) return
    await setPower(on)
    await refreshPowerState()
  }

  async function setDevices(list: cfg.BaseStation[]) {
    await mutate(() => api.SetDevices(list))
    await refreshSettings()
  }

  async function setPowerSome(on: boolean, addrs: string[]) {
    if (!addrs.length) return
    await mutate(() => api.SetPowerSome(on, addrs))
  }

  async function saveSettings(c: cfg.Content) {
    await mutate(() => api.SaveSettings(c))
    await refreshSettings()
  }

  return {
    status,
    settings,
    powerStates,
    powerStatesLoading,
    powerChanging,
    loaded,
    devices,
    deviceAddrs,
    enabledDevices,
    followSteamVR,
    isSetuped,
    powerMismatchDevices,
    hasPowerMismatch,
    refreshStatus,
    refreshSettings,
    refreshPowerState,
    refreshAll,
    startPolling,
    stopPolling,
    bindEvents,
    unbindEvents,
    setPower,
    syncPowerState,
    setDevices,
    setPowerSome,
    saveSettings,
    deviceIsOn,
  }
})
