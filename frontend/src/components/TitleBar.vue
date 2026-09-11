<script setup lang="ts">
import { NButton, NFlex, NIcon } from 'naive-ui'
import { CloseOutline, CopyOutline, RemoveOutline, SquareOutline } from '@vicons/ionicons5'
import AppBrand from './AppBrand.vue'
import { useWindowControls } from '../composables/useWindow'

/** brand=false のときはアイコン・名前を出さない（サイドバー側に表示する場合） */
withDefaults(defineProps<{ brand?: boolean }>(), { brand: true })

const { isMaximised, minimise, toggleMaximise, hide } = useWindowControls()
</script>

<template>
  <NFlex class="titlebar" align="center" justify="space-between" :wrap="false" style="padding-left: 12px">
    <AppBrand v-if="brand" />
    <span v-else />
    <NFlex :size="0" :wrap="false" style="--wails-draggable: no-drag">
      <NButton quaternary style="width: 46px; height: 40px" title="最小化" @click="minimise">
        <template #icon><NIcon :component="RemoveOutline" /></template>
      </NButton>
      <NButton
        quaternary
        style="width: 46px; height: 40px"
        :title="isMaximised ? '元に戻す' : '最大化'"
        @click="toggleMaximise"
      >
        <template #icon>
          <NIcon :size="14" :component="isMaximised ? CopyOutline : SquareOutline" />
        </template>
      </NButton>
      <NButton
        quaternary
        type="error"
        style="width: 46px; height: 40px"
        title="閉じる（バックグラウンドで動作を継続。終了はタスクトレイから）"
        @click="hide"
      >
        <template #icon><NIcon :component="CloseOutline" /></template>
      </NButton>
    </NFlex>
  </NFlex>
</template>
