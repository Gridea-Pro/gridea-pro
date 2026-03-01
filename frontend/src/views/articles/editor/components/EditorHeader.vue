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
                class="rounded-full text-primary hover:bg-primary/10 hover:text-primary h-8 w-12 p-0"
                :title="$t('article.publish')" @click="$emit('publish')">
                <PaperAirplaneIcon class="size-3 -rotate-45" />
            </Button>
        </div>
    </div>

    <!-- Right Tools -->
    <div class="right-tool-container">
        <!-- Info Popover -->
        <Popover>
            <PopoverTrigger as-child>
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:text-foreground hover:bg-primary/10 h-8 w-8 p-0">
                    <InformationCircleIcon class="size-4" />
                </Button>
            </PopoverTrigger>
            <PopoverContent side="left" align="start" class="w-48 p-4 bg-primary/10 transition-colors duration-200">
                <div class="post-stats">
                    <div class="item">
                        <h4>{{ $t('article.words') }}</h4>
                        <div class="number text-foreground">{{ articleStats.wordsNumber }}</div>
                    </div>
                    <div class="item">
                        <h4>{{ $t('article.readingTime') }}</h4>
                        <div class="number text-foreground">{{ articleStats.formatTime }}</div>
                    </div>
                </div>
            </PopoverContent>
        </Popover>

        <!-- Emoji Popover -->
        <Popover>
            <PopoverTrigger as-child>
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0">
                    <FaceSmileIcon class="size-4" />
                </Button>
            </PopoverTrigger>
            <PopoverContent side="left" align="start" class="w-[320px] p-0 overflow-hidden" :side-offset="10">
                <EmojiCard @select="$emit('emojiSelect', $event)" />
            </PopoverContent>
        </Popover>

        <Button
variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
            :title="$t('article.insertImage')" @click="$emit('insertImage')">
            <PhotoIcon class="size-4" />
        </Button>

        <Button
variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
            :title="$t('article.insertMore')" @click="$emit('insertMore')">
            <EllipsisHorizontalIcon class="size-4" />
        </Button>

        <Button
variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
            :title="$t('article.settings')" @click="$emit('openSettings')">
            <Cog6ToothIcon class="size-4" />
        </Button>

        <Button
variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
            :title="`${$t('nav.preview')} [Ctrl + P]`" @click="$emit('preview')">
            <EyeIcon class="size-4" />
        </Button>
    </div>

    <!-- Right Bottom Tool -->
    <div class="right-bottom-tool-container">
        <Popover>
            <PopoverTrigger as-child>
                <Button
variant="ghost" size="sm"
                    class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0">
                    <i class="ri-keyboard-line"></i>
                </Button>
            </PopoverTrigger>
            <PopoverContent side="left" align="end" class="w-64 p-4 max-h-[400px] overflow-y-auto">
                <div class="keyboard-tip mb-2">
                    💁‍♂️ 编辑区域右键能弹出快捷菜单哦
                </div>
                <div class="keyboard-container w-full">
                    <div v-for="(item, index) in shortcutKeys" :key="index" class="item">
                        <div class="keyboard-group-title text-xs font-bold text-muted-foreground my-2 border-b pb-1">{{
                            item.name }}</div>
                        <div class="list">
                            <div v-for="(listItem, listIndex) in item.list" :key="listIndex" class="list-item">
                                <div class="list-item-title text-foreground">{{ listItem.title }}</div>
                                <div class="text-muted-foreground">
                                    <span v-for="(keyCode, keyIndex) in listItem.keyboard" :key="keyIndex">
                                        <code>{{ keyCode }}</code> <span
                                            v-if="keyIndex !== listItem.keyboard.length - 1"> + </span>
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </PopoverContent>
        </Popover>
    </div>
</template>

<script lang="ts" setup>
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import EmojiCard from '@/components/EmojiCard/index.vue'
import shortcutKeys from '@/helpers/shortcut-keys'
import {
    ArrowLeftIcon,
    CheckIcon,
    InformationCircleIcon,
    FaceSmileIcon,
    PhotoIcon,
    Cog6ToothIcon,
    EyeIcon,
    PaperAirplaneIcon,
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

.right-tool-container,
.right-bottom-tool-container {
    position: fixed;
    right: 12px;
    display: flex;
    flex-direction: column;
    color: var(--muted-foreground);
    transition: color 0.3s ease;
    transition: opacity 700ms ease;
    z-index: 45;
    pointer-events: none;

    &:hover {
        color: var(--foreground);
    }
}

.right-tool-container :deep(button),
.right-tool-container :deep([role="button"]),
.right-bottom-tool-container :deep(button),
.right-bottom-tool-container :deep([role="button"]) {
    pointer-events: auto;
}

.right-tool-container {
    bottom: 50%;
    transform: translateY(50%);
}

.right-bottom-tool-container {
    bottom: 2px;
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
    width: 200px;

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

.keyboard-tip {
    font-size: 12px;
    color: var(--muted-foreground);
}
</style>
