<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { CalendarCellTrigger, type CalendarCellTriggerProps, useForwardProps } from 'radix-vue'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarCellTriggerProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props

  return delegated
})

const forwarded = useForwardProps(delegatedProps)
</script>

<template>
  <CalendarCellTrigger
    :class="cn(
      buttonVariants({ variant: 'ghost' }),
      'h-10 w-10 p-0 font-normal hover:bg-primary/20 hover:text-primary cursor-pointer rounded-md',
      '[&[data-today]:not([data-selected])]:bg-primary/20 [&[data-today]:not([data-selected])]:text-primary',
      // Selected
      'data-[selected]:bg-primary data-[selected]:text-primary-foreground data-[selected]:opacity-100 data-[selected]:hover:bg-primary data-[selected]:hover:text-primary-foreground data-[selected]:focus:bg-primary data-[selected]:focus:text-primary-foreground',
      // Disabled
      'data-[disabled]:text-muted-foreground data-[disabled]:opacity-50',
      // Outside View
      'data-[outside-view]:text-muted-foreground data-[outside-view]:opacity-50 [&[data-outside-view][data-selected]]:bg-primary/50 [&[data-outside-view][data-selected]]:text-primary [&[data-outside-view][data-selected]]:opacity-30',
      props.class,
    )"
    v-bind="forwarded"
  >
    <slot />
  </CalendarCellTrigger>
</template>
