<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { CalendarCell, type CalendarCellProps, useForwardProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarCellProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props

  return delegated
})

const forwarded = useForwardProps(delegatedProps)
</script>

<template>
  <CalendarCell
    :class="cn('relative p-0 text-center text-sm focus-within:relative focus-within:z-20 [&:has([data-selected])]:bg-accent [&:has([data-selected][data-outside-view])]:bg-accent/50 [&:has([data-inside-view][data-selected])]:bg-accent', props.class)"
    v-bind="forwarded"
  >
    <slot />
  </CalendarCell>
</template>
