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
                <!-- 标题/元信息通过 #header 插槽渲染到编辑器工具栏「下方」的滚动区，
                     使工具栏固定在顶部、其余随内容滚动（贴合 editor-vue 母版布局）。 -->
                <TiptapEditor ref="tiptapEditor" v-model:value="form.content" :is-post-page="true"
                    :placeholder="$t('article.editorPlaceholder')"
                    class="post-editor" @focus="handleEditorFocus"
                    @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)">
                    <template #header>
                        <input ref="titleInputRef" v-model="form.title"
                            class="post-title py-4 border-none pt-10 pb-2 bg-transparent text-xl focus:outline-none focus:ring-0 text-foreground placeholder:text-muted-foreground/50 font-bold"
                            :placeholder="$t('article.title')" @change="handleTitleChange" @focus="handleTitleFocus"
                            @keydown="(e: KeyboardEvent) => handleInputKeydown(e, form.content)" />

                        <div class="post-meta">
                            <span class="meta-item">
                                <CalendarIcon class="meta-icon" />
                                {{ form.createdAt.isValid() ? form.createdAt.format('YYYY-MM-DD') : '' }}
                            </span>
                            <span v-if="form.category" class="meta-item">
                                <FolderIcon class="meta-icon" />
                                {{ form.category }}
                            </span>
                            <span v-if="form.tags.length" class="meta-item">
                                <TagIcon class="meta-icon" />
                                {{ form.tags.join(', ') }}
                            </span>
                            <span class="meta-item meta-status" :class="form.published ? 'text-emerald-500' : 'text-amber-500'">
                                {{ form.published ? $t('article.published') : $t('article.draft') }}
                            </span>
                        </div>
                    </template>
                </TiptapEditor>
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
                :date-time-value="dateTimeValue" :feature-display-value="featureDisplayValue"
                :feature-image-preview-src="featureImagePreviewSrc" :is-generating-slug="isGeneratingSlug"
                @update:tag-input="tagInput = $event"
                @update:date-value="dateValue = $event" @update:date-time-value="dateTimeValue = $event"
                @update:feature-display-value="featureDisplayValue = $event" @add-tag="addTag" @remove-tag="removeTag"
                @select-tag="selectTag" @file-name-change="handleFileNameChange"
                @select-feature-image="selectFeatureImage" @clear-feature-image="clearFeatureImage"
                @confirm-publish="handleConfirmPublish" @generate-slug="handleGenerateSlug" />

            <!-- Unsaved Dialog -->
            <UnsavedDialog v-model:open="showUnsavedDialog" @confirm-close="confirmClose" />

            <span class="save-tip">{{ articleStatusTip }}</span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useSiteStore } from '@/stores/site'
import TiptapEditor from '@/components/editor/index.vue'
import { useI18n } from 'vue-i18n'
import { toast } from '@/helpers/toast'
import { GenerateSlug } from '@/wailsjs/go/facade/AIFacade'

import { CalendarIcon, FolderIcon, TagIcon } from '@heroicons/vue/24/outline'

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

const { t } = useI18n()
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
    dateTimeValue,
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
const isGeneratingSlug = ref(false)

const handleGenerateSlug = async () => {
    if (!form.title.trim()) {
        toast.warning(t('settings.ai.noTitle'))
        return
    }
    isGeneratingSlug.value = true
    try {
        const slug = await GenerateSlug(form.title)
        form.fileName = slug
        toast.success(t('settings.ai.generateSuccess'))
    } catch (e: any) {
        console.error('Generate slug failed:', e)
        const msg = String(e?.message || e || '')
        let toastMsg = t('settings.ai.generateFailed')
        if (msg.includes('[DAILY_LIMIT]')) {
            toastMsg = t('settings.ai.dailyLimitReached')
        } else if (msg.includes('[RATE_LIMIT]')) {
            toastMsg = t('settings.ai.rateLimited')
        } else if (msg.includes('[UPSTREAM_429]') || msg.includes('429')) {
            toastMsg = t('settings.ai.upstream429')
        }
        toast.error(toastMsg)
    } finally {
        isGeneratingSlug.value = false
    }
}

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
    tiptapEditor,
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
    // 当点击标题时，确保编辑器失去焦点
    const ed = tiptapEditor.value?.editor as any
    try {
        ed?.commands?.blur?.()
    } catch (e) {
        console.warn('[Focus] Failed to blur editor', e)
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
    /* 关键：flex 列子项默认 min-height:auto 会被内容撑高，导致编辑器无法限定高度、正文无法滚动 */
    min-height: 0;

    .post-title {
        width: 740px;
        margin: 0 auto;
        display: block;
    }

    .post-meta {
        width: 740px;
        margin: 0 auto;
        display: flex;
        align-items: center;
        gap: 16px;
        padding-bottom: 16px;
        font-size: 12px;
        color: var(--muted-foreground);
        flex-wrap: wrap;

        .meta-item {
            display: inline-flex;
            align-items: center;
            gap: 4px;
        }

        .meta-icon {
            width: 14px;
            height: 14px;
            flex-shrink: 0;
            opacity: 0.7;
        }

        .meta-status {
            font-weight: 500;
        }
    }

    .post-editor {
        flex: 1;
        min-height: 0;

        /* 限宽/居中/分栏排版统一在编辑器母版自身的 editor.css 里按 mode 处理
           （那里是非 scoped 全局样式，能稳定命中 TipTap 动态生成的 .ProseMirror，
            不受跨组件 :deep + data-v 匹配的影响）。此处只负责把编辑器撑满高度。 */
        :deep(.gridea-tiptap) {
            height: 100%;
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
