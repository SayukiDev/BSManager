import { defineStore } from 'pinia'
import { ref } from 'vue'
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime'
import type { service as svc, settings as cfg } from '../../wailsjs/go/models'
import * as api from '../api/app'
import { getMessage } from '../plugins/discrete'
import { useAppStore } from './app'

const EVENT_UPDATE = 'scan:update'
const EVENT_STOPPED = 'scan:stopped'

export const useScanStore = defineStore('scan', () => {
  const scanning = ref(false)
  const results = ref<svc.BaseStation[]>([])
  const checked = ref<string[]>([])
  const confirming = ref(false)

  function merge(list: svc.BaseStation[]) {
    results.value = [...list].sort((a, b) => a.name.localeCompare(b.name) || a.addr.localeCompare(b.addr))
  }

  let fetching = false
  let dirty = false
  async function refresh() {
    if (fetching) {
      dirty = true
      return
    }
    fetching = true
    try {
      do {
        dirty = false
        merge((await api.GetScanedBaseStations()) ?? [])
      } while (dirty)
    } catch (e) {
      console.error('scan refresh failed', e)
    } finally {
      fetching = false
    }
  }

  let bound = false
  function bindEvents() {
    if (bound) return
    bound = true
    EventsOn(EVENT_UPDATE, () => {
      void refresh()
    })
    EventsOn(EVENT_STOPPED, (msg: string) => {
      scanning.value = false
      if (msg) getMessage().error(`スキャンエラー: ${msg}`)
    })
  }
  function unbindEvents() {
    if (!bound) return
    EventsOff(EVENT_UPDATE, EVENT_STOPPED)
    bound = false
  }

  async function sync() {
    try {
      scanning.value = await api.IsScanning()
      await refresh()
    } catch (e) {
      console.error('scan sync failed', e)
    }
  }

  async function start() {
    try {
      await api.StartScan()
      results.value = []
      scanning.value = true
    } catch (e) {
      getMessage().error(api.toError(e))
    }
  }

  async function stop() {
    try {
      await api.StopScan()
    } catch (e) {
      getMessage().error(api.toError(e))
    }
  }

  async function confirm(extra: cfg.BaseStation[] = []) {
    if (!checked.value.length) return
    confirming.value = true
    try {
      if (scanning.value) await stop()
      const devices = new Map(extra.map((d) => [d.addr, d]))
      for (const addr of checked.value) {
        if (devices.has(addr)) continue
        const found = results.value.find((r) => r.addr === addr)
        devices.set(addr, { enable: true, name: found?.name ?? '', addr } as cfg.BaseStation)
      }
      await useAppStore().setDevices(Array.from(devices.values()))
      checked.value = []
    } finally {
      confirming.value = false
    }
  }

  return {
    scanning,
    results,
    checked,
    confirming,
    bindEvents,
    unbindEvents,
    sync,
    start,
    stop,
    confirm,
  }
})
