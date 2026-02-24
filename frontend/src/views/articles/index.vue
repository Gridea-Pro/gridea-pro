<template>
    <ArticleList @new-article="newArticle" @edit-post="editPost" />

    <ArticleEditor v-if="editorVisible" :visible="editorVisible" :article-file-name="currentArticleFileName"
        @close="closeEditor" @fetch-data="reloadSite" />
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { EventsEmit } from '@/wailsjs/runtime'
import ArticleList from './list/index.vue'
import ArticleEditor from './editor/index.vue'
import type { IPost } from '@/interfaces/post'

const editorVisible = ref(false)
const currentArticleFileName = ref('')

const newArticle = () => {
    currentArticleFileName.value = ''
    editorVisible.value = true
}

const editPost = (post: IPost) => {
    currentArticleFileName.value = post.fileName
    editorVisible.value = true
}

const closeEditor = () => {
    editorVisible.value = false
    currentArticleFileName.value = ''
}

const reloadSite = () => {
    EventsEmit('app-site-reload')
}
</script>
