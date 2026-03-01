<template>
  <div>
    <Label class="mb-2 block">{{ t('appId') }} <span class="text-red-500">*</span></Label>
    <Input v-model="config.appId" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">{{ t('appKey') }} <span class="text-red-500">*</span></Label>
    <Input v-model="config.appKey" type="password" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">{{ t('masterKey') }}</Label>
    <Input v-model="config.masterKey" type="password" />
    <div class="ml-1 mt-2 text-xs text-muted-foreground/60 flex items-center gap-1 cursor-pointer">
      {{ t('masterKeyTip') }}
    </div>
  </div>

  <div class="mt-4">
    <Label class="mb-2 block">{{ t('serverURLs') }} <span class="text-red-500">*</span></Label>
    <Input v-model="config.serverURLs" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useI18n } from 'vue-i18n'

const config = defineModel<any>('config', { required: true })
const { t } = useI18n()
</script>
