<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t('editor.linkDialog.title') }}</DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-3 py-1">
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-foreground" :for="urlId">
            {{ t('editor.linkDialog.url') }}
          </label>
          <Input
            :id="urlId"
            ref="urlInputRef"
            v-model="urlValue"
            type="url"
            :placeholder="t('editor.linkDialog.urlPlaceholder')"
            @keydown.enter.prevent="onConfirm"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-foreground" :for="textId">
            {{ t('editor.linkDialog.text') }}
          </label>
          <Input
            :id="textId"
            v-model="textValue"
            type="text"
            :placeholder="t('editor.linkDialog.textPlaceholder')"
          />
        </div>
      </div>

      <DialogFooter>
        <Button
          v-if="props.url"
          variant="destructive"
          class="sm:mr-auto"
          @click="onRemove"
        >
          {{ t('editor.linkDialog.remove') }}
        </Button>
        <Button variant="outline" @click="onCancel">
          {{ t('editor.linkDialog.cancel') }}
        </Button>
        <Button :disabled="!canConfirm" @click="onConfirm">
          {{ t('editor.linkDialog.confirm') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  url?: string
  text?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [payload: { url: string; text: string }]
  remove: []
}>()

const { t } = useI18n()

const urlId = useId()
const textId = useId()

const urlValue = ref('')
const textValue = ref('')
const urlInputRef = ref<InstanceType<typeof Input> | null>(null)

const canConfirm = computed(() => urlValue.value.trim().length > 0)

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      urlValue.value = props.url ?? ''
      textValue.value = props.text ?? ''
      nextTick(() => {
        const el = urlInputRef.value?.$el as HTMLInputElement | undefined
        el?.focus()
      })
    }
  },
  { immediate: true },
)

function onOpenChange(value: boolean) {
  emit('update:open', value)
}

function close() {
  emit('update:open', false)
}

function onConfirm() {
  if (!canConfirm.value) return
  emit('save', { url: urlValue.value.trim(), text: textValue.value.trim() })
  close()
}

function onCancel() {
  close()
}

function onRemove() {
  emit('remove')
  close()
}
</script>
