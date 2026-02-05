<script setup lang="ts">
import { PaginationRoot, type PaginationRootEmits, type PaginationRootProps, useForwardPropsEmits } from 'radix-vue'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<PaginationRootProps & { class?: HTMLAttributes['class'] }>()
const emits = defineEmits<PaginationRootEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <PaginationRoot v-bind="forwarded" :class="cn('mx-auto flex w-full justify-center', props.class)">
    <slot />
  </PaginationRoot>
</template>
