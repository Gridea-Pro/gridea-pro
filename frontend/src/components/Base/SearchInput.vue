<template>
    <div class="flex items-center">
        <div v-if="isVisible" class="relative">
            <input
ref="inputRef"
                v-model="inputValue"
                class="w-[200px] h-8 pl-8 text-xs rounded-full bg-muted/50 border-none focus:ring-1 focus:ring-primary/20 outline-none transition-all placeholder:text-muted-foreground" :placeholder="placeholder" autofocus @blur="handleBlur" />
            <MagnifyingGlassIcon class="absolute left-2.5 top-2 size-4 text-muted-foreground pointer-events-none" />
        </div>
        <div
v-else
            class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
            :title="placeholder" @click="showInput">
            <MagnifyingGlassIcon class="size-4" />
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, watch, nextTick } from 'vue'
import { MagnifyingGlassIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
    modelValue: string
    placeholder?: string
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void
}>()

const isVisible = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const inputValue = ref(props.modelValue)

// Initialize visibility based on modelValue
if (props.modelValue) {
    isVisible.value = true
}

watch(() => props.modelValue, (newValue) => {
    if (newValue) {
        isVisible.value = true
    }
    if (newValue !== inputValue.value) {
        inputValue.value = newValue
    }
})

watch(inputValue, (newValue) => {
    console.log('SearchInput emit:', newValue)
    emit('update:modelValue', newValue)
})

const showInput = async () => {
    isVisible.value = true
    await nextTick()
    inputRef.value?.focus()
}

const handleBlur = () => {
    if (!inputValue.value) {
        isVisible.value = false
    }
}
</script>
