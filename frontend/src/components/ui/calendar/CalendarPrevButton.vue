<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { CalendarPrev, type CalendarPrevProps, useForwardProps } from 'radix-vue'
import { ChevronLeftIcon } from '@heroicons/vue/24/outline'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const props = defineProps<CalendarPrevProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props

  return delegated
})

const forwarded = useForwardProps(delegatedProps)
</script>

<template>
  <CalendarPrev
    :class="cn(
      buttonVariants({ variant: 'outline' }),
      'h-7 w-7 bg-transparent p-0 opacity-50 hover:bg-primary/20 hover:text-primary cursor-pointer',
      props.class,
    )"
    v-bind="forwarded"
  >
    <slot>
      <ChevronLeftIcon class="h-4 w-4" />
    </slot>
  </CalendarPrev>
</template>
