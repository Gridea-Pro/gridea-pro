<template>
  <div class="pb-20 max-w-6xl mx-auto pt-4 px-4 w-full">
    <div class="space-y-8">
      <!-- Current Theme -->
      <div v-if="currentTheme" class="space-y-4">
        <h2 class="text-sm text-primary font-medium border-l-[3px] border-primary pl-3 flex items-center h-4">{{
          $t('settings.theme.current') }}</h2>
        <div class="border border-primary/20 rounded-xl overflow-hidden">
          <div class="flex flex-col md:flex-row h-full gap-6 md:gap-12">
            <!-- Cover Image -->
            <div class="relative w-full md:w-[280px] aspect-[4/3] group overflow-hidden bg-muted flex-shrink-0">
              <template v-if="getPreviewUrl(currentTheme)">
                <img :src="getPreviewUrl(currentTheme)"
                  class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105" alt="" />
              </template>
              <template v-else>
                <div class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105">
                  <PhotoIcon class="w-12 h-12 mb-2" />
                  <span class="text-[10px] uppercase tracking-wider font-medium">No Preview</span>
                </div>
              </template>

              <div
                class="absolute inset-0 bg-black/20 opacity-0 text-white gap-2 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center backdrop-blur-sm cursor-pointer"
                @click="previewTheme(currentTheme)">
                <EyeIcon class="w-8 h-8 text-white" />{{ $t('nav.preview') }}
              </div>
            </div>

            <!-- Info -->
            <div class="flex-1 py-6 md:py-8 pr-6 md:pr-8 flex flex-col justify-center">
              <div class="flex items-center gap-3 mb-4">
                <h3 class="text-2xl font-bold">{{ currentTheme.name }}</h3>
                <span
                  class="flex items-center justify-center w-12 h-5 px-2 bg-primary/10 border border-primary/20 rounded-full text-[12px] text-primary/80">
                  v{{ currentTheme.version }}
                </span>
              </div>

              <div class="space-y-4 text-muted-foreground">
                <p v-if="currentTheme.description" class="text-xs leading-relaxed">
                  {{ currentTheme.description }}
                </p>

                <div class="flex items-center gap-6 text-sm">
                  <div class="flex items-center gap-2" :title="t('author')">
                    <UserIcon class="w-4 h-4" />
                    <span>{{ currentTheme.author || 'Unknown' }}</span>
                  </div>

                  <div v-if="currentTheme.repository"
                    class="flex items-center gap-2 cursor-pointer hover:text-primary transition-colors"
                    @click="openRepo(currentTheme.repository)" :title="t('repository')">
                    <svg class="w-3.5 h-3.5 fill-current" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                      <path
                        d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
                    </svg>
                  </div>
                </div>
              </div>
            </div>

            <!-- Divider -->
            <div class="hidden md:block w-px bg-border my-6"></div>

            <!-- Right Actions Area -->
            <div class="w-full md:w-[260px] p-6 flex flex-col text-center">
              <div class="flex-1 space-y-6">
                <div>
                  <div class="font-medium truncate" :title="siteStore.site.themeConfig.siteName">
                    {{ siteStore.site.themeConfig.siteName }}
                  </div>
                </div>

                <div>
                  <div class="text-xs line-clamp-2 text-muted-foreground/80 leading-relaxed"
                    :title="siteStore.site.themeConfig.siteDescription">
                    {{ siteStore.site.themeConfig.siteDescription }}
                  </div>
                </div>

                <div>
                  <div class="text-xs line-clamp-2 text-muted-foreground/80 leading-relaxed"
                    :title="siteStore.site.themeConfig.siteAuthor">
                    {{ siteStore.site.themeConfig.siteAuthor }}
                  </div>
                </div>

              </div>

              <div class="mt-auto pt-6 text-[10px] text-muted-foreground/80 font-normal tracking-widest">
                Powered by <span class="hover:text-primary cursor-pointer transition-colors"
                  @click="BrowserOpenURL('https://gridea.pro')">Gridea Pro</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Other Themes -->
      <div v-if="otherThemes.length > 0" class="space-y-4 mt-16">
        <div class="flex items-center justify-between">
          <h2 class="text-sm text-primary font-medium border-l-[3px] border-primary pl-3 flex items-center h-4">{{
            $t('settings.theme.other') }}</h2>
          <div class="text-xs text-primary flex items-center gap-1 cursor-pointer hover:opacity-80 transition-opacity"
            @click="BrowserOpenURL('https://gridea.pro/themes/')">
            {{ $t('settings.theme.more') }}
            <ArrowTopRightOnSquareIcon class="w-3 h-3" />
          </div>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          <div v-for="theme in otherThemes" :key="theme.folder"
            class="group border border-primary/20 rounded-xl overflow-hidden hover:shadow-md transition-all duration-300 flex flex-col h-full">

            <!-- Cover -->
            <div class="relative aspect-[4/3] overflow-hidden bg-muted">
              <template v-if="getPreviewUrl(theme)">
                <img :src="getPreviewUrl(theme)"
                  class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                  loading="lazy" alt="" />
              </template>
              <template v-else>
                <div
                  class="w-full h-full flex flex-col items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 text-muted-foreground/30 transition-transform duration-500 group-hover:scale-105">
                  <PhotoIcon class="w-8 h-8 mb-1" />
                  <span class="text-[10px] uppercase tracking-wider font-medium">No Preview</span>
                </div>
              </template>

              <div
                class="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center backdrop-blur-sm cursor-pointer gap-4"
                @click="previewTheme(theme)">
                <span class="flex items-center gap-2 text-white font-medium">
                  <EyeIcon class="w-5 h-5" />
                  {{ $t('nav.preview') }}
                </span>
              </div>
            </div>

            <!-- Info -->
            <div class="p-4 flex-1 flex flex-col">
              <div class="flex justify-between items-center mb-2">
                <div class="flex items-center gap-4 overflow-hidden">
                  <h4 class="font-bold text-base truncate" :title="theme.name">{{ theme.name }}</h4>
                  <span
                    class="flex items-center justify-center h-4 px-2 py-2 shrink-0 bg-primary/10 rounded-full text-[10px] text-primary/80 border border-primary/20">
                    v{{ theme.version }}
                  </span>
                </div>

                <div v-if="theme.repository"
                  class="shrink-0 cursor-pointer text-muted-foreground hover:text-foreground transition-colors p-1 rounded-full"
                  @click.stop="openRepo(theme.repository)" :title="t('repository')">
                  <svg class="w-3.5 h-3.5 fill-current hover:text-primary" viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg">
                    <path
                      d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
                  </svg>
                </div>
              </div>

              <p class="text-xs text-muted-foreground line-clamp-2 mb-4 flex-1 h-[32px]">
                {{ theme.description }}
              </p>

              <div class="flex items-center justify-between pt-2 mt-auto border-t border-border/50">
                <span class="flex items-center gap-1 text-xs text-muted-foreground truncate max-w-[100px]"
                  :title="theme.author">
                  <UserIcon class="w-3 h-3" />
                  {{ theme.author }}
                </span>
                <Button size="sm" variant="secondary"
                  class="h-7 text-[10px] rounded-full bg-primary/5 border border-primary/10 text-primary hover:bg-primary hover:text-white transition-colors cursor-pointer"
                  @click="useTheme(theme)">
                  {{ $t('settings.theme.use') }}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { useImageUrl } from '@/composables/useImageUrl'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import { EventsEmit, EventsOnce, BrowserOpenURL } from 'wailsjs/runtime'
import { EyeIcon, UserIcon, PhotoIcon, AdjustmentsHorizontalIcon, ArrowTopRightOnSquareIcon } from '@heroicons/vue/24/outline'

const emit = defineEmits(['change-tab'])
const { t } = useI18n()
const siteStore = useSiteStore()
const { getImageUrl } = useImageUrl()
const imageVersion = ref(Date.now())

// Update image version when site store updates (triggered by real-time events)
watch(() => siteStore.site, () => {
  imageVersion.value = Date.now()
}, { deep: true })

const currentTheme = computed(() => {
  return siteStore.site.themes.find(t => t.folder === siteStore.site.themeConfig.themeName)
})

const otherThemes = computed(() => {
  return siteStore.site.themes.filter(t => t.folder !== siteStore.site.themeConfig.themeName)
})

const getPreviewUrl = (theme: any) => {
  if (!theme.previewImage) return ''
  const appDir = siteStore.site.appDir
  const path = `${appDir}/themes/${theme.folder}/${theme.previewImage}`
  // Add timestamp using reactive imageVersion
  return `${getImageUrl(path)}&t=${imageVersion.value}`
}

const openRepo = (url: string) => {
  if (url) BrowserOpenURL(url)
}

const previewTheme = (theme: any) => {
  // If it's the active theme, open local preview
  if (theme.folder === siteStore.site.themeConfig.themeName) {
    EventsEmit('preview-site')
    return
  }

  // For other themes, fallback to repository or toast
  if (theme.repository) {
    BrowserOpenURL(theme.repository)
  } else {
    toast.info(t('settings.theme.previewNotAvailable'))
  }
}

const useTheme = (theme: any) => {
  const form = { ...siteStore.site.themeConfig, themeName: theme.folder }

  // Reuse save logic from BasicSetting (simplified)
  EventsOnce('theme-saved', (result: boolean) => {
    if (result) {
      toast.success(t('settings.theme.configSaved'))
      EventsEmit('app-site-reload')
    } else {
      toast.error('Theme save failed')
    }
  })

  EventsEmit('theme-save', form)
}
</script>
