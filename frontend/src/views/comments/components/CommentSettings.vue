<template>
  <Sheet v-model:open="visible">
    <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
      <SheetHeader class="px-6 py-6 border-b">
        <SheetTitle>{{ t('settings.comment.title') }}</SheetTitle>
      </SheetHeader>

      <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
        <!-- Form Content -->
        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <label class="text-sm font-medium text-[var(--text-primary)]">{{ t('settings.comment.enable') }}</label>
            <Switch
size="sm" class="focus-visible:ring-0 focus-visible:ring-offset-0" :checked="formState.enable"
              @update:checked="val => formState.enable = val" />
          </div>

          <div>
            <Label class="text-sm font-medium text-[var(--text-primary)] mb-2 block">{{ t('settings.comment.platform')
            }}</Label>
            <Select v-model="formState.platform">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="p in platforms" :key="p" :value="p">{{ p }}</SelectItem>
              </SelectContent>
            </Select>
            <div class="ml-1 mt-2 text-xs text-primary flex items-center gap-1 cursor-pointer" @click="openDocs">
              {{ t('settings.comment.viewDocs', { platform: formState.platform }) }}
              <ArrowTopRightOnSquareIcon class="size-3" />
            </div>
          </div>

          <div v-if="formState.enable" class="space-y-4 pt-4 border-t border-[var(--color-border)]">
            <template v-if="['Valine', 'Waline'].includes(formState.platform)">
              <!-- Shared config interface for Valine/Waline -->
              <ValineWalineConfig
v-if="formState.platform === 'Valine'"
                v-model:config="formState.platformConfigs['Valine']" />
              <ValineWalineConfig v-else v-model:config="formState.platformConfigs['Waline']" />
            </template>

            <template v-else-if="formState.platform === 'Twikoo'">
              <TwikooConfig v-model:config="formState.platformConfigs['Twikoo']" />
            </template>

            <template v-else-if="formState.platform === 'Gitalk'">
              <GitalkConfig v-model:config="formState.platformConfigs['Gitalk']" />
            </template>

            <template v-else-if="formState.platform === 'Disqus'">
              <DisqusConfig v-model:config="formState.platformConfigs['Disqus']" />
            </template>

            <template v-else-if="formState.platform === 'Cusdis'">
              <CusdisConfig v-model:config="formState.platformConfigs['Cusdis']" />
            </template>

            <template v-else-if="formState.platform === 'Giscus'">
              <GiscusConfig v-model:config="formState.platformConfigs['Giscus']" />
            </template>
          </div>
        </div>
      </div>
      <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
        <Button
variant="outline"
          class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
          @click="onClose">{{ t('common.cancel') }}</Button>
        <Button
variant="default"
          class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          @click="save">{{ t('common.save') }}</Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>

<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowTopRightOnSquareIcon } from '@heroicons/vue/24/outline'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useCommentStore } from '@/stores/comment'
import type { CommentSettings } from '@/types/comment'
import { BrowserOpenURL } from '@/wailsjs/runtime'

import ValineWalineConfig from './platforms/ValineWalineConfig.vue'
import TwikooConfig from './platforms/TwikooConfig.vue'
import GitalkConfig from './platforms/GitalkConfig.vue'
import DisqusConfig from './platforms/DisqusConfig.vue'
import CusdisConfig from './platforms/CusdisConfig.vue'
import GiscusConfig from './platforms/GiscusConfig.vue'

const props = defineProps<{
  open: boolean
}>()

const { t } = useI18n()

const emit = defineEmits(['update:open'])

const visible = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val)
})

const commentStore = useCommentStore()
const platforms = ['Valine', 'Waline', 'Twikoo', 'Gitalk', 'Cusdis', 'Giscus', 'Disqus']

const formState = ref<CommentSettings>({
  enable: false,
  platform: 'Valine',
  platformConfigs: {},
})

// Initialize form state from store
// Initialize form state from store
watch(() => props.open, async (newVal) => {
  if (newVal) {
    // Load settings from backend first
    await commentStore.loadSettings()

    // Safely update formState without losing reactivity references if possible, 
    // or just strict replacement.
    const newSettings = JSON.parse(JSON.stringify(commentStore.settings))

    // Ensure platformConfigs exists on the source
    if (!newSettings.platformConfigs) {
      newSettings.platformConfigs = {}
    }

    // Ensure all platform keys exist
    platforms.forEach(p => {
      if (!newSettings.platformConfigs[p]) {
        newSettings.platformConfigs[p] = {}
      }
    })

    formState.value = newSettings
  }
})

const onClose = () => {
  visible.value = false
}

import { toast } from '@/helpers/toast'

const save = async () => {
  // Ensure we are saving the current state
  const settingsToSave = {
    ...formState.value,
    platformConfigs: formState.value.platformConfigs || {}
  }

  // Prune nulls if necessary (JSON.stringify usually implies keys exist)
  // Ensure we are not sending null
  if (!settingsToSave.platformConfigs) {
    settingsToSave.platformConfigs = {}
  }

  await commentStore.saveSettings(settingsToSave)
  toast.success(t('settings.comment.saveSuccess'))
  onClose()
  // Refresh in background so UI doesn't freeze
  commentStore.fetchComments()
}

const openDocs = () => {
  const urls: Record<string, string> = {
    Gitalk: 'https://github.com/gitalk/gitalk',
    Disqus: 'https://disqus.com/',
    Valine: 'https://valine.js.org/',
    Waline: 'https://waline.js.org/',
    Twikoo: 'https://twikoo.js.org/',
    Cusdis: 'https://cusdis.com/',
    Giscus: 'https://giscus.app/',
  }
  const url = urls[formState.value.platform]
  if (url) {
    BrowserOpenURL(url)
  }
}
</script>
