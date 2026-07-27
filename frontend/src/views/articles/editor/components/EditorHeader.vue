<template>
    <!-- Top Header Bar -->
    <div ref="pageTitle" class="page-title">
        <div class="flex justify-end gap-2">
            <Button
variant="ghost" size="sm"
                class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
                :title="$t('common.back')" @click="$emit('close')">
                <ArrowLeftIcon class="size-3" />
            </Button>

            <Button
variant="ghost" size="sm"
                class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
                :disabled="!canSubmit" :title="$t('article.saveDraft')" @click="$emit('saveDraft')">
                <CheckIcon class="size-3" />
            </Button>

            <Button
variant="ghost" size="sm"
                class="rounded-full text-primary hover:bg-primary/10 hover:text-primary h-8 w-12 p-0 disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="!canSubmit" :title="$t('article.publish')" @click="$emit('publish')">
                <PaperAirplaneIcon class="size-3 -rotate-45" />
            </Button>
        </div>
    </div>

    <!-- 右侧浮动 icon 列已移除：表情/插图/摘要分隔均由编辑器工具栏与斜杠菜单提供；
         文章设置抽屉只在点「发布」时弹出（publishPost → 抽屉 → 确认发布）；
         预览保留 Ctrl+P 快捷键（PreviewDialog）。 -->

    <!-- 旧版（Monaco 时代）快捷键提示入口已移除：图标用的 ri-keyboard-line 字体类
         从未加载（不可见但可点击），且所列键位已随 TipTap 编辑器失效。
         现行入口为编辑器右下角键盘图标（ShortcutsPanel）。 -->
</template>

<script lang="ts" setup>
import { Button } from '@/components/ui/button'
import { ArrowLeftIcon, CheckIcon, PaperAirplaneIcon } from "@heroicons/vue/24/outline"

defineProps<{
    canSubmit: string | false
}>()

defineEmits<{
    close: []
    saveDraft: []
    publish: []
}>()
</script>

<style lang="less" scoped>
.page-title {
    padding: 8px 16px;
    z-index: 41;
    background: var(--background);
    transition: opacity 700ms ease;
    border-bottom: 1px solid var(--border);
    --wails-draggable: drag;
}

</style>
