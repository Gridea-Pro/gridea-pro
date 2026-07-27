<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t('editor.imageDialog.insert') }}</DialogTitle>
      </DialogHeader>

      <Tabs v-model="tab" class="w-full">
        <TabsList class="grid w-full grid-cols-2">
          <TabsTrigger value="url">{{ t('editor.imageDialog.fromUrl') }}</TabsTrigger>
          <TabsTrigger value="local">{{ t('editor.imageDialog.fromLocal') }}</TabsTrigger>
        </TabsList>

        <TabsContent value="url" class="space-y-4 pt-2">
          <Input
            v-model="url"
            :placeholder="t('editor.imageDialog.urlPlaceholder')"
            @keydown.enter.prevent="confirmUrl"
          />
          <DialogFooter>
            <Button variant="outline" @click="close">{{ t('editor.imageDialog.cancel') }}</Button>
            <Button :disabled="!canInsertUrl" @click="confirmUrl">{{ t('editor.imageDialog.insert') }}</Button>
          </DialogFooter>
        </TabsContent>

        <TabsContent value="local" class="space-y-4 pt-2">
          <p class="text-sm text-muted-foreground">
            {{ t('editor.imageDialog.localHint') }}
          </p>
          <DialogFooter>
            <Button variant="outline" @click="close">{{ t('editor.imageDialog.cancel') }}</Button>
            <Button @click="confirmLocal">
              <ImagePlus class="mr-2 h-4 w-4" />
              {{ t('editor.imageDialog.pickLocal') }}
            </Button>
          </DialogFooter>
        </TabsContent>
      </Tabs>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ImagePlus } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  'insert-url': [src: string]
  'pick-local': []
}>()

const { t } = useI18n()

const tab = ref<'url' | 'local'>('url')
const url = ref('')

const canInsertUrl = computed(() => url.value.trim().length > 0)

watch(
  () => props.open,
  (value) => {
    if (value) {
      tab.value = 'url'
      url.value = ''
    }
  },
)

function onOpenChange(value: boolean) {
  emit('update:open', value)
}

function close() {
  emit('update:open', false)
}

function confirmUrl() {
  const src = url.value.trim()
  if (!src) return
  emit('insert-url', src)
  close()
}

function confirmLocal() {
  emit('pick-local')
  close()
}
</script>
