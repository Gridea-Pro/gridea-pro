<script setup lang="ts">
import { DropdownMenuSeparator, type DropdownMenuSeparatorProps } from 'radix-vue'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<DropdownMenuSeparatorProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <DropdownMenuSeparator v-bind="props" :class="cn('-mx-1 my-1 h-px bg-muted', props.class)" />
</template>
