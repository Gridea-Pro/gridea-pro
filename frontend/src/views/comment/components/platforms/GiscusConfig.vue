<template>
  <div>
    <Label class="mb-2 block">Repository <span class="text-red-500">*</span></Label>
    <Input v-model="config.repo" placeholder="owner/repo" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Repository ID <span class="text-red-500">*</span></Label>
    <Input v-model="config.repoId" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Category <span class="text-red-500">*</span></Label>
    <Input v-model="config.category" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Category ID <span class="text-red-500">*</span></Label>
    <Input v-model="config.categoryId" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const config = defineModel<any>('config', { required: true })
</script>
