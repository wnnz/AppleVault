<template>
  <div class="app-table" :class="{ 'app-table--loading': loading }">
    <div class="app-table__frame" :class="{ 'app-table__frame--scroll-x': !!scrollX }">
      <div class="app-table__viewport" :style="{ overflowX: scrollX ? 'auto' : undefined }">
        <table :style="{ minWidth: scrollX ? `${scrollX}px` : undefined }">
          <thead>
            <tr class="app-table__row">
              <th
                v-for="header in table.getHeaderGroups()[0]?.headers ?? []"
                :key="header.id"
                class="app-table__header"
                :style="columnStyle(header.column.columnDef.meta as AppTableColumn<Row> | undefined)"
              >
                <FlexRender v-if="!header.isPlaceholder" :render="header.column.columnDef.header" :props="header.getContext()" />
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in table.getRowModel().rows" :key="row.id" class="app-table__row">
              <td
                v-for="cell in row.getVisibleCells()"
                :key="cell.id"
                class="app-table__cell"
                :style="columnStyle(cell.column.columnDef.meta as AppTableColumn<Row> | undefined)"
              >
                <div :class="{ 'app-table__ellipsis': (cell.column.columnDef.meta as AppTableColumn<Row> | undefined)?.ellipsis }">
                  <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
                </div>
              </td>
            </tr>
            <tr v-if="!loading && table.getRowModel().rows.length === 0">
              <td class="app-table__empty app-table__cell" :colspan="columns.length"><slot name="empty">暂无数据</slot></td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="loading" class="app-table__loading">加载中…</div>
    </div>

    <div v-if="pagination && table.getPageCount() > 1" class="app-table__pagination">
      <button class="app-table__page" :disabled="!table.getCanPreviousPage()" @click="table.previousPage()">‹</button>
      <span>{{ table.getState().pagination.pageIndex + 1 }} / {{ table.getPageCount() }}</span>
      <button class="app-table__page" :disabled="!table.getCanNextPage()" @click="table.nextPage()">›</button>
    </div>
  </div>
</template>

<script setup lang="ts" generic="Row extends Record<string, any>">
import { computed } from 'vue'
import {
  FlexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useVueTable,
  type ColumnDef
} from '@tanstack/vue-table'

export interface AppTableColumn<Row> {
  title: string
  key: keyof Row | string
  width?: number
  minWidth?: number
  fixed?: 'left' | 'right'
  ellipsis?: boolean | { tooltip?: boolean }
  render?: (row: Row) => unknown
}

const props = withDefaults(defineProps<{
  columns: AppTableColumn<Row>[]
  data: Row[]
  loading?: boolean
  pagination?: { pageSize: number } | false
  scrollX?: number
  flexHeight?: boolean
  size?: 'small' | 'medium'
}>(), {
  loading: false,
  pagination: false,
  flexHeight: false,
  size: 'medium'
})

const tanstackColumns = computed<ColumnDef<Row>[]>(() => props.columns.map((column, index) => ({
  id: String(column.key ?? index),
  accessorFn: row => row[column.key as keyof Row],
  header: () => column.title,
  cell: info => column.render ? column.render(info.row.original) : String(info.getValue() ?? '-'),
  meta: column
})))

const table = useVueTable({
  get data() { return props.data },
  get columns() { return tanstackColumns.value },
  getCoreRowModel: getCoreRowModel(),
  getPaginationRowModel: getPaginationRowModel(),
  initialState: {
    pagination: { pageIndex: 0, pageSize: props.pagination ? props.pagination.pageSize : Number.MAX_SAFE_INTEGER }
  }
})

function columnStyle(column?: AppTableColumn<Row>) {
  if (!column) return undefined
  const fixedStyle = column.fixed === 'right'
    ? { position: 'sticky' as const, right: '0', zIndex: 2 }
    : column.fixed === 'left'
      ? { position: 'sticky' as const, left: '0', zIndex: 2 }
      : {}
  return {
    ...fixedStyle,
    width: column.width ? `${column.width}px` : undefined,
    minWidth: column.minWidth ? `${column.minWidth}px` : undefined
  }
}
</script>

<style>
.app-table { position: relative; display: flex; flex-direction: column; width: 100%; height: 100%; color: var(--ui-text); }
.app-table__frame { position: relative; flex: 1; min-height: 0; }
.app-table__viewport { position: relative; width: 100%; height: 100%; min-height: 0; overflow: auto; }
.app-table table { width: 100%; border-spacing: 0; border-collapse: collapse; table-layout: fixed; }
.app-table__header, .app-table__cell { height: 40px; padding: 0 12px; border-bottom: 1px solid var(--ui-border); text-align: left; box-sizing: border-box; }
.app-table__header { height: 38px; background: var(--ui-table-header); font-size: 13px; font-weight: 600; }
.app-table__cell { background: var(--ui-surface); font-size: 13px; }
.app-table tbody tr:hover .app-table__cell { background: var(--ui-table-hover); }
.app-table__ellipsis { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.app-table__empty { border-bottom: 0; text-align: center; color: var(--ui-muted); }
.app-table__loading { position: absolute; inset: 38px 0 0; display: grid; place-items: center; background: color-mix(in srgb, var(--ui-surface) 78%, transparent); color: var(--ui-muted); }
.app-table__pagination { display: flex; align-items: center; justify-content: flex-end; gap: 8px; min-height: 42px; padding: 4px 10px; color: var(--ui-muted); font-size: 12px; }
.app-table__page { width: 30px; height: 30px; border: 1px solid var(--ui-border); border-radius: 4px; background: var(--ui-control-bg); color: var(--ui-text); cursor: pointer; }
.app-table__page:disabled { opacity: .4; cursor: not-allowed; }
.app-button-group { display: inline-flex; align-items: center; gap: 6px; }
</style>
