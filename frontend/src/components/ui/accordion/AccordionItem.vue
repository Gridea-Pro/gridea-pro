<script setup lang="ts">
import { AccordionItem, type AccordionItemProps } from 'radix-vue'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<AccordionItemProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <AccordionItem
    v-bind="props"
    :class="cn('border-b', props.class)"
  >
    <slot />
  </AccordionItem>
</template>
