<template>
    <div v-if="visible" class="article-update-page" :class="{ 'is-entering': entering }"
        @mousemove="handlePageMousemove">
        <!-- Header & Tools -->
        <EditorHeader :can-submit="canSubmit" :article-stats="articleStats" @close="close" @save-draft="saveDraft"
            @publish="publishPost" @emoji-select="handleEmojiSelect" @insert-image="insertImage"
            @insert-more="insertMore" @open-settings="handleArticleSettingClick" @preview="previewPost(form.content)" />

        <!-- Content -->
        <div class="page-content">
            <div class="editor-wrapper">
                <input ref="titleInputRef" v-model="form.title"
                    class="post-title py-4 border-none pt-10 pb-10 bg-transparent text-xl focus:outline-none focus:ring-0 text-foreground placeholder:text-muted-foreground/50 font-bold"
                    :placeholder="$t('article.title')" @change="handleTitleChange" @focus="handleTitleFocus"
                    @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)" />

                <monaco-markdown-editor ref="monacoMarkdownEditor" v-model:value="form.content" :is-post-page="true"
                    class="post-editor" @focus="handleEditorFocus"
                    @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)"></monaco-markdown-editor>
            </div>

            <div class="footer-info">
                {{ $t('article.writingIn') }} <a class="link hover:text-primary cursor-pointer"
                    @click.prevent="openPage('https://gridea.pro')">Gridea Pro</a>
            </div>

            <!-- Preview Sheet -->
            <PreviewDialog ref="previewDialogRef" v-model:open="previewVisible" :title="form.title"
                :date-formatted="form.createdAt.format(siteStore.site.themeConfig.dateFormat)" :tags="form.tags"
                :html-content="previewHtml" />

            <!-- Settings Drawer -->
            <ArticleSettingsDrawer v-model:open="articleSettingsVisible" :form="form" :tag-input="tagInput"
                :available-tags="availableTags" :available-categories="availableCategories" :date-value="dateValue"
                :time-value="timeValue" :feature-display-value="featureDisplayValue"
                :feature-image-preview-src="featureImagePreviewSrc" @update:tag-input="tagInput = $event"
                @update:date-value="dateValue = $event" @update:time-value="timeValue = $event"
                @update:feature-display-value="featureDisplayValue = $event" @add-tag="addTag" @remove-tag="removeTag"
                @select-tag="selectTag" @file-name-change="handleFileNameChange"
                @select-feature-image="selectFeatureImage" @clear-feature-image="clearFeatureImage"
                @confirm-publish="handleConfirmPublish" />

            <!-- Unsaved Dialog -->
            <UnsavedDialog v-model:open="showUnsavedDialog" @confirm-close="confirmClose" />

            <span class="save-tip">{{ articleStatusTip }}</span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useSiteStore } from '@/stores/site'
import MonacoMarkdownEditor from '@/components/MonacoMarkdownEditor/index.vue'

import EditorHeader from './components/EditorHeader.vue'
import ArticleSettingsDrawer from './components/ArticleSettingsDrawer.vue'
import PreviewDialog from './components/PreviewDialog.vue'
import UnsavedDialog from './components/UnsavedDialog.vue'

import { useArticleForm } from './composables/useArticleForm'
import { useArticleActions } from './composables/useArticleActions'
import { useEditorHelper } from './composables/useEditorHelper'

const props = defineProps<{
    visible: boolean
    articleFileName: string
}>()

const emit = defineEmits<{
    close: []
    fetchData: []
}>()

const siteStore = useSiteStore()

// ── Composables ─────────────────────────────────────────

const {
    form,
    tagInput,
    changedAfterLastSave,
    articleStatusTip,
    canSubmit,
    articleStats,
    availableTags,
    availableCategories,
    dateValue,
    timeValue,
    featureDisplayValue,
    featureImagePreviewSrc,
    selectFeatureImage,
    clearFeatureImage,
    addTag,
    removeTag,
    selectTag,
    buildCurrentForm,
    handleTitleChange,
    handleFileNameChange,
    formatForm,
    updateArticleSavedStatus,
} = useArticleForm(() => props.articleFileName)

const articleSettingsVisible = ref(false)
const showUnsavedDialog = ref(false)

const {
    saveDraft,
    publishPost,
    handleConfirmPublish,
    handleArticleSettingClick,
    setupEvents,
    cleanupEvents,
} = useArticleActions({
    form,
    canSubmit,
    changedAfterLastSave,
    articleSettingsVisible,
    formatForm,
    updateArticleSavedStatus,
    onClose: () => emit('close'),
    onFetchData: () => emit('fetchData'),
})

const {
    monacoMarkdownEditor,
    previewVisible,
    entering,
    insertImage,
    insertMore,
    handleEmojiSelect,
    previewPost,
    handleInputKeydown,
    handlePageMousemove,
    openPage,
    previewHtml,
} = useEditorHelper()

const titleInputRef = ref<HTMLInputElement | null>(null)

// 焦点互斥逻辑：修复标题和正文同时出现光标
const handleTitleFocus = () => {
    // 当点击标题时，确保 monaco 失去焦点
    const editor = monacoMarkdownEditor.value?.editor
    if (editor && (editor as any)._focusTracker) {
        // 让编辑器失去焦点，使用更安全的方式
        try {
            (editor as any)._focusTracker.onBlur()
        } catch (e) {
            console.warn('[Focus] Failed to blur monaco editor', e)
        }
    }
}

const handleEditorFocus = () => {
    // 当编辑器获得焦点时，确保标题 input 失去焦点
    titleInputRef.value?.blur()
}

const previewDialogRef = ref<InstanceType<typeof PreviewDialog> | null>(null)

// 调试内容流转
watch(() => form.content, (newVal) => {
    console.log('[Editor View] form.content changed, length:', newVal?.length)
}, { immediate: true })

// ── 关闭逻辑 ────────────────────────────────────────────

const close = () => {
    if (changedAfterLastSave.value) {
        showUnsavedDialog.value = true
        return
    }
    emit('close')
}

const confirmClose = () => {
    showUnsavedDialog.value = false
    emit('close')
}

// ── 生命周期 ────────────────────────────────────────────

onMounted(() => {
    buildCurrentForm()
    setupEvents()
})

onUnmounted(() => {
    cleanupEvents()
})
</script>

<style lang="less" scoped>
.article-update-page {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 40;
    background: var(--background);
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .page-content {
        background: var(--background);
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    &.is-entering {

        :deep(.page-title),
        :deep(.right-tool-container),
        :deep(.right-bottom-tool-container) {
            opacity: 0;
        }
    }
}

.footer-info {
    text-align: center;
    color: var(--muted-foreground);
    font-size: 12px;
    font-weight: lighter;
    -webkit-font-smoothing: antialiased;
    padding: 4px 0;
    border-top: 1px solid var(--border);
    flex-shrink: 0;

    .link {
        color: var(--muted-foreground);

        &:hover {
            color: var(--foreground);
        }
    }
}

.editor-wrapper {
    width: 100%;
    margin: 0 auto;
    position: relative;
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow-y: hidden;
    overflow-x: hidden;

    .post-title {
        width: 728px;
        margin: 0 auto;
        display: block;
    }

    .post-editor {
        flex: 1;

        :deep(.monaco-markdown-editor) {
            width: 728px;
        }

        :deep(.monaco-editor),
        :deep(.monaco-editor-background) {
            background-color: transparent !important;
        }

        :deep(.monaco-editor .inputarea.ime-input) {
            z-index: 100 !important;
        }

        :deep(.monaco-editor .view-lines) {
            user-select: none !important;
        }
    }
}

.save-tip {
    padding: 4px 10px;
    line-height: 22px;
    font-size: 12px;
    color: var(--muted-foreground);
    position: fixed;
    left: 0;
    bottom: 0;
}
</style>
