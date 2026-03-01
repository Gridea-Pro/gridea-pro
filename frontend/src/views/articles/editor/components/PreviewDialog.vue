<template>
    <Sheet v-model:open="openModel">
        <SheetContent side="right" class="w-screen max-w-4xl sm:max-w-4xl p-0">
            <div class="flex h-full flex-col overflow-y-scroll bg-background py-6 shadow-xl">
                <div class="px-4 sm:px-6">
                    <div class="flex items-start justify-between">
                        <SheetTitle class="text-lg font-medium text-foreground"></SheetTitle>
                    </div>
                </div>
                <div class="relative mt-6 flex-1 px-4 sm:px-6">
                    <h1 class="preview-title text-foreground">{{ title }}</h1>
                    <div class="preview-date">{{ dateFormatted }}</div>
                    <div class="preview-tags">
                        <span v-for="(tag, index) in tags" :key="index" class="tag">
                            {{ tag }}
                        </span>
                    </div>
                    <div ref="containerRef" class="preview-container"></div>
                </div>
            </div>
        </SheetContent>
    </Sheet>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'

const props = defineProps<{
    open: boolean
    title: string
    dateFormatted: string
    tags: string[]
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
}>()

const containerRef = ref<HTMLElement | null>(null)

const openModel = computed({
    get: () => props.open,
    set: (val: boolean) => emit('update:open', val),
})

defineExpose({ containerRef })
</script>

<style lang="less" scoped>
.preview-title {
    font-size: 24px;
    font-weight: bold;
    font-family: "Noto Serif", "PingFang SC", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;
}

.preview-date {
    font-size: 13px;
    color: var(--muted-foreground);
    margin-bottom: 16px;
}

.preview-tags {
    font-size: 12px;
    margin-bottom: 16px;

    .tag {
        display: inline-block;
        margin: 0 8px 8px 0;
        padding: 4px 8px;
        background: var(--secondary);
        color: var(--muted-foreground);
        border-radius: 20px;
    }
}

.preview-feature-image {
    max-width: 100%;
    margin-bottom: 16px;
    border-radius: 2px;
}

.preview-container {
    width: 100%;
    flex-shrink: 0;
    font-family: "Noto Serif", "PingFang SC", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;
    font-size: 15px;
    color: var(--foreground);

    :deep(a) {
        color: var(--foreground);
        word-wrap: break-word;
        text-decoration: none;
        border-bottom: 1px solid var(--border);

        &:hover {
            color: var(--primary);
            border-bottom: 1px solid var(--primary);
        }
    }

    :deep(img) {
        display: block;
        max-width: 100%;
        border-radius: 2px;
        margin: 24px auto;
    }

    :deep(p) {
        line-height: 1.62;
        margin-bottom: 1.12em;
        font-size: 15px;
        letter-spacing: .05em;
        hyphens: auto;
    }

    :deep(p),
    :deep(li) {
        line-height: 1.62;

        code {
            font-family: 'Source Code Pro', Consolas, Menlo, Monaco, 'Courier New', monospace;
            line-height: initial;
            word-wrap: break-word;
            border-radius: 0;
            background-color: var(--secondary);
            color: var(--primary);
            padding: .2em .33333333em;
            font-size: .875rem;
            margin-left: .125em;
            margin-right: .125em;
        }
    }

    :deep(pre) {
        background: var(--secondary);
        padding: 16px;
        border-radius: 2px;

        code {
            color: var(--foreground);
            font-family: 'Source Code Pro', Consolas, Menlo, Monaco, 'Courier New', monospace;
        }
    }

    :deep(blockquote) {
        color: var(--muted-foreground);
        position: relative;
        padding: .4em 0 0 2.2em;
        font-size: .96em;

        &:before {
            position: absolute;
            top: -4px;
            left: 0;
            content: "\201c";
            font: 700 62px/1 serif;
            color: var(--border);
        }
    }

    :deep(table) {
        border-collapse: collapse;
        margin: 1rem 0;
        width: 100%;

        tr {
            border-top: 1px solid var(--border);

            &:nth-child(2n) {
                background-color: var(--secondary);
            }
        }

        td,
        th {
            border: 1px solid var(--border);
            padding: .6em 1em;
        }
    }

    :deep(ul),
    :deep(ol) {
        padding-left: 35px;
        line-height: 1.62;
        margin-bottom: 16px;
    }

    :deep(ol) {
        list-style: decimal !important;
    }

    :deep(ul) {
        list-style-type: square !important;
    }

    :deep(h1),
    h2,
    h3,
    h4,
    h5,
    h6 {
        margin: 16px 0;
        font-weight: 700;
        padding-top: 16px;
    }

    :deep(h1) {
        font-size: 1.8em;
    }

    :deep(h2) {
        font-size: 1.42em;
    }

    :deep(h3) {
        font-size: 1.17em;
    }

    :deep(h4) {
        font-size: 1em;
    }

    :deep(h5) {
        font-size: 1em;
    }

    :deep(h6) {
        font-size: 1em;
        font-weight: 500;
    }

    :deep(hr) {
        display: block;
        border: 0;
        margin: 2.24em auto 2.86em;

        &:before {
            color: rgba(0, 0, 0, .2);
            font-size: 1.1em;
            display: block;
            content: "* * *";
            text-align: center;
        }
    }

    :deep(.footnotes) {
        margin-left: auto;
        margin-right: auto;
        max-width: 760px;
        padding-left: 18px;
        padding-right: 18px;

        &:before {
            content: "";
            display: block;
            border-top: 4px solid rgba(0, 0, 0, .1);
            width: 50%;
            max-width: 100px;
            margin: 40px 0 20px;
        }
    }

    :deep(.contains-task-list) {
        list-style-type: none;
        padding-left: 30px;
    }

    :deep(.task-list-item) {
        position: relative;
    }

    :deep(.task-list-item-checkbox) {
        position: absolute;
        cursor: pointer;
        width: 16px;
        height: 16px;
        margin: 4px 0 0;
        top: -1px;
        left: -22px;
        transform-origin: center;
        transform: rotate(-90deg);
        transition: all .2s ease;

        &:checked {
            transform: rotate(0);

            &:before {
                border: transparent;
                background-color: #9AE6B4;
            }

            &:after {
                transform: rotate(-45deg) scale(1);
            }

            +.task-list-item-label {
                color: #999;
                text-decoration: line-through;
            }
        }

        &:before {
            content: "";
            width: 16px;
            height: 16px;
            box-sizing: border-box;
            display: inline-block;
            border: 1px solid #9AE6B4;
            border-radius: 2px;
            background-color: #fff;
            position: absolute;
            top: 0;
            left: 0;
            transition: all .2s ease;
        }

        &:after {
            content: "";
            transform: rotate(-45deg) scale(0);
            width: 9px;
            height: 5px;
            border: 1px solid #22543D;
            border-top: none;
            border-right: none;
            position: absolute;
            display: inline-block;
            top: 4px;
            left: 4px;
            transition: all .2s ease;
        }
    }

    :deep(.markdownIt-TOC) {
        list-style: none;
        background: #f7fafc;
        padding: 1.5rem;
        border-radius: 0.5rem;
        color: #4a5568;
    }

    :deep(.markdownIt-TOC ul) {
        list-style: none;
        padding-left: 16px;
    }

    :deep(mark) {
        background: #FAF089;
        color: #744210;
    }
}
</style>
