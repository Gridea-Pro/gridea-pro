<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { CalendarHeader, type CalendarHeaderProps, useForwardProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarHeaderProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props

  return delegated
})

const forwarded = useForwardProps(delegatedProps)
</script>

<template>
  <CalendarHeader :class="cn('relative flex items-center justify-between pt-1', props.class)" v-bind="forwarded">
    <slot />
  </CalendarHeader>
</template>
