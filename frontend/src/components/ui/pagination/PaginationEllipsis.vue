<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import { MoreHorizontal } from 'lucide-vue-next'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <span
    aria-hidden
    :class="cn('flex h-8 w-8 items-center justify-center', props.class)"
  >
    <MoreHorizontal class="h-3 w-3" />
    <span class="sr-only">More pages</span>
  </span>
</template>
