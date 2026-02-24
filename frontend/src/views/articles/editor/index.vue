<template>
    <div class="article-update-page" :class="{ 'is-entering': entering }" v-if="visible"
        @mousemove="handlePageMousemove">
        <!-- Header & Tools -->
        <EditorHeader :can-submit="canSubmit" :article-stats="articleStats" @close="close" @save-draft="saveDraft"
            @publish="publishPost" @emoji-select="handleEmojiSelect" @insert-image="insertImage"
            @insert-more="insertMore" @open-settings="handleArticleSettingClick" @preview="previewPost(form.content)" />

        <!-- Content -->
        <div class="page-content">
            <div class="editor-wrapper">
                <input
                    class="post-title py-2 border-none mt-4 mb-4 bg-transparent text-2xl focus:outline-none focus:ring-0 text-foreground placeholder:text-muted-foreground font-bold"
                    :placeholder="$t('article.title')" v-model="form.title" @change="handleTitleChange"
                    @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)" />

                <monaco-markdown-editor ref="monacoMarkdownEditor" v-model:value="form.content"
                    @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)" :isPostPage="true"
                    class="post-editor"></monaco-markdown-editor>

                <div class="footer-info">
                    {{ $t('article.writingIn') }} <a @click.prevent="openPage('https://gridea.pro')"
                        class="link cursor-pointer">Gridea Pro</a>
                </div>
            </div>

            <!-- Preview Sheet -->
            <PreviewDialog v-model:open="previewVisible" :title="form.title"
                :date-formatted="form.date.format(siteStore.site.themeConfig.dateFormat)" :tags="form.tags"
                ref="previewDialogRef" />

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

            <!-- Hidden file input for image upload -->
            <input ref="uploadInputRef" class="upload-input hidden" type="file" accept="image/*"
                @change="fileChangeHandler">

            <span class="save-tip">{{ articleStatusTip }}</span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
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
    uploadInputRef,
    monacoMarkdownEditor,
    previewVisible,
    entering,
    insertImage,
    insertMore,
    handleEmojiSelect,
    fileChangeHandler,
    previewPost,
    handleInputKeydown,
    handlePageMousemove,
    openPage,
} = useEditorHelper()

const previewDialogRef = ref<InstanceType<typeof PreviewDialog> | null>(null)

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

    .page-content {
        background: var(--background);
        flex: 1;
        display: flex;
        overflow: scroll;
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
    padding-top: 6px;
    margin-top: 8px;
    border-top: 1px solid var(--border);

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

        :deep(.monaco-editor) {
            .scrollbar {
                position: fixed !important;
                top: 110px !important;
            }
        }

        :deep(.monaco-editor),
        :deep(.monaco-editor-background) {
            background-color: transparent !important;
        }

        :deep(.monaco-editor .selected-text) {
            background-color: var(--primary) !important;
            opacity: 0.2;
        }

        :deep(.monaco-editor .view-lines) {
            user-select: text;
        }

        :deep(.monaco-editor .cursor) {
            visibility: inherit !important;
        }

        :deep(.monaco-editor .inputarea.ime-input) {
            z-index: 100 !important;
        }

        :deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider) {
            background: rgba(121, 121, 121, 0.4) !important;

            &:hover {
                background: rgba(121, 121, 121, 0.6) !important;
            }
        }

        :deep(.monaco-editor .monaco-scrollable-element > .scrollbar.invisible.fade) {
            transition: opacity 0.3s cubic-bezier(0.3, 0.5, 0.5, 1);
        }
    }
}

.page-content {
    cursor: text;
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
