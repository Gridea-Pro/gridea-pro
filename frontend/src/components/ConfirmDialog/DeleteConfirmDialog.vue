<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle class="text-lg font-bold text-destructive">{{ title || t('common.warning') }}</DialogTitle>
      </DialogHeader>
      <div class="flex items-center gap-3 text-destructive">
        <ExclamationTriangleIcon class="size-5" />
        <p class="text-sm text-foreground">{{ content || t('common.deleteWarning') }}</p>
        <!-- // TODO: Check i18n key -->
      </div>
      <DialogFooter>
        <Button
variant="outline" class="border border-primary/10 hover:bg-primary/10 cursor-pointer"
          @click="handleCancel">{{ cancelText || t('common.cancel') }}</Button>
        <Button variant="destructive" class="hover:bg-destructive cursor-pointer" @click="handleConfirm">{{ confirmText
          || t('common.delete') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()

defineProps<{
  open: boolean
  title?: string
  content?: string
  confirmText?: string
  cancelText?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const handleCancel = () => {
  emit('update:open', false)
  emit('cancel')
}

const handleConfirm = () => {
  emit('confirm')
  // 注意：确认后是否自动关闭弹窗由父组件控制，或者在这里也可以关闭
  // 通常父组件处理完逻辑后会关闭，或者这里直接关闭
  // 为了灵活，这里不强制关闭，但通常确认就是为了关闭并执行操作
}
</script>
