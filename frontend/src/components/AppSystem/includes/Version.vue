<template>
  <div class="mb-4 py-4 border-b border-gray-200">
    <div class="text-base font-medium mb-4">{{ t('settings.system.version') }}</div>
    <div>{{ version }}</div>
    <a href="#" @click.prevent="openWebsite"> Gridea Pro </a>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import pkg from '../../../../package.json'
import { BrowserOpenURL } from 'wailsjs/runtime'

const { t } = useI18n()
const version = pkg.version

const openWebsite = () => {
  BrowserOpenURL('https://gridea.dev')
}
</script>

<style lang="less" scoped></style>
