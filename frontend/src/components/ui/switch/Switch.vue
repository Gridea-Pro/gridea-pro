<script setup lang="ts">
import {
  SwitchRoot,
  type SwitchRootEmits,
  type SwitchRootProps,
  SwitchThumb,
  useForwardPropsEmits,
} from 'radix-vue'
import { type HTMLAttributes, computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(defineProps<SwitchRootProps & { class?: HTMLAttributes['class'], size?: 'default' | 'sm' }>(), {
  size: 'default',
})

const emits = defineEmits<SwitchRootEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <SwitchRoot
    v-bind="forwarded"
    :class="cn(
      'peer inline-flex shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=unchecked]:bg-input',
      props.size === 'sm' ? 'h-5 w-9' : 'h-6 w-11',
      props.class,
    )"
  >
    <SwitchThumb
      :class="cn(
        'pointer-events-none block rounded-full bg-background shadow-lg ring-0 transition-transform data-[state=unchecked]:translate-x-0',
        props.size === 'sm' ? 'h-4 w-4 data-[state=checked]:translate-x-4' : 'h-5 w-5 data-[state=checked]:translate-x-5'
      )"
    />
  </SwitchRoot>
</template>
