<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">
    <div class="space-y-6">
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.imageHosting.enable') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.enabled" @update:checked="(v: boolean) => form.enabled = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.imageHosting.enableDesc') }}</span>
          <a href="https://s.ee" target="_blank" class="text-xs text-primary/70 hover:text-primary underline ml-1" style="--wails-draggable: no-drag">S.EE</a>
        </div>
      </div>

      <template v-if="form.enabled">
        <div class="grid grid-cols-[180px_1fr] items-start gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.imageHosting.apiKey') }}</label>
          <div class="max-w-sm">
            <div class="relative">
              <Input
                v-model="form.apiKey"
                :type="showKey ? 'text' : 'password'"
                :placeholder="t('settings.imageHosting.apiKeyPlaceholder')"
              />
              <button
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
                @click="showKey = !showKey"
              >
                <svg v-if="showKey" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>
              </button>
            </div>
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.imageHosting.apiKeyDesc') }}</div>
          </div>
        </div>

        <div class="border-t border-border pt-6 mt-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-medium">{{ t('settings.imageHosting.imageBrowser') }}</h3>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                class="h-7 text-xs rounded-full cursor-pointer"
                :disabled="listLoading"
                @click="loadImages(1)">
                {{ t('settings.imageHosting.refresh') }}
              </Button>
            </div>
          </div>

          <div v-if="listLoading" class="flex items-center justify-center py-12 text-muted-foreground text-sm">
            {{ t('settings.imageHosting.loading') }}
          </div>

          <template v-else-if="files.length > 0">
            <div class="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-3">
              <div
                v-for="file in files"
                :key="file.hash"
                class="group relative aspect-square rounded-lg overflow-hidden border border-border bg-muted cursor-pointer"
                @click="copyUrl(file.url)"
              >
                <img :src="file.url" :alt="file.filename" class="w-full h-full object-cover" />
                <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                  <button
                    class="text-white text-xs bg-red-500/80 hover:bg-red-500 rounded px-2 py-1 cursor-pointer"
                    @click.stop="deleteImage(file.hash)"
                  >
                    {{ t('common.delete') }}
                  </button>
                </div>
                <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent px-1.5 pb-1 pt-4">
                  <p class="text-[10px] text-white truncate">{{ file.filename }}</p>
                </div>
              </div>
            </div>

            <div class="flex items-center justify-center gap-2 mt-4">
              <Button
                variant="outline"
                class="h-7 w-7 text-xs rounded-full cursor-pointer"
                :disabled="currentPage <= 1"
                @click="loadImages(currentPage - 1)">
                &lt;
              </Button>
              <span class="text-xs text-muted-foreground">{{ currentPage }}</span>
              <Button
                variant="outline"
                class="h-7 w-7 text-xs rounded-full cursor-pointer"
                :disabled="!hasNextPage"
                @click="loadImages(currentPage + 1)">
                &gt;
              </Button>
            </div>
          </template>

          <div v-else class="flex items-center justify-center py-12 text-muted-foreground text-sm">
            {{ t('settings.imageHosting.noImages') }}
          </div>
        </div>
      </template>
    </div>

    <footer-box>
      <div class="flex justify-end items-center gap-3 w-full">
        <Button
          variant="default"
          class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          @click="submit">
          {{ t('common.save') }}
        </Button>
      </div>
    </footer-box>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { GetSetting, SaveSetting, ListImages, DeleteImage } from '@/wailsjs/go/facade/ImageHostingFacade'
import { domain } from '@/wailsjs/go/models'

const { t } = useI18n()

const showKey = ref(false)
const listLoading = ref(false)

const form = reactive({
  enabled: false,
  apiKey: '',
})

const files = ref<domain.ImageHostingFile[]>([])
const currentPage = ref(1)
const hasNextPage = ref(true)

const loadImages = async (page: number) => {
  listLoading.value = true
  try {
    const resp = await ListImages(page)
    currentPage.value = page
    if (resp.data && Array.isArray(resp.data) && resp.data.length > 0) {
      files.value = resp.data
      hasNextPage.value = true
    } else {
      files.value = resp.data || []
      hasNextPage.value = false
    }
  } catch (e: any) {
    console.error(e)
    if (e.message) {
      toast.error(e.message)
    }
  } finally {
    listLoading.value = false
  }
}

const copyUrl = async (url: string) => {
  try {
    await navigator.clipboard.writeText(url)
    toast.success(t('settings.imageHosting.copied'))
  } catch {
    toast.error(t('settings.imageHosting.copyFailed'))
  }
}

const deleteImage = async (hash: string) => {
  try {
    await DeleteImage(hash)
    toast.success(t('settings.imageHosting.deleteSuccess'))
    loadImages(currentPage.value)
  } catch (e: any) {
    console.error(e)
    toast.error(e.message || t('settings.imageHosting.deleteFailed'))
  }
}

onMounted(async () => {
  try {
    const setting = await GetSetting()
    if (setting) {
      form.enabled = setting.enabled || false
      form.apiKey = setting.apiKey || ''
    }
  } catch (e) {
    console.error('Failed to load image hosting settings', e)
  }
})

const submit = async () => {
  try {
    const s = new domain.ImageHostingSetting(form)
    await SaveSetting(s)
    toast.success(t('common.saveSuccess'))
  } catch (e: any) {
    console.error(e)
    toast.error(e.message || t('common.saveFailed'))
  }
}
</script>
