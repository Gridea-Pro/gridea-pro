<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        :title="t('siteMenu.new')" style="--wails-draggable: no-drag" @click="newMenu">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable v-model="menuList" handle=".handle" item-key="name" @change="handleMenuSort">
        <template #item="{ element: menu, index }">
          <MenuCard
            :menu="menu"
            :index="index"
            @edit="editMenu"
            @delete="confirmDelete"
            @add-child="newSubMenu"
            @edit-child="editSubMenu"
            @delete-child="confirmDeleteChild"
          />
        </template>
      </draggable>
    </div>

    <!-- Edit/New Drawer -->
    <MenuEditor
      v-model:open="visible"
      :form="form"
      :menu-types="menuTypes"
      :menu-links="menuLinks"
      :can-submit="canSubmit"
      :parent-name="parentName"
      @name-change="handleNameChange"
      @open-type-change="handleOpenTypeChange"
      @link-change="handleLinkChange"
      @close="closeSheet"
      @save="saveMenu"
    />

    <DeleteConfirmDialog v-model:open="deleteModalVisible" :confirm-text="t('common.delete')" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import { PlusIcon } from '@heroicons/vue/24/outline'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import MenuCard from './components/MenuCard.vue'
import MenuEditor from './components/MenuEditor.vue'
import { useMenu } from './composables/useMenu'

const { t } = useI18n()

const {
  menuList,
  visible,
  menuTypes,
  deleteModalVisible,
  form,
  menuLinks,
  canSubmit,
  parentName,
  newMenu,
  newSubMenu,
  closeSheet,
  editMenu,
  editSubMenu,
  saveMenu,
  confirmDelete,
  confirmDeleteChild,
  handleDelete,
  handleMenuSort,
  handleNameChange,
  handleOpenTypeChange,
  handleLinkChange,
} = useMenu()
</script>
