<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        :title="t('category.new')" style="--wails-draggable: no-drag" @click="openCreateSheet">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable
v-model="categoryList" handle=".handle" item-key="slug" class="grid grid-cols-1 md:grid-cols-2 gap-4"
        @change="handleCategorySort">
        <template #item="{ element: category, index }">
          <CategoryCard :category="category" :index="index" @edit="openEditSheet" @delete="confirmDelete" />
        </template>
      </draggable>
    </div>

    <!-- Edit/New Drawer -->
    <CategoryEditor
v-model:open="visible" :form="form" :can-submit="canSubmit" @save="saveCategory" @close="closeSheet"
      @name-change="handleNameChange" @slug-change="handleSlugChange" />

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import { PlusIcon } from '@heroicons/vue/24/outline'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import CategoryCard from './components/CategoryCard.vue'
import CategoryEditor from './components/CategoryEditor.vue'
import { useCategory } from './composables/useCategory'

const { t } = useI18n()

const {
  visible,
  form,
  categoryList,
  deleteModalVisible,
  canSubmit,
  openCreateSheet,
  openEditSheet,
  closeSheet,
  saveCategory,
  confirmDelete,
  handleDelete,
  handleCategorySort,
  handleNameChange,
  handleSlugChange,
} = useCategory()

</script>
