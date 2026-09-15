<script setup lang="ts">
/**
 * AI 摘要弹窗：生成中态 → 结果可编辑 → 插入文首（自动补 <!--more-->）/ 复制 / 重新生成。
 */
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconLoader2, IconCopy, IconCheck } from '@tabler/icons-vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'

const props = defineProps<{ open: boolean; loading: boolean; text: string }>()
const emit = defineEmits<{
  'update:open': [v: boolean]
  regenerate: []
  insert: [text: string]
}>()

const { t } = useI18n()

const draft = ref('')
watch(
  () => props.text,
  (v) => {
    draft.value = v || ''
  },
)
watch(
  () => props.open,
  (o) => {
    if (o) draft.value = props.text || ''
  },
)

const copied = ref(false)
async function copy() {
  if (!draft.value.trim()) return
  try {
    await navigator.clipboard.writeText(draft.value.trim())
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    /* ignore */
  }
}

function insert() {
  if (!draft.value.trim()) return
  emit('insert', draft.value.trim())
  emit('update:open', false)
}
</script>

<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ t('editor.summaryDialog.title') }}</DialogTitle>
      </DialogHeader>

      <div v-if="loading" class="flex items-center justify-center gap-2 py-10 text-sm text-muted-foreground">
        <IconLoader2 class="size-4 animate-spin" />
        {{ t('editor.summaryDialog.generating') }}
      </div>
      <Textarea
        v-else
        v-model="draft"
        rows="5"
        class="min-h-28"
        :placeholder="t('editor.summaryDialog.placeholder')"
      />

      <DialogFooter class="gap-2">
        <Button variant="outline" :disabled="loading" @click="emit('regenerate')">
          {{ t('editor.summaryDialog.regenerate') }}
        </Button>
        <Button variant="outline" :disabled="loading || !draft.trim()" @click="copy">
          <IconCheck v-if="copied" class="mr-1 size-3.5" />
          <IconCopy v-else class="mr-1 size-3.5" />
          {{ copied ? t('editor.summaryDialog.copied') : t('editor.summaryDialog.copy') }}
        </Button>
        <Button :disabled="loading || !draft.trim()" @click="insert">
          {{ t('editor.summaryDialog.insertTop') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
