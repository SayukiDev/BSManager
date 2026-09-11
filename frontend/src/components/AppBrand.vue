<script setup lang="ts">
import { NFlex, NText } from 'naive-ui'
import appicon from '../assets/appicon.png'

/** アプリアイコン + 名前。collapsed のときはアイコンだけ表示する */
defineProps<{ collapsed?: boolean }>()
</script>

<template>
  <NFlex align="center" :size="8" :wrap="false" style="overflow: hidden; white-space: nowrap">
    <img :src="appicon" width="22" height="22" alt="" style="flex-shrink: 0" />
    <NText v-show="!collapsed" strong>BSManager</NText>
  </NFlex>
</template>
