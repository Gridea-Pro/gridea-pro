<template>
    <div
        class="group flex flex-col p-4 rounded-xl relative transition-all duration-200 bg-primary/2 border border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5">
        <div class="flex items-start gap-3 mb-2 pr-2">
            <div
                class="w-10 h-10 rounded-full flex-shrink-0 bg-primary/10 flex items-center justify-center overflow-hidden">
                <img
:src="link.avatar || generateAvatar(link.name || link.id)" :alt="link.name"
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
                    :title="t('nav.preview')" @click.stop="$emit('preview', link.url)">
                    <EyeIcon class="size-3" />
                </button>
                <button
                    class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors"
                    :title="t('common.edit')" @click.stop="$emit('edit', link, index)">
                    <PencilIcon class="size-3" />
                </button>
                <button
                    class="p-1.5 text-muted-foreground hover:text-destructive hover:bg-primary/10 rounded-md transition-colors"
                    :title="t('common.delete')" @click.stop="$emit('delete', link.id)">
                    <TrashIcon class="size-3" />
                </button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ILink } from '@/stores/site'
import { generateAvatar } from '@/utils/avatarGenerator'
import {
    TrashIcon,
    PencilIcon,
    Bars2Icon,
    EyeIcon,
} from '@heroicons/vue/24/outline'

defineProps<{
    link: ILink
    index: number
}>()

defineEmits(['preview', 'edit', 'delete'])

const { t } = useI18n()
</script>
