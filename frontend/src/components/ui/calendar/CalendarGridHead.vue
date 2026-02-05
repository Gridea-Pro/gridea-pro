<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { CalendarGridHead, type CalendarGridHeadProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarGridHeadProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <CalendarGridHead :class="cn(props.class)" v-bind="props">
    <slot />
  </CalendarGridHead>
</template>
