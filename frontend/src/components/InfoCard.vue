<script setup lang="ts">
import type { Component } from 'vue'
import { NCard, NFlex, NIcon, NText } from 'naive-ui'

/** 参考UIの「左アイコン + タイトル + 説明行 + 右側の値」カード */
defineProps<{
  title: string
  icon: Component
}>()
</script>

<template>
  <NCard>
    <NFlex align="center" :size="24" :wrap="false">
      <NIcon :size="28" :depth="3" :component="icon" style="flex-shrink: 0; margin: 0 12px" />
      <NFlex vertical :size="6" style="flex: 1; min-width: 0">
        <NText strong style="font-size: 16px">{{ title }}</NText>
        <slot />
      </NFlex>
      <NFlex v-if="$slots.aside" vertical align="end" :size="4" style="flex-shrink: 0; margin-right: 12px">
        <slot name="aside" />
      </NFlex>
    </NFlex>
  </NCard>
</template>
