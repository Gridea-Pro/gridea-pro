<script setup lang="ts">
import { PaginationPrev, type PaginationPrevProps } from 'radix-vue'
import { ChevronLeft } from 'lucide-vue-next'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

const props = withDefaults(defineProps<PaginationPrevProps & { class?: HTMLAttributes['class'] }>(), {
  asChild: true,
})
</script>

<template>
  <PaginationPrev v-bind="props">
    <Button :class="cn('h-8 w-8 p-0 cursor-pointer hover:bg-primary/10', props.class)" variant="ghost" size="icon">
      <slot>
        <ChevronLeft class="h-3 w-3" />
      </slot>
    </Button>
  </PaginationPrev>
</template>
