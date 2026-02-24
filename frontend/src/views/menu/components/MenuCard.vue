<template>
    <div class="group flex mb-4 rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
        @click="$emit('edit', menu, index)">
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
            <button class="p-2 text-muted-foreground hover:text-primary hover:bg-secondary rounded-lg transition-colors"
                @click.stop="$emit('edit', menu, index)" :title="t('common.edit')">
                <PencilIcon class="size-3" />
            </button>
            <button
                class="p-2 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-lg transition-colors"
                @click.stop="$emit('delete', index)" :title="t('common.delete')">
                <TrashIcon class="size-3" />
            </button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import type { IMenu } from '@/interfaces/menu'
import {
    Bars3Icon,
    ArrowTopRightOnSquareIcon,
    TrashIcon,
    PencilIcon,
} from '@heroicons/vue/24/outline'

defineProps<{
    menu: IMenu
    index: number
}>()

defineEmits<{
    (e: 'edit', menu: IMenu, index: number): void
    (e: 'delete', index: number): void
}>()

const { t } = useI18n()
</script>
