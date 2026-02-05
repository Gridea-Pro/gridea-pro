<template>
  <div class="mb-4 py-4 border-b border-border">
    <div class="text-base font-medium mb-4 text-foreground">{{ t('sourceFolder') }}</div>
    <div class="space-y-4">
      <div class="flex items-center gap-2">
        <Input 
          v-model="currentFolderPath" 
          readonly 
          class="flex-1"
        />
        <Button 
          variant="outline"
          size="icon"
          @click="handleFolderSelect"
        >
          <FolderOpenIcon class="size-5" />
        </Button>
      </div>
      <div>
        <Button 
          @click="save" 
        >
          {{ t('save') }}
        </Button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import { FolderOpenIcon } from '@heroicons/vue/24/outline'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { EventsEmit, EventsOnce } from 'wailsjs/runtime'
import { OpenFolderDialog } from 'wailsjs/wailsjs/go/app/App'

const { t } = useI18n()
const siteStore = useSiteStore()

const currentFolderPath = ref('-')

onMounted(() => {
  currentFolderPath.value = siteStore.appDir
})

const save = () => {
  EventsEmit('app-source-folder-setting', currentFolderPath.value)
  EventsOnce('app-source-folder-set', (data: any) => {
    if (data) {
      toast.success(t('saved'))
      EventsEmit('app-site-reload')
      EventsEmit('app-relaunch')
    } else {
      toast.error(t('saveError'))
    }
  })
}

const handleFolderSelect = async () => {
  const filePaths = await OpenFolderDialog()
  if (filePaths && filePaths.length > 0) {
    currentFolderPath.value = filePaths[0].replace(/\\/g, '/')
  }
}
</script>
