<script setup lang="ts">
import { type HTMLAttributes, computed } from 'vue'
import { PaginationListItem, type PaginationListItemProps } from 'radix-vue'
import { Button, type ButtonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface Props extends PaginationListItemProps {
  class?: HTMLAttributes['class']
  isActive?: boolean
  size?: ButtonVariants['size']
}

const props = withDefaults(defineProps<Props>(), {
  size: 'icon',
  isActive: false,
})
</script>

<template>
  <PaginationListItem v-bind="props" as-child>
    <Button
      :class="cn(
        'h-8 w-8 p-0 cursor-pointer text-xs',
        !isActive && 'hover:bg-primary/10 hover:text-primary',
        props.class,
      )"
      :variant="isActive ? 'default' : 'ghost'"
      :size="size"
    >
      <slot />
    </Button>
  </PaginationListItem>
</template>
