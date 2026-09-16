<template>
    <div
class="group relative flex rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
        @click="$emit('edit', category, index)">
        <div class="flex items-center pl-4 handle cursor-move">
            <Bars3Icon class="size-3 text-muted-foreground" />
        </div>
        <div v-if="coverUrl" class="flex items-center pl-3">
            <img :src="coverUrl" class="w-10 h-10 rounded-md object-cover border border-primary/10" />
        </div>
        <div class="p-4 flex-1 min-w-0">
            <div class="text-xs font-medium text-foreground mb-1 truncate group-hover:text-primary">
                {{ category.name }}
            </div>
            <div v-if="category.description" class="text-xs font-normal text-muted-foreground opacity-50 truncate">
                {{ category.description }}
            </div>
            <div v-else class="text-xs font-normal text-muted-foreground opacity-50 truncate">
                /{{ category.slug }}
            </div>
        </div>
        <div class="flex items-center px-4">
            <div class="text-xs text-muted-foreground opacity-70 group-hover:hidden">
                {{ postCount }}
            </div>
            <div class="hidden group-hover:flex items-center gap-2">
                <button
                    class="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    :title="t('common.edit')" @click.stop="$emit('edit', category, index)">
                    <PencilIcon class="size-3" />
                </button>
                <button
                    class="p-2 text-muted-foreground hover:text-destructive hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    :title="t('common.delete')" @click.stop="$emit('delete', category.id)">
                    <TrashIcon class="size-3" />
                </button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ICategory } from '@/stores/site'
import { useImageUrl } from '@/composables/useImageUrl'
import {
    Bars3Icon,
    TrashIcon,
    PencilIcon,
} from '@heroicons/vue/24/outline'

const props = defineProps<{
    category: ICategory
    index: number
}>()

defineEmits(['edit', 'delete'])

const { t } = useI18n()
const siteStore = useSiteStore()
const { getImageUrl } = useImageUrl()

const coverUrl = computed(() => {
    const cover = props.category.cover
    if (!cover) return ''
    if (cover.startsWith('http') || cover.startsWith('data:')) return cover
    return getImageUrl(`${siteStore.site.appDir}${cover}`)
})

const postCount = computed(() => {
    return siteStore.posts.filter(p => p.published && (p.categories || []).includes(props.category.name)).length
})
</script>
