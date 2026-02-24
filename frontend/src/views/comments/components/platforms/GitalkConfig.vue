<template>
  <div>
    <Label class="mb-2 block">Client ID <span class="text-red-500">*</span></Label>
    <Input v-model="config.clientId" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Client Secret <span class="text-red-500">*</span></Label>
    <Input type="password" v-model="config.clientSecret" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Repository <span class="text-red-500">*</span></Label>
    <Input v-model="config.repo" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Owner <span class="text-red-500">*</span></Label>
    <Input v-model="config.owner" />
  </div>
  <div class="mt-4">
    <Label class="mb-2 block">Admin <span class="text-red-500">*</span></Label>
    <Input v-model="config.admin" placeholder="多位管理员用逗号分隔" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const config = defineModel<any>('config', { required: true })
</script>
