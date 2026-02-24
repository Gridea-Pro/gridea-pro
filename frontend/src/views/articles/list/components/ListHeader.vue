<template>
    <div class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background border-b border-border select-none"
        style="--wails-draggable: drag">
        <div class="flex items-center min-h-[32px]" style="--wails-draggable: no-drag">
            <Transition enter-active-class="transition ease-out duration-200"
                enter-from-class="opacity-0 translate-x-[-10px]" enter-to-class="opacity-100 translate-x-0"
                leave-active-class="transition ease-in duration-150" leave-from-class="opacity-100 translate-x-0"
                leave-to-class="opacity-0 translate-x-[-10px]">
                <div v-if="selectedCount > 0" @click="$emit('deleteSelected')"
                    class="flex items-center py-1.5 px-3 bg-destructive/10 text-destructive cursor-pointer hover:bg-destructive/20 rounded-md text-xs transition-colors border border-destructive/20">
                    <TrashIcon class="size-3 mr-2 mb-0.5" />
                    <span>{{ t('common.delete') }} {{ selectedCount }}</span>
                </div>
            </Transition>
        </div>
        <div class="flex items-center gap-3" style="--wails-draggable: no-drag">
            <SearchInput v-model="searchModel" :placeholder="t('article.search')" />
            <div class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
                @click="$emit('newArticle')" :title="t('article.new')">
                <PlusIcon class="size-4" />
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchInput from '@/components/Base/SearchInput.vue'
import { TrashIcon, PlusIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
    keyword: string
    selectedCount: number
}>()

const emit = defineEmits<{
    'update:keyword': [value: string]
    deleteSelected: []
    newArticle: []
}>()

const { t } = useI18n()

const searchModel = computed({
    get: () => props.keyword,
    set: (val: string) => emit('update:keyword', val),
})
</script>
