<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        @click="newCategory" :title="t('category.new')" style="--wails-draggable: no-drag">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable v-model="categoryList" handle=".handle" item-key="slug" @change="handleCategorySort"
        class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <template #item="{ element: category, index }">
          <div
            class="group relative flex rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
            @click="editCategory(category, index)">
            <div class="flex items-center pl-4 handle cursor-move">
              <Bars3Icon class="size-3 text-muted-foreground" />
            </div>
            <div class="p-4 flex-1 min-w-0">
              <div class="text-xs font-medium text-foreground mb-1 truncate group-hover:text-primary">
                {{ category.name }}
              </div>
              <div class="text-xs font-normal text-muted-foreground opacity-50 truncate" v-if="category.description">
                {{ category.description }}
              </div>
              <div class="text-xs font-normal text-muted-foreground opacity-50 truncate" v-else>
                /{{ category.slug }}
              </div>
            </div>
            <div class="flex items-center px-4">
              <div class="text-xs text-muted-foreground opacity-70 group-hover:hidden">
                {{siteStore.posts.filter(p => p.data.published && (p.data.categories ||
                  []).includes(category.name)).length }}
              </div>
              <div class="hidden group-hover:flex items-center gap-2">
                <button
                  class="p-2 text-muted-foreground hover:text-primary hover:bg-secondary rounded-lg transition-colors"
                  @click.stop="editCategory(category, index)" :title="t('common.edit')">
                  <PencilIcon class="size-3" />
                </button>
                <button
                  class="p-2 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-lg transition-colors"
                  @click.stop="confirmDelete(category.slug)" :title="t('common.delete')">
                  <TrashIcon class="size-3" />
                </button>
              </div>
            </div>
          </div>
        </template>
      </draggable>
    </div>

    <!-- Edit/New Drawer -->
    <Sheet v-model:open="visible">
      <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
        <SheetHeader class="px-6 py-6 border-b">
          <SheetTitle>{{ t('nav.category') }}</SheetTitle>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
          <div class="space-y-4">
            <div>
              <Label class="mb-1 block">{{ t('category.name') }} <span class="text-destructive">*</span></Label>
              <Input v-model="form.name" @input="handleNameChange" />
            </div>
            <div>
              <Label class="mb-1 block">{{ t('category.url') }} <span class="text-destructive">*</span></Label>
              <Input v-model="form.slug" @input="handleSlugChange" />
            </div>
            <div>
              <Label class="mb-1 block">{{ t('category.description') }}</Label>
              <Textarea v-model="form.description" rows="3" />
            </div>
          </div>
        </div>
        <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
          <Button variant="outline"
            class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
            @click="close">{{ t('common.cancel') }}</Button>
          <Button variant="default"
            class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            :disabled="!canSubmit" @click="saveCategory">{{ t('common.save') }}</Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ICategory } from '@/stores/site'
import shortid from 'shortid'
import Draggable from 'vuedraggable'
import slug from '@/helpers/slug'
import { toast } from '@/helpers/toast'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'
import {
  PlusIcon,
  Bars3Icon,
  TrashIcon,
  PencilIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit, EventsOnce, EventsOff } from 'wailsjs/runtime'

const { t } = useI18n()
const siteStore = useSiteStore()

const visible = ref(false)
const isUpdate = ref(false)
const slugChanged = ref(false)
const deleteModalVisible = ref(false)
const categoryToDelete = ref<string | null>(null)

interface IForm {
  name: string
  slug: string
  description: string
  index: number
  originalSlug?: string
}

const form = reactive<IForm>({
  name: '',
  slug: '',
  description: '',
  index: -1,
  originalSlug: '',
})

const categoryList = ref<ICategory[]>([])

const canSubmit = computed(() => {
  return form.name && form.slug
})

const handleNameChange = (e: any) => {
  const val = e.target.value || form.name // Handle both Event and direct value if needed, but Input emits update:modelValue
  if (!slugChanged.value) {
    form.slug = slug(val)
  }
}

const handleSlugChange = (e: any) => {
  const val = e.target.value || form.slug
  slugChanged.value = !!val
}

const close = () => {
  visible.value = false
}

const newCategory = () => {
  form.name = ''
  form.slug = ''
  form.description = ''
  form.index = -1
  form.originalSlug = ''
  slugChanged.value = false
  visible.value = true
  isUpdate.value = false
}

const buildSlug = () => {
  if (form.slug === '') {
    form.slug = slug(form.name) || shortid.generate()
  }
}

const editCategory = (category: ICategory, index: number) => {
  visible.value = true
  isUpdate.value = true
  form.name = category.name
  form.slug = category.slug
  form.description = category.description || ''
  form.index = index
  form.originalSlug = category.slug
  slugChanged.value = true
}

const checkCategoryValid = () => {
  const categories = [...siteStore.categories]
  if (isUpdate.value) {
    categories.splice(form.index, 1)
  }
  const foundIndex = categories.findIndex((c: ICategory) => c.slug === form.slug)
  return foundIndex === -1
}

const saveCategory = () => {
  buildSlug()

  const valid = checkCategoryValid()
  if (!valid) {
    toast.error(t('category.urlRepeat'))
    return
  }

  EventsEmit('category-save', { ...form })
  EventsOnce('category-saved', (result: any) => {
    if (result.success && result.categories) {
      siteStore.categories = result.categories
      categoryList.value = [...result.categories]
    }
    toast.success(t('category.saved'))
    visible.value = false
  })
}

const confirmDelete = (slug: string) => {
  categoryToDelete.value = slug
  deleteModalVisible.value = true
}

const handleDelete = () => {
  if (categoryToDelete.value) {
    EventsEmit('category-delete', categoryToDelete.value)
    EventsOnce('category-deleted', (result: any) => {
      if (result.success && result.categories) {
        siteStore.categories = result.categories
        categoryList.value = [...result.categories]
      }
      toast.success(t('category.deleted'))
    })
  }
  deleteModalVisible.value = false
  categoryToDelete.value = null
}

const handleCategorySort = () => {
  EventsEmit('category-sort', JSON.parse(JSON.stringify(categoryList.value)))
}

onMounted(() => {
  categoryList.value = [...siteStore.categories]
  EventsOff('category-saved')
  EventsOff('category-deleted')
})

onUnmounted(() => {
  EventsOff('category-saved')
  EventsOff('category-deleted')
})
</script>
