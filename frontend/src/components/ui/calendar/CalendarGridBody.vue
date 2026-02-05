<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { CalendarGridBody, type CalendarGridBodyProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarGridBodyProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <CalendarGridBody :class="cn(props.class)" v-bind="props">
    <slot />
  </CalendarGridBody>
</template>
