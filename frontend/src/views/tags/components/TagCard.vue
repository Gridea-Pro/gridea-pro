<template>
    <div
class="group flex items-center justify-between h-16 p-5 rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
        @click="$emit('edit', tag, index)">
        <div class="flex items-center gap-2.5 flex-1 min-w-0">
            <div class="w-2 h-2 rounded-full flex-shrink-0" :style="{ backgroundColor: tag.color || '#888' }"></div>
            <div class="text-xs font-medium text-foreground truncate">
                {{ tag.name }}
            </div>
        </div>

        <div class="flex-shrink-0 ml-2 h-6 flex items-center">
            <div class="text-xs text-muted-foreground opacity-70 group-hover:hidden">
                {{ postCount }}
            </div>
            <div class="hidden group-hover:flex items-center gap-1">
                <button
                    class="p-1 text-muted-foreground hover:text-primary hover:bg-secondary rounded-md transition-colors"
                    :title="t('common.edit')" @click.stop="$emit('edit', tag, index)">
                    <PencilIcon class="size-3" />
                </button>
                <button
                    class="p-1 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-md transition-colors"
                    :title="t('common.delete')" @click.stop="$emit('delete', tag.slug || '')">
                    <TrashIcon class="size-3" />
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ITag } from '@/stores/site'
import { PencilIcon, TrashIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
    tag: ITag
    index: number
}>()

defineEmits<{
    (e: 'edit', tag: ITag, index: number): void
    (e: 'delete', slug: string): void
}>()

const { t } = useI18n()
const siteStore = useSiteStore()

const postCount = computed(() => {
    return siteStore.posts.filter((p) => p.published && (p.tags || []).includes(props.tag.name || '')).length
})
</script>
