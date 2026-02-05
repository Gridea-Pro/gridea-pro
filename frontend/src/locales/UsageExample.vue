<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { setI18nLanguage, type LocaleType } from '@/locales'

const { t, locale } = useI18n()

const handleLanguageChange = async (newLang: string) => {
    // Call the helper to update the language globally, in localStorage, and HTML tag
    await setI18nLanguage(newLang as LocaleType)
}
</script>

<template>
    <div>
        <p>{{ t('some.key') }}</p>
        <p>Current Locale: {{ locale }}</p>

        <button @click="handleLanguageChange('zh-CN')">中文</button>
        <button @click="handleLanguageChange('en')">English</button>
    </div>
</template>
