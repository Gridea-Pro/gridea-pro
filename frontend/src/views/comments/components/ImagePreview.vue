<template>
    <div
v-if="modelValue" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/30 backdrop-blur-sm"
        style="--wails-draggable: no-drag" @click="close">
        <div class="relative max-w-[90vw] max-h-[90vh] flex items-center justify-center p-4">
            <img
:src="modelValue"
                class="max-w-full max-h-[90vh] rounded-lg shadow-2xl object-contain selection:bg-transparent"
                style="user-select: none;" alt="Preview" @click.stop />
            <button
class="absolute -top-10 right-0 p-2 text-white/80 hover:text-white transition-colors"
                @click.stop="close">
                <XMarkIcon class="size-6" />
            </button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { XMarkIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
    modelValue: string | null
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string | null): void
}>()

const close = () => {
    emit('update:modelValue', null)
}
</script>
