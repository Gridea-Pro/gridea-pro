<script setup lang="ts">
/**
 * 图片编辑弹窗（对齐 ctzhian 原版：编辑图片地址 + 描述）。
 */
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const props = defineProps<{ open: boolean; src: string; alt: string }>()
const emit = defineEmits<{
  'update:open': [v: boolean]
  save: [payload: { src: string; alt: string }]
}>()

const { t } = useI18n()

const srcInput = ref('')
const altInput = ref('')

watch(
  () => props.open,
  (o) => {
    if (o) {
      srcInput.value = props.src
      altInput.value = props.alt
    }
  },
)

function onOpenChange(v: boolean) {
  emit('update:open', v)
}
function save() {
  if (!srcInput.value.trim()) return
  emit('save', { src: srcInput.value.trim(), alt: altInput.value.trim() })
  emit('update:open', false)
}
</script>

<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t('editor.imageEdit.title') }}</DialogTitle>
      </DialogHeader>
      <div class="space-y-3">
        <div class="space-y-1.5">
          <label class="text-xs text-muted-foreground">{{ t('editor.imageEdit.src') }}</label>
          <Input v-model="srcInput" :placeholder="t('editor.imageEdit.srcPlaceholder')" @keydown.enter.prevent="save" />
        </div>
        <div class="space-y-1.5">
          <label class="text-xs text-muted-foreground">{{ t('editor.imageEdit.alt') }}</label>
          <Input v-model="altInput" :placeholder="t('editor.imageEdit.altPlaceholder')" @keydown.enter.prevent="save" />
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="onOpenChange(false)">{{ t('editor.imageEdit.cancel') }}</Button>
        <Button :disabled="!srcInput.trim()" @click="save">{{ t('editor.imageEdit.save') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
