<template>
    <div class="memo-list space-y-3">
        <template v-if="memos.length > 0">
            <MemoItem
v-for="memo in memos" :key="memo.id" :memo="memo" @update="$emit('update', $event)"
                @delete="$emit('delete', $event)" @tag-click="$emit('tagClick', $event)" />
        </template>
        <div v-else class="text-center py-12 text-muted-foreground">
            <LightBulbIcon class="size-10 mx-auto mb-3 opacity-30" />
            <p class="text-sm">{{ emptyText }}</p>
        </div>
    </div>
</template>

<script lang="ts" setup>
import type { IMemo } from '@/interfaces/memo'
import MemoItem from './MemoItem.vue'
import { LightBulbIcon } from '@heroicons/vue/24/outline'

interface Props {
    memos: IMemo[]
    emptyText?: string
}

withDefaults(defineProps<Props>(), {
    emptyText: '暂无闪念',
})

defineEmits<{
    update: [memo: IMemo]
    delete: [id: string]
    tagClick: [tag: string]
}>()
</script>
