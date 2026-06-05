<script setup lang="ts">
import { onBeforeUnmount, ref, computed, watch } from 'vue'
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
  <div v-if="inTable" class="table-menu">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border border-border bg-popover text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
          :title="t('editor.tableMenu.menu')"
          aria-haspopup="menu"
          @mousedown.prevent
        >
          <TableIcon class="size-4" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" :side-offset="6" class="min-w-44">
        <DropdownMenuItem :disabled="!can('addRowBefore')" @select="run('addRowBefore')">
          <ArrowUpToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addRowBefore') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('addRowAfter')" @select="run('addRowAfter')">
          <ArrowDownToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addRowAfter') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('addColumnBefore')" @select="run('addColumnBefore')">
          <ArrowLeftToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addColumnBefore') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('addColumnAfter')" @select="run('addColumnAfter')">
          <ArrowRightToLine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.addColumnAfter') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem :disabled="!can('deleteRow')" @select="run('deleteRow')">
          <Rows3 class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.deleteRow') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('deleteColumn')" @select="run('deleteColumn')">
          <Columns3 class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.deleteColumn') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem :disabled="!can('mergeCells')" @select="run('mergeCells')">
          <Combine class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.mergeCells') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('splitCell')" @select="run('splitCell')">
          <Split class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.splitCell') }}</span>
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="!can('toggleHeaderRow')" @select="run('toggleHeaderRow')">
          <Heading class="mr-2 size-4" />
          <span>{{ t('editor.tableMenu.toggleHeaderRow') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          class="text-destructive focus:text-destructive"
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
  top: 8px;
  right: 8px;
  z-index: 6;
}
</style>
