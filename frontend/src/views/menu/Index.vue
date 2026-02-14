<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        @click="newMenu" :title="t('siteMenu.new')" style="--wails-draggable: no-drag">
        <PlusIcon class="size-4" />
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6">
      <draggable v-model="menuList" handle=".handle" item-key="name" @change="handleMenuSort">
        <template #item="{ element: menu, index }">
          <div
            class="group flex mb-4 rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
            @click="editMenu(menu, index)">
            <div class="flex items-center pl-4 handle cursor-move">
              <Bars3Icon class="size-3 text-muted-foreground" />
            </div>
            <div class="p-4 flex-1">
              <div class="text-sm font-medium text-foreground mb-2">
                {{ menu.name }}
              </div>
              <div class="text-xs flex items-center gap-3">
                <div
                  class="px-2 py-0.5 bg-primary/10 border border-primary/20 rounded-full text-[10px] text-primary/80">
                  {{ menu.openType }}
                  <ArrowTopRightOnSquareIcon class="w-3 h-3 ml-1" v-if="menu.openType === 'External'" />
                </div>
                <div class="text-muted-foreground truncate">
                  {{ menu.link }}
                </div>
              </div>
            </div>
            <div class="flex items-center px-4 gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
              <button
                class="p-2 text-muted-foreground hover:text-primary hover:bg-secondary rounded-lg transition-colors"
                @click.stop="editMenu(menu, index)" :title="t('common.edit')">
                <PencilIcon class="size-3" />
              </button>
              <button
                class="p-2 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-lg transition-colors"
                @click.stop="confirmDelete(index)" :title="t('common.delete')">
                <TrashIcon class="size-3" />
              </button>
            </div>
          </div>
        </template>
      </draggable>
    </div>

    <!-- Edit/New Drawer -->
    <Sheet :open="visible" @update:open="visible = $event">
      <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
        <SheetHeader class="px-6 py-6 border-b">
          <SheetTitle>{{ t('nav.menu') }}</SheetTitle>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
          <div class="space-y-4">
            <div>
              <Label class="mb-1 block">{{ t('siteMenu.name') }}</Label>
              <Input v-model="form.name" />
            </div>
            <div>
              <Label class="mb-1 block">Open Type</Label>
              <Select v-model="form.openType">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="item in menuTypes" :key="item" :value="item">
                    {{ item }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label class="mb-1 block">Link</Label>
              <div class="space-y-2">
                <Input v-model="form.link" placeholder="输入或从下面选择" />
                <Select v-model="form.link">
                  <SelectTrigger>
                    <SelectValue placeholder="选择内部链接..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="item in menuLinks" :key="item.value" :value="item.value">
                      <span class="truncate max-w-[300px] block" :title="item.text">{{ item.text }}</span>
                    </SelectItem>
                  </SelectContent>
                </Select>
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
            :disabled="!canSubmit" @click="saveMenu">{{ t('common.save') }}</Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <DeleteConfirmDialog v-model:open="deleteModalVisible" :confirm-text="t('common.delete')" @confirm="handleDelete" />

  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import urlJoin from 'url-join'
import Draggable from 'vuedraggable'
import { MenuTypes } from '@/helpers/enums'
import { IMenu } from '@/interfaces/menu'
import { IPost } from '@/interfaces/post'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'
import {
  PlusIcon,
  Bars3Icon,
  ArrowTopRightOnSquareIcon,
  TrashIcon,
  PencilIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit } from '@/wailsjs/runtime'

const { t } = useI18n()
const siteStore = useSiteStore()

interface IForm {
  name: any
  index: any
  openType: string
  link: string
}

const form = reactive<IForm>({
  name: '',
  index: '',
  openType: MenuTypes.Internal,
  link: '',
})

import { SaveMenuFromFrontend, DeleteMenuFromFrontend, SaveMenus } from '@/wailsjs/go/facade/MenuFacade'
import { domain, facade } from '@/wailsjs/go/models'

const menuList = ref<IMenu[]>([])
const visible = ref(false)
const menuTypes = MenuTypes
const deleteModalVisible = ref(false)
const menuToDelete = ref<number | null>(null)

const menuLinks = computed(() => {
  const { setting, themeConfig } = siteStore.site
  const posts = siteStore.posts.map((item: IPost) => {
    return {
      text: `📄 ${item.data.title}`,
      value: urlJoin(setting.domain, themeConfig.postPath, item.fileName),
    }
  })
  return [
    {
      text: '🏠 Homepage',
      value: setting.domain,
    },
    {
      text: '📚 Archives',
      value: urlJoin(setting.domain, themeConfig.archivesPath),
    },
    {
      text: '🏷️ Tags',
      value: urlJoin(setting.domain, themeConfig.tagPath || 'tags'),
    },
    ...posts,
  ].filter((item) => typeof item.value === 'string' && item.value.trim() !== '')
})

const canSubmit = computed(() => {
  return form.name && form.link
})

const newMenu = () => {
  form.name = null
  form.index = null
  form.openType = MenuTypes.Internal
  form.link = ''
  visible.value = true

  ga('Menu', 'Menu - new', siteStore.site.setting.domain)
}

const close = () => {
  visible.value = false
}

const editMenu = (menu: IMenu, index: number) => {
  visible.value = true
  form.index = index
  form.name = menu.name
  form.openType = menu.openType
  form.link = menu.link
}

const reloadSite = () => {
  EventsEmit('app-site-reload')
}

const saveMenu = async () => {
  try {
    const menuForm = new facade.MenuForm({
      name: form.name,
      openType: form.openType,
      link: form.link,
      index: form.index,
    })
    const menus = await SaveMenuFromFrontend(menuForm)

    if (menus) {
      siteStore.menus = menus
      menuList.value = [...menus]
      toast.success(t('siteMenu.saved'))
      visible.value = false
      ga('Menu', 'Menu - save', form.name)
    }
  } catch (e: any) {
    toast.error(e.message || 'Error saving menu')
  }
}

const confirmDelete = (index: number) => {
  menuToDelete.value = index
  deleteModalVisible.value = true
}

const handleDelete = async () => {
  if (menuToDelete.value !== null) {
    try {
      const menus = await DeleteMenuFromFrontend(menuToDelete.value)
      if (menus) {
        siteStore.menus = menus
        menuList.value = [...menus]
        toast.success(t('siteMenu.deleted'))
      }
    } catch (e: any) {
      toast.error(e.message || 'Error deleting menu')
    }
  }
  deleteModalVisible.value = false
  menuToDelete.value = null
}

const handleMenuSort = async () => {
  try {
    const menus = menuList.value.map(m => new domain.Menu(m))
    await SaveMenus(menus)
  } catch (e: any) {
    toast.error(e.message || 'Error sorting menu')
  }
}

onMounted(() => {
  menuList.value = [...siteStore.menus]
})
</script>
