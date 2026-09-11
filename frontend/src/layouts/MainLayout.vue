<script setup lang="ts">
import { h, type Component } from 'vue'
import { NH1, NIcon, NLayout, NLayoutSider, NMenu, type MenuOption } from 'naive-ui'
import { BluetoothOutline, DocumentTextOutline, HomeOutline, SettingsOutline } from '@vicons/ionicons5'
import AppBrand from '../components/AppBrand.vue'
import TitleBar from '../components/TitleBar.vue'
import HomePage from '../pages/HomePage.vue'
import BaseStationsPage from '../pages/BaseStationsPage.vue'
import LogsPage from '../pages/LogsPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'
import { useUiStore, type PageKey } from '../stores/ui'

const SIDER_WIDTH = 220
const SIDER_COLLAPSED_WIDTH = 56

const ui = useUiStore()

const renderIcon = (icon: Component) => () => h(NIcon, null, { default: () => h(icon) })

const topOptions: MenuOption[] = [
  { label: 'ホーム', key: 'home', icon: renderIcon(HomeOutline) },
  { label: 'ベースステーション', key: 'stations', icon: renderIcon(BluetoothOutline) },
  { label: 'ログ', key: 'logs', icon: renderIcon(DocumentTextOutline) },
]
const bottomOptions: MenuOption[] = [{ label: '設定', key: 'settings', icon: renderIcon(SettingsOutline) }]

const pages: Record<PageKey, Component> = {
  home: HomePage,
  stations: BaseStationsPage,
  logs: LogsPage,
  settings: SettingsPage,
}
const titles: Record<PageKey, string> = {
  home: 'ホーム',
  stations: 'ベースステーション',
  logs: 'ログ',
  settings: '設定',
}

function select(key: string) {
  ui.page = key as PageKey
}
</script>

<template>
  <NLayout has-sider style="height: 100vh">
    <NLayoutSider
      v-model:collapsed="ui.siderCollapsed"
      bordered
      collapse-mode="width"
      :width="SIDER_WIDTH"
      :collapsed-width="SIDER_COLLAPSED_WIDTH"
      show-trigger="arrow-circle"
    >
      <div style="display: flex; flex-direction: column; height: 100%">
        <!-- アイコン + アプリ名。この行もウィンドウのドラッグ領域にする。折りたたみ時はアイコンだけ -->
        <div class="titlebar" style="display: flex; align-items: center; padding-left: 17px">
          <AppBrand :collapsed="ui.siderCollapsed" />
        </div>
        <NMenu
          :options="topOptions"
          :value="ui.page"
          :collapsed="ui.siderCollapsed"
          :collapsed-width="SIDER_COLLAPSED_WIDTH"
          :collapsed-icon-size="22"
          @update:value="select"
        />
        <NMenu
          style="margin-top: auto"
          :options="bottomOptions"
          :value="ui.page"
          :collapsed="ui.siderCollapsed"
          :collapsed-width="SIDER_COLLAPSED_WIDTH"
          :collapsed-icon-size="22"
          @update:value="select"
        />
      </div>
    </NLayoutSider>
    <div style="display: flex; flex-direction: column; flex: 1; min-width: 0">
      <TitleBar :brand="false" />
      <NLayout :native-scrollbar="false" style="flex: 1; min-height: 0" content-style="padding: 24px 32px">
        <NH1>{{ titles[ui.page] }}</NH1>
        <component :is="pages[ui.page]" />
      </NLayout>
    </div>
  </NLayout>
</template>
