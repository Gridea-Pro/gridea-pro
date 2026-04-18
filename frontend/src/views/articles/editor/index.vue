<template>
  <div
    v-if="visible"
    class="article-update-page"
    :class="{ 'is-entering': entering }"
    @mousemove="handlePageMousemove"
  >
    <EditorHeader
      :can-submit="canSubmit"
      :article-stats="articleStats"
      @close="close"
      @save-draft="saveDraft"
      @publish="publishPost"
      @emoji-select="handleEmojiSelect"
      @insert-image="insertImage"
      @insert-more="insertMore"
      @open-settings="handleArticleSettingClick"
      @preview="previewPost"
      @editor-action="handleEditorAction"
    />

    <div class="page-content">
      <div class="workspace-shell">
        <section class="editor-column">
          <div class="editor-scroll">
            <div class="editor-wrapper">
              <input
                ref="titleInputRef"
                v-model="form.title"
                class="post-title"
                :placeholder="$t('article.title')"
                @change="handleTitleChange"
                @focus="handleTitleFocus"
                @keydown="(event: KeyboardEvent) => handleInputKeydown(event)"
              >

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

              <TiptapMarkdownEditor
                ref="tiptapMarkdownEditor"
                v-model:value="form.content"
                :placeholder="$t('article.editorPlaceholder')"
                class="post-editor"
                @focus="handleEditorFocus"
                @keydown="(event: KeyboardEvent) => handleInputKeydown(event)"
              />
            </div>
          </div>
        </section>
      </div>

      <div class="footer-info">
        {{ $t('article.writingIn') }}
        <a class="link hover:text-primary cursor-pointer" @click.prevent="openPage('https://gridea.pro')">Gridea Pro</a>
      </div>

      <PreviewDialog
        v-model:open="previewVisible"
        :title="form.title"
        :date-formatted="formattedDate"
        :tags="form.tags"
        :html-content="previewHtml"
      />

      <ArticleSettingsDrawer
        v-model:open="articleSettingsVisible"
        :form="form"
        :tag-input="tagInput"
        :available-tags="availableTags"
        :available-categories="availableCategories"
        :date-value="dateValue"
        :time-value="timeValue"
        :feature-display-value="featureDisplayValue"
        :feature-image-preview-src="featureImagePreviewSrc"
        :is-generating-slug="isGeneratingSlug"
        @update:file-name="form.fileName = $event"
        @update:tag-input="tagInput = $event"
        @update:date-value="dateValue = $event"
        @update:time-value="timeValue = $event"
        @update:feature-display-value="featureDisplayValue = $event"
        @update:hide-in-list="form.hideInList = $event"
        @update:is-top="form.isTop = $event"
        @add-tag="addTag"
        @remove-tag="removeTag"
        @select-tag="selectTag"
        @category-change="handleCategoryChange"
        @file-name-change="handleFileNameChange"
        @select-feature-image="selectFeatureImage"
        @clear-feature-image="clearFeatureImage"
        @confirm-publish="handleConfirmPublish"
        @generate-slug="handleGenerateSlug"
      />

      <UnsavedDialog v-model:open="showUnsavedDialog" @confirm-close="confirmClose" />

      <span class="save-tip">{{ articleStatusTip }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'

import { CalendarIcon, FolderIcon, TagIcon } from '@heroicons/vue/24/outline'
import { useI18n } from 'vue-i18n'

import TiptapMarkdownEditor from '@/components/TiptapMarkdownEditor/index.vue'
import { toast } from '@/helpers/toast'
import { useSiteStore } from '@/stores/site'
import { GenerateSlug } from '@/wailsjs/go/facade/AIFacade'

import ArticleSettingsDrawer from './components/ArticleSettingsDrawer.vue'
import EditorHeader from './components/EditorHeader.vue'
import PreviewDialog from './components/PreviewDialog.vue'
import UnsavedDialog from './components/UnsavedDialog.vue'
import { useArticleActions } from './composables/useArticleActions'
import { useArticleForm } from './composables/useArticleForm'
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
const { t } = useI18n()

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
const isGeneratingSlug = ref(false)
const titleInputRef = ref<HTMLInputElement | null>(null)

const formattedDate = computed(() => form.createdAt.format(siteStore.site.themeConfig.dateFormat))

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
  } catch (error: any) {
    console.error('Generate slug failed:', error)
    const message = String(error?.message || error || '')
    let toastMessage = t('settings.ai.generateFailed')

    if (message.includes('[DAILY_LIMIT]')) {
      toastMessage = t('settings.ai.dailyLimitReached')
    } else if (message.includes('[RATE_LIMIT]')) {
      toastMessage = t('settings.ai.rateLimited')
    } else if (message.includes('[UPSTREAM_429]') || message.includes('429')) {
      toastMessage = t('settings.ai.upstream429')
    }

    toast.error(toastMessage)
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
  tiptapMarkdownEditor,
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
} = useEditorHelper(() => form.content)

const handleTitleFocus = () => {
  const editor = tiptapMarkdownEditor.value?.getEditor() ?? null
  editor?.commands.blur()
}

const handleEditorFocus = () => {
  titleInputRef.value?.blur()
}

const handleEditorAction = (action: string) => {
  switch (action) {
    case 'heading-2':
      tiptapMarkdownEditor.value?.toggleHeading(2)
      break
    case 'bold':
      tiptapMarkdownEditor.value?.toggleBold()
      break
    case 'italic':
      tiptapMarkdownEditor.value?.toggleItalic()
      break
    case 'strike':
      tiptapMarkdownEditor.value?.toggleStrike()
      break
    case 'inline-code':
      tiptapMarkdownEditor.value?.toggleInlineCode()
      break
    case 'bullet-list':
      tiptapMarkdownEditor.value?.toggleBulletList()
      break
    case 'ordered-list':
      tiptapMarkdownEditor.value?.toggleOrderedList()
      break
    case 'task-list':
      tiptapMarkdownEditor.value?.toggleTaskList()
      break
    case 'blockquote':
      tiptapMarkdownEditor.value?.toggleBlockquote()
      break
    case 'code-block':
      tiptapMarkdownEditor.value?.toggleCodeBlock()
      break
  }
}

const handleCategoryChange = (id: string) => {
  if (id === '_none_') {
    form.category = ''
    form.categoryId = ''
    return
  }

  const matched = availableCategories.value.find((category) => category.id === id)
  form.category = matched?.name || id
  form.categoryId = id
}

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
  inset: 0;
  z-index: 40;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--background);

  .page-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background:
      radial-gradient(circle at top left, color-mix(in srgb, var(--primary) 5%, transparent), transparent 30%),
      var(--background);
  }

  &.is-entering {
    :deep(.page-title) {
      opacity: 0;
    }
  }
}

.workspace-shell {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.editor-column {
  flex: 1;
  min-width: 0;
  min-height: 0;
}

.editor-scroll {
  height: 100%;
  overflow: auto;
}

.editor-wrapper {
  max-width: 900px;
  margin: 0 auto;
  padding: 28px 28px 40px;
}

.post-title {
  width: 100%;
  border: none;
  outline: none;
  background: transparent;
  padding: 18px 0 10px;
  font-size: 2rem;
  line-height: 1.2;
  font-weight: 700;
  color: var(--foreground);
  font-family: "Noto Serif", "PingFang SC", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;

  &::placeholder {
    color: color-mix(in srgb, var(--muted-foreground) 72%, transparent);
  }
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 20px;
  font-size: 12px;
  color: var(--muted-foreground);
  flex-wrap: wrap;
}

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

.footer-info {
  text-align: center;
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: lighter;
  -webkit-font-smoothing: antialiased;
  padding: 6px 0;
  border-top: 1px solid var(--border);
  flex-shrink: 0;

  .link {
    color: var(--muted-foreground);

    &:hover {
      color: var(--foreground);
    }
  }
}

.save-tip {
  position: fixed;
  left: 0;
  bottom: 0;
  padding: 4px 10px;
  line-height: 22px;
  font-size: 12px;
  color: var(--muted-foreground);
}

@media (max-width: 768px) {
  .editor-wrapper {
    padding: 20px 16px 28px;
  }

  .post-title {
    font-size: 1.72rem;
  }
}
</style>
