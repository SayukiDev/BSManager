<script setup lang="ts">
import { computed, h } from 'vue'
import { NDataTable, NEmpty, NTag, type DataTableColumns, type DataTableRowKey } from 'naive-ui'
import type { service } from '../../wailsjs/go/models'

type Row = service.BaseStation

const props = withDefaults(
  defineProps<{
    rows: Row[]
    /** 登録済みアドレス。選択不可にして「登録済み」タグを付ける */
    registered?: string[]
    hideState?: boolean
  }>(),
  { registered: () => [], hideState: false },
)

const checked = defineModel<string[]>('checkedKeys', { default: () => [] })

const rowKey = (row: Row) => row.addr

const columns = computed<DataTableColumns<Row>>(() => {
  const cols: DataTableColumns<Row> = [
    { type: 'selection', disabled: (row) => props.registered.includes(row.addr) },
    { title: '名前', key: 'name' },
    { title: 'アドレス', key: 'addr' },
    { title: 'RSSI', key: 'rssi', width: 100, render: (row) => `${row.rssi} dBm` },
  ]
  if (!props.hideState) {
    cols.push({
      title: '状態',
      key: 'state',
      width: 100,
      render: (row) =>
        props.registered.includes(row.addr)
          ? h(NTag, { size: 'small', type: 'success', bordered: false }, { default: () => '登録済み' })
          : null,
    })
  }
  return cols
})

function onUpdateChecked(keys: DataTableRowKey[]) {
  checked.value = keys.map(String)
}
</script>

<template>
  <NDataTable
    size="small"
    :columns="columns"
    :data="rows"
    :row-key="rowKey"
    :checked-row-keys="checked"
    :bordered="false"
    @update:checked-row-keys="onUpdateChecked"
  >
    <template #empty>
      <NEmpty description="まだ見つかっていません" />
    </template>
  </NDataTable>
</template>
