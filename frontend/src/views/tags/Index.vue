<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        @click="newTag" :title="t('tag.new')" style="--wails-draggable: no-drag">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable v-model="tagList" handle=".handle" item-key="slug" @change="handleTagSort"
        class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <template #item="{ element: tag, index }">
          <div
            class="group flex items-center justify-between h-16 p-5 rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
            @click="editTag(tag, index)">
            <div class="flex items-center gap-2.5 flex-1 min-w-0">
              <div class="w-2 h-2 rounded-full flex-shrink-0" :style="{ backgroundColor: tag.color || '#888' }"></div>
              <div class="text-xs font-medium text-foreground truncate">
                {{ tag.name }}
              </div>
            </div>

            <div class="flex-shrink-0 ml-2 h-6 flex items-center">
              <div class="text-xs text-muted-foreground opacity-70 group-hover:hidden">
                {{siteStore.posts.filter(p => p.data.published && (p.data.tags || []).includes(tag.name)).length}}
              </div>
              <div class="hidden group-hover:flex items-center gap-1">
                <button
                  class="p-1 text-muted-foreground hover:text-primary hover:bg-secondary rounded-md transition-colors"
                  @click.stop="editTag(tag, index)" :title="t('common.edit')">
                  <PencilIcon class="size-3" />
                </button>
                <button
                  class="p-1 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-md transition-colors"
                  @click.stop="confirmDelete(tag.slug)" :title="t('common.delete')">
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
          <SheetTitle>{{ t('nav.tag') }}</SheetTitle>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
          <div class="space-y-4">
            <div>
              <Label class="mb-1.5 block">{{ t('tag.name') }} <span class="text-destructive">*</span></Label>
              <Input v-model="form.name" @input="handleNameChange" />
            </div>
            <div>
              <Label class="mb-1.5 block">{{ t('tag.slug') }} <span class="text-destructive">*</span></Label>
              <div class="relative">
                <span class="absolute left-3 top-2.5 text-muted-foreground text-sm">/tags/</span>
                <Input v-model="form.slug" @input="handleSlugChange" class="pl-14" />
              </div>
            </div>
            <div>
              <Label class="mb-3 block">{{ t('tag.color') }}</Label>
              <div class="flex flex-wrap gap-2">
                <div v-for="color in PRESET_COLORS" :key="color"
                  class="w-6 h-6 rounded-full cursor-pointer transition-transform hover:scale-110 border border-transparent"
                  :class="{ 'ring-2 ring-primary ring-offset-2': form.color === color }"
                  :style="{ backgroundColor: color }" @click="form.color = color"></div>
                <div class="relative w-6 h-6 rounded-full overflow-hidden border border-border cursor-pointer">
                  <input type="color" v-model="form.color"
                    class="absolute inset-0 w-full h-full opacity-0 cursor-pointer" title="Custom Color" />
                  <div
                    class="absolute inset-0 bg-gradient-to-br from-red-500 via-green-500 to-blue-500 pointer-events-none"
                    v-if="!PRESET_COLORS.includes(form.color || '')"></div>
                  <div class="absolute inset-0" :style="{ backgroundColor: form.color }" v-else></div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
          <Button variant="outline"
            class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
            @click="close">{{ t('common.cancel') }}</Button>
          <Button variant="default"
            class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            :disabled="!canSubmit" @click="saveTag">{{ t('common.save') }}</Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ITag } from '@/stores/site'
import shortid from 'shortid'
import Draggable from 'vuedraggable'
import slug from '@/helpers/slug'
import { toast } from '@/helpers/toast'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'
import {
  PlusIcon,
  TrashIcon,
  PencilIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit } from '@/wailsjs/runtime'

const { t } = useI18n()
const siteStore = useSiteStore()

const visible = ref(false)
const isUpdate = ref(false)
const slugChanged = ref(false)
const deleteModalVisible = ref(false)
const tagToDelete = ref<string | null>(null)

interface IForm {
  name: string
  slug: string
  index: number
  originalSlug?: string
  originalName?: string
  color?: string
}

const form = reactive<IForm>({
  name: '',
  slug: '',
  index: -1,
  originalSlug: '',
  originalName: '',
  color: '#3b82f6',
})

import { GetTagColors, SaveTagFromFrontend, DeleteTagFromFrontend } from '@/wailsjs/go/facade/TagFacade'
import { domain, facade } from '@/wailsjs/go/models'

const PRESET_COLORS = ref<string[]>([])

onMounted(async () => {
  PRESET_COLORS.value = await GetTagColors()
  tagList.value = [...siteStore.tags]
})

const tagList = ref<ITag[]>([])

const canSubmit = computed(() => {
  return form.name && form.slug
})

const handleNameChange = (e: any) => {
  const val = e.target.value || form.name
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

const newTag = () => {
  form.name = ''
  form.slug = ''
  form.index = -1
  form.originalSlug = ''
  form.originalName = ''
  if (PRESET_COLORS.value.length > 0) {
    form.color = PRESET_COLORS.value[Math.floor(Math.random() * PRESET_COLORS.value.length)]
  } else {
    form.color = '#3b82f6'
  }
  slugChanged.value = false
  visible.value = true
  isUpdate.value = false
}

const buildSlug = () => {
  if (form.slug === '') {
    form.slug = slug(form.name) || shortid.generate()
  }
}

const editTag = (tag: ITag, index: number) => {
  visible.value = true
  isUpdate.value = true
  form.name = tag.name
  form.slug = tag.slug || ''
  form.index = index
  form.originalSlug = tag.slug || ''
  form.originalName = tag.name
  form.color = tag.color || '#3b82f6'
  slugChanged.value = true
}

const checkTagValid = () => {
  const tags = [...siteStore.tags]
  if (isUpdate.value) {
    tags.splice(form.index, 1)
  }
  const foundIndex = tags.findIndex((t: ITag) => t.slug === form.slug)
  return foundIndex === -1
}

const saveTag = async () => {
  buildSlug()

  const valid = checkTagValid()
  if (!valid) {
    toast.error(t('tagUrlRepeat')) // TODO: Check i18n key
    return
  }

  try {
    const tagForm = new facade.TagForm({
      name: form.name,
      slug: form.slug,
      color: form.color || '',
      originalName: form.originalName || '',
    })
    const newTags = await SaveTagFromFrontend(tagForm)

    if (newTags) {
      siteStore.tags = newTags
      tagList.value = [...newTags]
      toast.success(t('tagSaved'))
      visible.value = false
    }
  } catch (e: any) {
    toast.error(e.message || 'Error saving tag')
  }
}

const confirmDelete = (slug: string) => {
  tagToDelete.value = slug
  deleteModalVisible.value = true
}

const handleDelete = async () => {
  if (tagToDelete.value) {
    // Find tag name by slug because backend DeleteTag expects name currently (based on my refactor analysis, 
    // but wait, DeleteTagFromFrontend calls DeleteTag which calls internal.DeleteTag(ctx, name).
    // Let's check if we have the name available.
    // The previous loop found the tag name by slug. I should do the same or pass the name if possible.
    // However, the confirmDelete only sets slug.

    // Let's find the name from the slug
    const tag = siteStore.tags.find(t => t.slug === tagToDelete.value)
    if (tag) {
      try {
        const newTags = await DeleteTagFromFrontend(tag.name)
        if (newTags) {
          siteStore.tags = newTags
          tagList.value = [...newTags]
          toast.success(t('tagDeleted'))
        }
      } catch (e: any) {
        toast.error(e.message || 'Error deleting tag')
      }
    }
  }
  deleteModalVisible.value = false
  tagToDelete.value = null
}

const handleTagSort = () => {
  const tags = tagList.value.map(t => new domain.Tag(t))
  EventsEmit('tag-sort', tags)
}

watch(() => siteStore.tags, (newTags) => {
  tagList.value = [...newTags]
})

onMounted(() => {
  tagList.value = [...siteStore.tags]
})
</script>
