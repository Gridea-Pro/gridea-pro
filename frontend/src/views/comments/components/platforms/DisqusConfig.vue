<template>
  <div>
    <Label class="mb-2 block">Shortname <span class="text-red-500">*</span></Label>
    <Input v-model="config.shortname" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">API <span class="text-red-500">*</span></Label>
    <Input v-model="config.api" placeholder="https://disqus.skk.moe/disqus" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">API Key <span class="text-red-500">*</span></Label>
    <Input v-model="config.apiKey" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const config = defineModel<any>('config', { required: true })
</script>
