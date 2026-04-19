<template>
    <!-- Top Header Bar -->
    <div class="page-title">
        <div class="flex items-center justify-between gap-2">
            <!-- 左侧：窗口控制 + 返回按钮 -->
            <div class="flex items-center gap-1" style="--wails-draggable: no-drag">
                <WindowControls />
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
                    :title="$t('common.back')" @click="$emit('close')">
                    <ArrowLeftIcon class="size-3" />
                </Button>
            </div>
            
            <!-- 中间：拖拽区域 -->
            <div class="flex-1"></div>
            
            <!-- 右侧：操作按钮 -->
            <div class="flex items-center gap-2" style="--wails-draggable: no-drag">
                <!-- 插入/格式 下拉菜单 -->
                <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                        <Button variant="ghost" size="sm" class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0">
                            <PencilIcon class="size-3" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-48">
                        <DropdownMenuItem @click="$emit('editorAction', 'heading-2')">H2</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'bold')">加粗</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'italic')">斜体</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'strike')">删除线</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'inline-code')">行内代码</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'bullet-list')">无序列表</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'ordered-list')">有序列表</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'task-list')">任务列表</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'blockquote')">引用</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('editorAction', 'code-block')">代码块</DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
                
                <!-- Emoji 按钮 -->
                <Popover>
                    <PopoverTrigger as-child>
                        <Button
variant="ghost" size="sm"
                            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-12 p-0"
                            title="表情">
                            <FaceSmileIcon class="size-3" />
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent side="bottom" align="end" class="w-[320px] p-0 overflow-hidden" :side-offset="10">
                        <EmojiCard @select="$emit('emojiSelect', $event)" />
                    </PopoverContent>
                </Popover>
                
                <!-- 预览按钮 -->
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
                    :title="`${$t('nav.preview')} [Ctrl + P]`" @click="$emit('preview')">
                    <EyeIcon class="size-3" />
                </Button>
                
                <!-- 保存草稿 -->
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
                    :disabled="!canSubmit" :title="$t('article.saveDraft')" @click="$emit('saveDraft')">
                    <CheckIcon class="size-3" />
                </Button>
                
                <!-- 发布 -->
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-primary hover:bg-primary/10 hover:text-primary h-8 w-12 p-0"
                    :title="$t('article.publish')" @click="$emit('publish')">
                    <PaperAirplaneIcon class="size-3 -rotate-45" />
                </Button>
                
                <!-- 更多 下拉菜单 -->
                <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                        <Button variant="ghost" size="sm" class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0">
                            <EllipsisHorizontalIcon class="size-3" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-48">
                        <DropdownMenuItem @click="$emit('insertImage')">插入图片</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('insertMore')">插入摘要分隔</DropdownMenuItem>
                        <DropdownMenuItem @click="$emit('openSettings')">设置</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <div class="px-4 py-2">
                            <div class="text-xs font-semibold text-muted-foreground">统计</div>
                            <div class="flex justify-between mt-1">
                                <span class="text-sm">{{ $t('article.words') }}</span>
                                <span class="text-sm font-medium">{{ articleStats.wordsNumber }}</span>
                            </div>
                            <div class="flex justify-between mt-1">
                                <span class="text-sm">{{ $t('article.readingTime') }}</span>
                                <span class="text-sm font-medium">{{ articleStats.formatTime }}</span>
                            </div>
                        </div>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem @click="showShortcuts = true">快捷键帮助</DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </div>
        
        <!-- 快捷键帮助弹窗 -->
        <Dialog v-model="showShortcuts">
            <DialogContent class="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>快捷键</DialogTitle>
                </DialogHeader>
                <div class="keyboard-container">
                    <div v-for="(item, index) in shortcutKeys" :key="index" class="item">
                        <div class="keyboard-group-title text-xs font-bold text-muted-foreground my-2 border-b pb-1">{{ item.name }}</div>
                        <div class="list">
                            <div v-for="(listItem, listIndex) in item.list" :key="listIndex" class="list-item">
                                <div class="list-item-title text-foreground">{{ listItem.title }}</div>
                                <div class="text-muted-foreground">
                                    <span v-for="(keyCode, keyIndex) in listItem.keyboard" :key="keyIndex">
                                        <code>{{ keyCode }}</code> <span v-if="keyIndex !== listItem.keyboard.length - 1"> + </span>
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import EmojiCard from '@/components/EmojiCard/index.vue'
import WindowControls from '@/components/WindowControls/index.vue'
import { getShortcutKeys } from '@/helpers/shortcut-keys'
import {
    ArrowLeftIcon,
    CheckIcon,
    FaceSmileIcon,
    EyeIcon,
    PaperAirplaneIcon,
    PencilIcon
} from '@heroicons/vue/24/outline'
import { EllipsisHorizontalIcon } from '@heroicons/vue/24/solid'

defineProps<{
    canSubmit: string | false
    articleStats: { wordsNumber: number; formatTime: string }
}>()

defineEmits<{
    close: []
    saveDraft: []
    publish: []
    emojiSelect: [emoji: string]
    insertImage: []
    insertMore: []
    openSettings: []
    preview: []
    editorAction: [action: 'heading-2' | 'bold' | 'italic' | 'strike' | 'inline-code' | 'bullet-list' | 'ordered-list' | 'task-list' | 'blockquote' | 'code-block']
}>()

const { t } = useI18n()
const shortcutKeys = computed(() => getShortcutKeys(t))
const showShortcuts = ref(false)
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

.post-stats {
    display: flex;

    .item {
        width: 50%;
        min-width: 80px;

        h4 {
            color: var(--muted-foreground);
            font-size: 12px;
            font-weight: normal;
        }

        .number {
            font-size: 18px;
            font-family: 'Noto Serif';
        }
    }
}

.keyboard-container {
    width: 100%;

    .keyboard-group-title {
        margin: 8px 0;
        font-size: 12px;
    }

    .list {
        .list-item {
            display: flex;
            justify-content: space-between;
            font-size: 12px;
            padding: 4px;
            border-radius: 2px;

            &:not(:last-child) {
                border-bottom: 1px solid var(--secondary);
            }

            &:hover {
                background: var(--secondary);
                color: var(--primary);
            }

            code {
                padding: 0px 4px;
                border-radius: 2px;
                background: var(--secondary);
            }
        }
    }
}
</style>
