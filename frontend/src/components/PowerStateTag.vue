<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'

const props = defineProps<{
  /** GetPowerState が返す文字列: ON / STANDBY / SLEEP / Booting / UNKNOWN。未取得なら undefined */
  state?: string
}>()

type TagType = 'default' | 'success' | 'warning' | 'info' | 'error'

const view = computed<{ label: string; type: TagType }>(() => {
  switch ((props.state ?? '').toUpperCase()) {
    case 'ON':
      return { label: '稼働中', type: 'success' }
    case 'BOOTING':
      return { label: '起動中', type: 'info' }
    case 'STANDBY':
      return { label: 'スタンバイ', type: 'warning' }
    case 'SLEEP':
      return { label: 'スリープ', type: 'default' }
    case '':
      return { label: '未取得', type: 'default' }
    default:
      return { label: props.state ?? '不明', type: 'error' }
  }
})
</script>

<template>
  <NTag size="small" :type="view.type" :bordered="false">{{ view.label }}</NTag>
</template>
