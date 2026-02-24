<template>
    <div
        class="memo-input-wrapper bg-card/50 border border-border/50 rounded-xl transition-all duration-200 ring-offset-background focus-within:ring-1 focus-within:ring-primary/10 relative overflow-visible">
        <div class="px-6 py-6">
            <textarea ref="textareaRef" v-model="content" :placeholder="placeholderText"
                class="w-full bg-transparent border-none focus:ring-0 resize-none p-0 min-h-[80px] text-sm leading-5 tracking-wider text-foreground placeholder:text-muted-foreground outline-none"
                :rows="1" @input="handleInput" @keydown="handleKeydown" @click="handleInput" />

            <!-- Tag Suggestions Dropdown -->
            <div v-if="showTagSuggestions && filteredTags.length > 0"
                class="absolute z-500 bg-card text-popover-foreground border border-border rounded-md shadow-md min-w-[120px] max-h-[200px] overflow-y-auto"
                :style="suggestionStyle">
                <div v-for="(tag, index) in filteredTags" :key="tag.name"
                    class="px-3 py-1.5 text-xs cursor-pointer hover:bg-primary/10 hover:text-primary transition-colors flex items-center justify-between"
                    :class="{ 'bg-primary/10 text-primary text-xs': index === selectedTagIndex }"
                    @click="selectTag(tag.name)">
                    <span># {{ tag.name }}</span>
                    <span class="text-xs text-muted-foreground ml-2 opacity-50">{{ tag.count }}</span>
                </div>
            </div>
        </div>
        <div class="flex items-center justify-end px-4 pb-3 pt-2 border-t border-border/30">
            <div class="flex items-center gap-2">
                <Button v-if="isEditing" variant="outline" size="sm" @click="handleCancel"
                    class="h-7 px-4 text-xs justify-center rounded-full bg-primary/5 border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer">
                    {{ t('common.cancel') }}
                </Button>

                <Button variant="default" size="sm"
                    class="h-7 px-4 rounded-full text-[10px] font-medium transition-all shadow-sm hover:shadow-md"
                    :disabled="!canSubmit" @click="handleSubmit">
                    <PaperAirplaneIcon class="w-3 h-3 mr-1 mb-0.5 -rotate-45" />
                    {{ submitBtnText }}
                </Button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button/index'
import { useMemoStore } from '@/stores/memo'
import { PaperAirplaneIcon } from '@heroicons/vue/24/outline'

interface Props {
    placeholder?: string
    submitText?: string
    isEditing?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    placeholder: '',
    submitText: '',
    isEditing: false,
})

const { t } = useI18n()
const memoStore = useMemoStore()
const content = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const placeholderText = computed(() => props.placeholder || t('memo.inputPlaceholder'))
const submitBtnText = computed(() => props.submitText || t('memo.publish'))

const isMac = computed(() => navigator.platform.toUpperCase().indexOf('MAC') >= 0)

const canSubmit = computed(() => content.value.trim().length > 0)

function autoResize() {
    nextTick(() => {
        if (textareaRef.value) {
            textareaRef.value.style.height = 'auto'
            textareaRef.value.style.height = Math.min(textareaRef.value.scrollHeight, 200) + 'px'
        }
    })
}

// Tag Autocomplete
const showTagSuggestions = ref(false)
const suggestionStyle = ref({ top: '0px', left: '0px' })
const currentTagQuery = ref('')
const selectedTagIndex = ref(0)
const cursorPosition = ref(0)
const hashIndex = ref(-1)

const filteredTags = computed(() => {
    if (!currentTagQuery.value) return memoStore.tagStats
    return memoStore.tagStats.filter(tag =>
        tag.name.toLowerCase().includes(currentTagQuery.value.toLowerCase())
    )
})

function handleInput() {
    autoResize()
    checkTagTrigger()
}

// 简单的光标位置计算 (针对textarea)
// 这是一个简化的实现，为了更精确的效果通常需要创建一个隐藏的 div 来模拟
function getCaretCoordinates() {
    if (!textareaRef.value) return { top: 0, left: 0 }

    // 简单估算，实际生产中建议使用专门的库如 textarea-caret
    // 这里我们暂时固定显示在输入框上方或跟随光标的大致位置
    // 由于精确计算 textarea 光标位置比较复杂，我们先用一个简单策略：
    // 显示在 textarea 的顶部，水平位置稍微偏移
    // 或者，我们可以引入一个库，但为了保持无依赖，我们尝试简单定位

    return { top: '40px', left: '20px' }
}



function checkTagTrigger() {
    if (!textareaRef.value) return

    const cursor = textareaRef.value.selectionStart
    const text = content.value

    // 向前查找 #
    // 匹配模式： ... #tag
    // 1. 找到光标前的最后一个 #
    const lastHash = text.lastIndexOf('#', cursor - 1)

    if (lastHash === -1) {
        showTagSuggestions.value = false
        return
    }

    // 检查 # 和光标之间是否有空格（除了 # 后面的那个位置外，通常标签不包含空格）
    // 简单起见，如果 # 和光标之间有换行或空格，则视为结束标签输入
    const textBetween = text.slice(lastHash + 1, cursor)
    if (/\s/.test(textBetween)) {
        showTagSuggestions.value = false
        return
    }

    // 更新状态
    hashIndex.value = lastHash
    currentTagQuery.value = textBetween
    showTagSuggestions.value = true
    selectedTagIndex.value = 0

    // 计算位置 (这里简化处理，显示在输入框左下方，稍微偏移)
    // 理想情况下应该跟随光标
    // 为了简单且不引入大库，我们让它显示在 textarea 内部的左上角，或者跟随文字流
    // 我们暂时先硬编码位置，稍后如果用户需要精确跟随再优化
    // 实际上，为了用户体验，我们应该尽量让它跟随。
    // 这里尝试一种基于 text measurement 的简单方法

    // 临时方案：显示在输入区域下方
    suggestionStyle.value = {
        top: '60px', // 估算值
        left: '24px'
    }
}

function selectTag(tagName: string) {
    if (hashIndex.value === -1 || !textareaRef.value) return

    const before = content.value.slice(0, hashIndex.value)
    const after = content.value.slice(textareaRef.value.selectionStart)

    // 插入标签并加空格
    const newContent = `${before}#${tagName} ${after}`
    content.value = newContent

    showTagSuggestions.value = false

    nextTick(() => {
        if (textareaRef.value) {
            textareaRef.value.focus()
            // 移动光标到标签后
            const newCursorPos = hashIndex.value + tagName.length + 2 // +2 for # and space
            textareaRef.value.setSelectionRange(newCursorPos, newCursorPos)
            autoResize()
        }
    })
}

function handleKeydown(event: KeyboardEvent) {
    if (showTagSuggestions.value && filteredTags.value.length > 0) {
        if (event.key === 'ArrowDown') {
            event.preventDefault()
            selectedTagIndex.value = (selectedTagIndex.value + 1) % filteredTags.value.length
            return
        }
        if (event.key === 'ArrowUp') {
            event.preventDefault()
            selectedTagIndex.value = (selectedTagIndex.value - 1 + filteredTags.value.length) % filteredTags.value.length
            return
        }
        if (event.key === 'Enter' || event.key === 'Tab') {
            event.preventDefault()
            selectTag(filteredTags.value[selectedTagIndex.value].name)
            return
        }
        if (event.key === 'Escape') {
            showTagSuggestions.value = false
            return
        }
    }

    if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault()
        handleSubmit()
    }
}

// ... existing handleSubmit, handleCancel, etc. ...

function handleSubmit() {
    if (!canSubmit.value) return

    console.log('MemoInput emitting submit:', content.value.trim())
    emit('submit', content.value.trim())
    // Keep content if in editing mode, cleaner for parent to handle clear
    if (!props.isEditing) {
        content.value = ''
        showTagSuggestions.value = false // Reset suggestions
        nextTick(() => {
            if (textareaRef.value) {
                textareaRef.value.style.height = 'auto'
            }
        })
    }
}

function handleCancel() {
    emit('cancel')
    content.value = ''
    showTagSuggestions.value = false // Reset suggestions
    nextTick(() => {
        if (textareaRef.value) {
            textareaRef.value.style.height = 'auto'
        }
    })
}

const emit = defineEmits<{
    submit: [content: string]
    cancel: []
}>()

// Expose method to set content
const setContent = (text: string) => {
    content.value = text
    autoResize()
}

const clearContent = () => {
    content.value = ''
    showTagSuggestions.value = false
    nextTick(() => {
        if (textareaRef.value) {
            textareaRef.value.style.height = 'auto'
        }
    })
}

defineExpose({
    setContent,
    clearContent
})

onMounted(() => {
    autoResize()
})
</script>
