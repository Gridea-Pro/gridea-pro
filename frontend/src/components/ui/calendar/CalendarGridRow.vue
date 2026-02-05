<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { CalendarGridRow, type CalendarGridRowProps, useForwardProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarGridRowProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props

  return delegated
})

const forwarded = useForwardProps(delegatedProps)
</script>

<template>
  <CalendarGridRow :class="cn('flex', props.class)" v-bind="forwarded">
    <slot />
  </CalendarGridRow>
</template>
