<template>
  <div>
    <Label class="mb-2 block">Environment ID <span class="text-red-500">*</span></Label>
    <Input v-model="config.envId" />
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const config = defineModel<any>('config', { required: true })
</script>
