<template>
  <Sheet v-model:open="openModel">
    <SheetContent side="right" class="w-screen max-w-4xl sm:max-w-4xl p-0">
      <div class="preview-sheet">
        <ArticlePreviewContent
          :title="title"
          :date-formatted="dateFormatted"
          :tags="tags"
          :html-content="htmlContent"
        />
      </div>
    </SheetContent>
  </Sheet>
</template>

<script lang="ts" setup>
import { computed } from 'vue'

import { Sheet, SheetContent } from '@/components/ui/sheet'

import ArticlePreviewContent from './ArticlePreviewContent.vue'

const props = defineProps<{
  open: boolean
  title: string
  dateFormatted: string
  tags: string[]
  htmlContent?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const openModel = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value),
})
</script>

<style lang="less" scoped>
.preview-sheet {
  height: 100%;
  overflow-y: auto;
  background: var(--background);
  padding: 24px;
}
</style>
