<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        :title="t('tag.new')" style="--wails-draggable: no-drag" @click="openCreateSheet">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable
v-model="tagList" handle=".handle" item-key="slug" class="grid grid-cols-1 md:grid-cols-4 gap-4"
        @change="handleTagSort">
        <template #item="{ element: tag, index }">
          <TagCard :tag="tag" :index="index" @edit="editTag" @delete="confirmDelete" />
        </template>
      </draggable>
    </div>

    <!-- Edit/New Drawer -->
    <TagEditor
v-model:open="visible" :form="form" :preset-colors="presetColors" :can-submit="canSubmit"
      @name-change="handleNameChange" @slug-change="handleSlugChange" @color-change="handleColorChange"
      @close="closeSheet" @save="saveTag" />

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import { PlusIcon } from '@heroicons/vue/24/outline'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import TagCard from './components/TagCard.vue'
import TagEditor from './components/TagEditor.vue'
import { useTag } from './composables/useTag'

const { t } = useI18n()

const {
  visible,
  form,
  tagList,
  presetColors,
  deleteModalVisible,
  canSubmit,
  openCreateSheet,
  editTag,
  closeSheet,
  saveTag,
  confirmDelete,
  handleDelete,
  handleTagSort,
  handleNameChange,
  handleSlugChange,
  handleColorChange,
} = useTag()
</script>
