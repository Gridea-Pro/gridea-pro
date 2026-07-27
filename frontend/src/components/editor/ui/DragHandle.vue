<template>
  <DragHandle v-if="editor" :editor="editor" :on-node-change="handleNodeChange">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          class="flex h-6 w-5 cursor-grab items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-foreground active:cursor-grabbing"
          :title="t('editor.dragHandle.title')"
          @mousedown.stop
        >
          <GripVertical class="size-4" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="start" side="bottom" class="min-w-44">
        <DropdownMenuItem :disabled="!hasBlock" @select="duplicateBlock">
          <Copy class="mr-2 size-4" />
          <span>{{ t('editor.dragHandle.duplicate') }}</span>
        </DropdownMenuItem>

        <DropdownMenuItem :disabled="!hasBlock" @select="clearFormatting">
          <Eraser class="mr-2 size-4" />
          <span>{{ t('editor.dragHandle.clearFormat') }}</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          :disabled="!hasBlock"
          class="text-destructive focus:text-destructive"
          @select="deleteBlock"
        >
          <Trash2 class="mr-2 size-4" />
          <span>{{ t('editor.dragHandle.delete') }}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </DragHandle>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import type { Node as ProseMirrorNode } from '@tiptap/pm/model'
import { DragHandle } from '@tiptap/extension-drag-handle-vue-3'
import { Copy, Eraser, GripVertical, Trash2 } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'

const props = defineProps<{ editor: Editor | null }>()

const { t } = useI18n()

interface CurrentBlock {
  node: ProseMirrorNode | null
  pos: number
}

const current = ref<CurrentBlock>({ node: null, pos: -1 })

const hasBlock = computed(() => current.value.node !== null && current.value.pos >= 0)

function handleNodeChange(data: { node: ProseMirrorNode | null; pos: number; editor?: unknown }): void {
  current.value = { node: data.node, pos: data.pos }
}

function getRange(): { from: number; to: number } | null {
  const { node, pos } = current.value
  if (!node || pos < 0) return null
  return { from: pos, to: pos + node.nodeSize }
}

function duplicateBlock(): void {
  const editor = props.editor
  const { node, pos } = current.value
  if (!editor || !node || pos < 0) return
  const insertAt = pos + node.nodeSize
  editor
    .chain()
    .focus()
    .insertContentAt(insertAt, node.toJSON())
    .run()
}

function deleteBlock(): void {
  const editor = props.editor
  const range = getRange()
  if (!editor || !range) return
  editor.chain().focus().deleteRange(range).run()
  current.value = { node: null, pos: -1 }
}

function clearFormatting(): void {
  const editor = props.editor
  const range = getRange()
  if (!editor || !range) return
  editor
    .chain()
    .focus()
    .setTextSelection(range)
    .unsetAllMarks()
    .clearNodes()
    .run()
}
</script>
