<script setup lang="ts">
import { onBeforeUnmount, ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Editor } from '@tiptap/vue-3'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import {
  IconTable as TableIcon,
  IconRowInsertTop as ArrowUpToLine,
  IconRowInsertBottom as ArrowDownToLine,
  IconColumnInsertLeft as ArrowLeftToLine,
  IconColumnInsertRight as ArrowRightToLine,
  IconRowRemove as Rows3,
  IconColumnRemove as Columns3,
  IconArrowsJoin2 as Combine,
  IconArrowsSplit2 as Split,
  IconHeading as Heading,
  IconTrash as Trash2,
} from '@tabler/icons-vue'

const props = defineProps<{ editor: Editor | null }>()

const { t } = useI18n()

// Force re-evaluation of isActive/can on selection or doc changes.
const tick = ref(0)
function bump() {
  tick.value++
}

// editor 在 onMounted 时可能仍为 undefined（useEditor 异步就绪），用 watch 待其出现再绑定
watch(
  () => props.editor,
  (ed, prev) => {
    prev?.off('transaction', bump)
    prev?.off('selectionUpdate', bump)
    ed?.on('transaction', bump)
    ed?.on('selectionUpdate', bump)
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  props.editor?.off('transaction', bump)
  props.editor?.off('selectionUpdate', bump)
})

const inTable = computed(() => {
  // eslint-disable-next-line @typescript-eslint/no-unused-expressions
  tick.value
  const e = props.editor
  return !!e && e.isActive('table')
})

// 菜单锚定到光标所在表格的右上角（相对 .rich-pane 绝对定位；原来固定飘在编辑器右上角，
// 远离表格基本不可发现）。rich-pane 在双栏模式自滚动，需叠加 scrollTop。
const menuStyle = ref<Record<string, string>>({ top: '8px', right: '8px' })

function locate() {
  const e = props.editor
  if (!e || !e.isActive('table')) return
  try {
    const { from } = e.state.selection
    const dom = e.view.domAtPos(from).node
    const el = (dom instanceof HTMLElement ? dom : (dom as Node).parentElement) as HTMLElement | null
    const tableEl = el?.closest('table')
    const pane = el?.closest('.rich-pane') as HTMLElement | null
    if (!tableEl || !pane) return
    const tr = tableEl.getBoundingClientRect()
    const pr = pane.getBoundingClientRect()
    menuStyle.value = {
      top: `${Math.max(0, tr.top - pr.top + pane.scrollTop + 4)}px`,
      // 夹在分栏可视宽度内，窄分栏/宽表格时不溢出
      left: `${Math.max(0, Math.min(tr.right - pr.left - 32, pane.clientWidth - 40))}px`,
      right: 'auto',
    }
  } catch {
    /* domAtPos 在过渡态可能抛错，忽略本次定位 */
  }
}

watch(tick, () => {
  void nextTick(locate)
})

type ChainKey =
  | 'addRowBefore'
  | 'addRowAfter'
  | 'addColumnBefore'
  | 'addColumnAfter'
  | 'deleteRow'
  | 'deleteColumn'
  | 'mergeCells'
  | 'splitCell'
  | 'toggleHeaderRow'
  | 'deleteTable'

function run(name: ChainKey) {
  const e = props.editor
  if (!e) return
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(e.chain().focus() as any)[name]().run()
}

function can(name: ChainKey): boolean {
  // eslint-disable-next-line @typescript-eslint/no-unused-expressions
  tick.value
  const e = props.editor
  if (!e) return false
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const chain = e.can().chain().focus() as any
  if (typeof chain[name] !== 'function') return false
  try {
    return !!chain[name]().run()
  } catch {
    return false
  }
}
</script>

<template>
  <div v-if="inTable" class="table-menu" :style="menuStyle">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border border-border bg-popover text-muted-foreground shadow-sm transition-colors ed-ctl"
          :title="t('editor.tableMenu.menu')"
          aria-haspopup="menu"
          @mousedown.prevent
        >
          <TableIcon class="size-4" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" :side-offset="6" class="min-w-44">
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('addRowBefore')" @select="run('addRowBefore')">
          <ArrowUpToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addRowBefore') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('addRowAfter')" @select="run('addRowAfter')">
          <ArrowDownToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addRowAfter') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('addColumnBefore')" @select="run('addColumnBefore')">
          <ArrowLeftToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addColumnBefore') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('addColumnAfter')" @select="run('addColumnAfter')">
          <ArrowRightToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addColumnAfter') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem class="ed-menu-item" :disabled="!can('deleteRow')" @select="run('deleteRow')">
          <Rows3 class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.deleteRow') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('deleteColumn')" @select="run('deleteColumn')">
          <Columns3 class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.deleteColumn') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem class="ed-menu-item" :disabled="!can('mergeCells')" @select="run('mergeCells')">
          <Combine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.mergeCells') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('splitCell')" @select="run('splitCell')">
          <Split class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.splitCell') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem class="ed-menu-item" :disabled="!can('toggleHeaderRow')" @select="run('toggleHeaderRow')">
          <Heading class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.toggleHeaderRow') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          class="ed-menu-item text-destructive focus:text-destructive"
          :disabled="!can('deleteTable')"
          @select="run('deleteTable')"
        >
          <Trash2 class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.deleteTable') }}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</template>

<style scoped>
.table-menu {
  position: absolute;
  z-index: 6;
}
</style>
