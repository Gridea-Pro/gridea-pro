<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        @click="newLink" :title="t('link.new')" style="--wails-draggable: no-drag">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable v-model="linkList" handle=".drag-handle" item-key="id" @change="handleLinkSort"
        class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <template #item="{ element: link, index }">
          <div
            class="group flex flex-col p-4 rounded-xl relative transition-all duration-200 bg-primary/2 border border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5">
            <div class="flex items-start gap-3 mb-2 pr-2">
              <div
                class="w-10 h-10 rounded-full flex-shrink-0 bg-primary/10 flex items-center justify-center overflow-hidden">
                <img :src="link.avatar || generateAvatar(link.name || link.id)" :alt="link.name"
                  class="w-full h-full object-cover" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-sm font-medium text-foreground truncate mb-1">
                  {{ link.name }}
                </div>
                <div class="text-xs text-muted-foreground truncate">
                  {{ link.url }}
                </div>
              </div>
            </div>

            <div v-if="link.description" class="text-xs text-muted-foreground opacity-70 line-clamp-2 mb-2">
              {{ link.description }}
            </div>

            <div class="flex items-center justify-between mt-auto pt-2 border-t border-border/50">
              <!-- Drag Handle (Left) -->
              <div class="drag-handle p-1.5 -ml-1.5 text-muted-foreground hover:text-foreground cursor-move">
                <Bars2Icon class="size-3" />
              </div>

              <!-- Action Buttons (Right) -->
              <div class="flex items-center gap-1">
                <button
                  class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors"
                  @click.stop="openLink(link.url)" :title="t('nav.preview')">
                  <EyeIcon class="size-3" />
                </button>
                <button
                  class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors"
                  @click.stop="editLink(link, index)" :title="t('common.edit')">
                  <PencilIcon class="size-3" />
                </button>
                <button
                  class="p-1.5 text-muted-foreground hover:text-destructive hover:bg-primary/10 rounded-md transition-colors"
                  @click.stop="confirmDelete(link.id)" :title="t('common.delete')">
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
          <SheetTitle>{{ t('nav.link') }}</SheetTitle>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
          <div class="space-y-4">
            <div>
              <Label class="mb-1.5 block">{{ t('link.name') }} <span class="text-destructive">*</span></Label>
              <Input v-model="form.name" @input="handleNameChange" />
            </div>
            <div>
              <Label class="mb-1.5 block">{{ t('link.url') }} <span class="text-destructive">*</span></Label>
              <Input v-model="form.url" placeholder="https://example.com" />
            </div>
            <div>
              <Label class="mb-1.5 block">{{ t('link.avatar') }}</Label>
              <Input v-model="form.avatar" placeholder="https://example.com/avatar.png" />
            </div>
            <div>
              <Label class="mb-1.5 block">{{ t('link.description') }}</Label>
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
            :disabled="!canSubmit" @click="saveLink">{{ t('common.save') }}</Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ILink } from '@/stores/site'
import { customAlphabet } from 'nanoid'
import Draggable from 'vuedraggable'
import slug from '@/helpers/slug'
import { toast } from '@/helpers/toast'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'
import {
  PlusIcon,
  TrashIcon,
  PencilIcon,
  GlobeAltIcon,
  Bars2Icon,
  EyeIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit, EventsOnce, EventsOff, BrowserOpenURL } from 'wailsjs/runtime'
import { generateAvatar } from '@/utils/avatarGenerator'

const { t } = useI18n()
const siteStore = useSiteStore()

const nanoid = customAlphabet('0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz', 6)

// State
const visible = ref(false)
const isUpdate = ref(false)
const deleteModalVisible = ref(false)
const linkToDelete = ref<string | null>(null)

interface IForm {
  name: string
  url: string
  description: string
  avatar: string
  id: string
  index: number
}

const form = reactive<IForm>({
  name: '',
  url: '',
  description: '',
  avatar: '',
  id: '',
  index: -1,
})

const linkList = ref<ILink[]>([])

const canSubmit = computed(() => {
  return form.name && form.url
})

const handleNameChange = (e: any) => {
  const val = e.target.value || form.name
}

const close = () => {
  visible.value = false
}

const newLink = () => {
  form.name = ''
  form.url = ''
  form.description = ''
  form.avatar = ''
  form.id = ''
  form.index = -1
  visible.value = true
  isUpdate.value = false
}

const buildId = () => {
  if (form.id === '') {
    form.id = nanoid()
  }
}

const editLink = (link: ILink, index: number) => {
  visible.value = true
  isUpdate.value = true
  form.name = link.name
  form.url = link.url
  form.description = link.description || ''
  form.avatar = link.avatar || ''
  form.id = link.id // Ensure ID is transmitted
  form.index = index
}

const saveLink = () => {
  buildId()

  EventsEmit('link-save', { ...form })
  EventsOnce('link-saved', (result: any) => {
    if (result.success && result.links) {
      siteStore.links = result.links
      linkList.value = [...result.links]
    }
    toast.success(t('link.saved'))
    visible.value = false
  })
}

const confirmDelete = (id: string) => {
  linkToDelete.value = id
  deleteModalVisible.value = true
}

const handleDelete = () => {
  if (linkToDelete.value) {
    EventsEmit('link-delete', linkToDelete.value)
    EventsOnce('link-deleted', (result: any) => {
      if (result.success && result.links) {
        siteStore.links = result.links
        linkList.value = [...result.links]
      }
      toast.success(t('link.deleted'))
    })
  }
  deleteModalVisible.value = false
  linkToDelete.value = null
}

const openLink = (url: string) => {
  BrowserOpenURL(url)
}

const handleLinkSort = () => {
  EventsEmit('link-sort', JSON.parse(JSON.stringify(linkList.value)))
}

onMounted(() => {
  linkList.value = [...siteStore.links]
  EventsOff('link-saved')
  EventsOff('link-deleted')
})

onUnmounted(() => {
  EventsOff('link-saved')
  EventsOff('link-deleted')
})
</script>
