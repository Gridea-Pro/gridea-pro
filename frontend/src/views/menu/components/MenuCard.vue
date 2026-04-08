<template>
    <div class="mb-4">
        <!-- 主菜单卡片 -->
        <div
            class="group flex rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:border-primary/20 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
            @click="$emit('edit', menu, index)">
            <div class="flex items-center pl-4 handle cursor-move">
                <Bars3Icon class="size-3 text-muted-foreground" />
            </div>
            <div class="p-4 flex-1">
                <div class="text-sm font-medium text-foreground mb-2 flex items-center gap-1.5">
                    {{ menu.name }}
                    <span v-if="menu.children?.length"
                        class="text-[10px] text-muted-foreground bg-primary/10 px-1.5 py-0.5 rounded-full">
                        {{ menu.children.length }}
                    </span>
                </div>
                <div class="text-xs flex items-center gap-3">
                    <div
                        class="px-2 py-0.5 bg-primary/10 border border-primary/20 rounded-full text-[10px] text-primary/80 flex items-center">
                        {{ menu.openType }}
                        <ArrowTopRightOnSquareIcon v-if="menu.openType === 'External'" class="w-3 h-3 ml-1" />
                    </div>
                    <div class="text-muted-foreground truncate">
                        {{ menu.link || '—' }}
                    </div>
                </div>
            </div>
            <div class="flex items-center px-4 gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                <button
                    class="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    :title="t('siteMenu.addSubmenu')" @click.stop="$emit('add-child', index)">
                    <PlusIcon class="size-3" />
                </button>
                <button v-if="menu.children?.length"
                    class="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    @click.stop="expanded = !expanded">
                    <ChevronDownIcon class="size-3 transition-transform duration-200"
                        :class="{ 'rotate-180': expanded }" />
                </button>
                <button
                    class="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    :title="t('common.edit')" @click.stop="$emit('edit', menu, index)">
                    <PencilIcon class="size-3" />
                </button>
                <button
                    class="p-2 text-muted-foreground hover:text-destructive hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                    :title="t('common.delete')" @click.stop="$emit('delete', index)">
                    <TrashIcon class="size-3" />
                </button>
            </div>
        </div>

        <!-- 子菜单列表（可拖拽排序） -->
        <div v-if="menu.children?.length && expanded" class="ml-5 mt-1">
            <draggable
                v-model="menu.children"
                handle=".child-handle"
                item-key="id"
                :animation="150"
                @change="$emit('sort-children', index)">
                <template #item="{ element: child, index: childIdx }">
                    <div
                        class="group flex items-center rounded-lg cursor-pointer transition-all duration-150 bg-primary/2 border border-primary/10 hover:border-primary/20 hover:bg-primary/5 mb-1"
                        @click="$emit('edit-child', index, menu, childIdx, child)">
                        <div class="flex items-center pl-3 child-handle cursor-move" @click.stop>
                            <Bars3Icon class="size-3 text-muted-foreground/50" />
                        </div>
                        <div class="pl-2 pr-2 text-muted-foreground/30">
                            <div class="w-2 h-px bg-current"></div>
                        </div>
                        <div class="py-2.5 flex-1 min-w-0">
                            <div class="text-xs font-medium text-foreground">{{ child.name }}</div>
                            <div class="text-[10px] text-muted-foreground truncate">{{ child.link }}</div>
                        </div>
                        <div class="flex items-center px-3 gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                                class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                                :title="t('common.edit')"
                                @click.stop="$emit('edit-child', index, menu, childIdx, child)">
                                <PencilIcon class="size-3" />
                            </button>
                            <button
                                class="p-1.5 text-muted-foreground hover:text-destructive hover:bg-primary/10 rounded-lg transition-colors cursor-pointer"
                                :title="t('common.delete')"
                                @click.stop="$emit('delete-child', index, childIdx)">
                                <TrashIcon class="size-3" />
                            </button>
                        </div>
                    </div>
                </template>
            </draggable>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import type { IMenu } from '@/interfaces/menu'
import {
    Bars3Icon,
    ArrowTopRightOnSquareIcon,
    TrashIcon,
    PencilIcon,
    PlusIcon,
    ChevronDownIcon,
} from '@heroicons/vue/24/outline'

defineProps<{
    menu: IMenu
    index: number
}>()

defineEmits<{
    (e: 'edit', menu: IMenu, index: number): void
    (e: 'delete', index: number): void
    (e: 'add-child', parentIndex: number): void
    (e: 'edit-child', parentIndex: number, parent: IMenu, childIndex: number, child: IMenu): void
    (e: 'delete-child', parentIndex: number, childIndex: number): void
    (e: 'sort-children', parentIndex: number): void
}>()

const { t } = useI18n()
const expanded = ref(true)
</script>
