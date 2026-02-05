<script setup lang="ts">
import { PaginationNext, type PaginationNextProps } from 'radix-vue'
import { ChevronRight } from 'lucide-vue-next'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

const props = withDefaults(defineProps<PaginationNextProps & { class?: HTMLAttributes['class'] }>(), {
  asChild: true,
})
</script>

<template>
  <PaginationNext v-bind="props">
    <Button :class="cn('h-8 w-8 p-0 cursor-pointer hover:bg-primary/10', props.class)" variant="ghost" size="icon">
      <slot>
        <ChevronRight class="h-3 w-3" />
      </slot>
    </Button>
  </PaginationNext>
</template>
