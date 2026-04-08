<script setup lang="ts">
import { DialogTitle, type DialogTitleProps } from 'radix-vue'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <DialogTitle
    v-bind="props"
    :class="cn('text-base font-semibold text-foreground', props.class)"
  >
    <slot />
  </DialogTitle>
</template>
