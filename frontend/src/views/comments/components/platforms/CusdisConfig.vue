<template>
  <div>
    <Label class="mb-2 block">App ID <span class="text-red-500">*</span></Label>
    <Input v-model="config.appId" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Host <span class="text-red-500">*</span></Label>
    <Input v-model="config.host" placeholder="https://cusdis.com" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const config = defineModel<any>('config', { required: true })
</script>
