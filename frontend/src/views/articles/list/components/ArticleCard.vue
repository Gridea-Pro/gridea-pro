<template>
    <div
class="group relative flex rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
        @click="$emit('edit', post)">
        <div class="p-3 flex-1 flex flex-col sm:flex-row">
            <!-- Checkbox & Info -->
            <div class="flex flex-1 min-w-0">
                <div class="flex flex-shrink-0 items-center p-3" @click.stop>
                    <Checkbox
:checked="selected" class="data-[state=checked]:bg-primary data-[state=checked]:border-primary"
                        @update:checked="$emit('select', post)" />
                </div>
                <div class="flex-1 h-14 min-w-0">
                    <div
                        class="text-[13px] font-medium mt-1.5 mb-2 text-foreground truncate pr-16 group-hover:text-primary transition-colors max-w-[600px]">
                        {{ post.title }}</div>
                    <div class="flex items-center text-xs text-muted-foreground gap-3 flex-wrap">
                        <div class="flex items-center text-[10px]">
                            <div
class="w-1.5 h-1.5 rounded-full mr-1.5"
                                :class="post.published ? 'bg-green-500' : 'bg-gray-300'"></div>
                            {{ post.published ? t('article.published') : t('article.draft') }}
                        </div>
                        <div class="w-px h-3 bg-primary/30"></div>
                        <div class="flex items-center text-[10px]">
                            <CalendarIcon class="size-3 mr-1 text-muted-foreground/70 translate-y-[-0.5px]" />
                            {{ dayjs(post.date).format('YYYY-MM-DD') }}
                        </div>
                        <template v-if="(post.categories || []).length > 0">
                            <div class="w-px h-3 bg-primary/30"></div>
                            <div class="flex items-center text-[10px] text-muted-foreground/70">
                                <FolderIcon class="size-3 mr-1" />
                                {{ (post.categories || [])[0] }}
                            </div>
                        </template>
                        <template v-if="(post.tags || []).length > 0">
                            <div class="w-px h-3 bg-primary/30"></div>
                            <div class="flex items-center flex-wrap gap-1 text-[10px]">
                                <span
v-for="(tag, index) in (post.tags || []).slice(0, 3)" :key="index"
                                    class="px-2 py-0.5 bg-primary/10 border border-primary/20 rounded-full text-[10px] text-primary/80">
                                    {{ tag }}
                                </span>
                                <span v-if="(post.tags || []).length > 3" class="text-[10px]">...</span>
                            </div>
                        </template>
                    </div>
                </div>
            </div>
            <!-- Actions -->
            <div class="absolute right-4 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-2 p-1 z-20">
                <div
class="p-1.5 hover:bg-primary/10 rounded-md cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    title="预览" @click.stop="$emit('preview', post)">
                    <EyeIcon class="size-3" />
                </div>
                <div
class="p-1.5 hover:bg-primary/10 rounded-md cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                    title="编辑" @click.stop="$emit('edit', post)">
                    <PencilIcon class="size-3" />
                </div>
                <div
class="p-1.5 hover:bg-destructive/10 hover:text-destructive rounded-md cursor-pointer text-muted-foreground transition-colors"
                    title="删除" @click.stop="$emit('delete', post)">
                    <TrashIcon class="size-3" />
                </div>
            </div>
        </div>

        <!-- Feature Image -->
        <div
v-if="post.feature"
            class="w-[100px] hidden sm:block relative overflow-hidden rounded-r-xl transition-opacity duration-200 group-hover:opacity-0">
            <img :src="featureUrl" class="absolute inset-0 w-full h-full object-cover" />
        </div>

        <!-- Status Badges -->
        <div class="absolute top-0 right-0 flex pointer-events-none">
            <div
v-if="post.hideInList"
                class="px-2 py-0.5 text-[10px] font-bold bg-foreground text-background rounded-bl-lg z-10 shadow-sm">
                HIDE
            </div>
            <div
v-if="post.isTop"
                class="px-2 py-0.5 text-[10px] font-bold bg-yellow-400 text-yellow-900 rounded-bl-lg ml-[-4px] z-10 shadow-sm">
                TOP
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { Checkbox } from '@/components/ui/checkbox'
import {
    CalendarIcon,
    FolderIcon,
    EyeIcon,
    PencilIcon,
    TrashIcon,
} from '@heroicons/vue/24/outline'
import type { IPost } from '@/interfaces/post'
import { useArticleImageUrl } from '../../shared/useImageUrl'

const props = defineProps<{
    post: IPost
    selected: boolean
}>()

defineEmits<{
    edit: [post: IPost]
    select: [post: IPost]
    preview: [post: IPost]
    delete: [post: IPost]
}>()

const { t } = useI18n()
const { getFeatureUrl } = useArticleImageUrl()

const featureUrl = computed(() => getFeatureUrl(props.post.feature))
</script>
 
