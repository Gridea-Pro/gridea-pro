<template>
    <Sheet :open="open" @update:open="$emit('update:open', $event)">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
            <SheetHeader class="px-6 py-6 border-b">
                <SheetTitle>{{ t('nav.tag') }}</SheetTitle>
            </SheetHeader>

            <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
                <div class="space-y-4">
                    <div>
                        <Label class="mb-1.5 block">{{ t('tag.name') }} <span class="text-destructive">*</span></Label>
                        <Input :model-value="form.name" @update:model-value="$emit('name-change', $event as string)" />
                    </div>
                    <div>
                        <Label class="mb-1.5 block">{{ t('tag.slug') }} <span class="text-destructive">*</span></Label>
                        <div class="relative">
                            <span class="absolute left-3 top-2.5 text-muted-foreground text-sm">/tags/</span>
                            <Input
:model-value="form.slug" class="pl-14"
                                @update:model-value="$emit('slug-change', $event as string)" />
                        </div>
                    </div>
                    <div>
                        <Label class="mb-3 block">{{ t('tag.color') }}</Label>
                        <div class="flex flex-wrap gap-2">
                            <div
v-for="color in presetColors" :key="color"
                                class="w-6 h-6 rounded-full cursor-pointer transition-transform hover:scale-110 border border-transparent"
                                :class="{ 'ring-2 ring-primary ring-offset-2': form.color === color }"
                                :style="{ backgroundColor: color }" @click="$emit('color-change', color)"></div>
                            <div
                                class="relative w-6 h-6 rounded-full overflow-hidden border border-border cursor-pointer">
                                <input
type="color" :value="form.color" class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                                    title="Custom Color"
                                    @input="handleColorEvent" />
                                <div
v-if="!presetColors.includes(form.color || '')"
                                    class="absolute inset-0 bg-gradient-to-br from-red-500 via-green-500 to-blue-500 pointer-events-none"></div>
                                <div v-else class="absolute inset-0" :style="{ backgroundColor: form.color }"></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
                <Button
variant="outline"
                    class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                    @click="$emit('close')">
                    {{ t('common.cancel') }}
                </Button>
                <Button
variant="default"
                    class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                    :disabled="!canSubmit" @click="$emit('save')">
                    {{ t('common.save') }}
                </Button>
            </SheetFooter>
        </SheetContent>
    </Sheet>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'

defineProps<{
    open: boolean
    form: {
        name: string
        slug: string
        color?: string
    }
    presetColors: string[]
    canSubmit: boolean
}>()

const emit = defineEmits<{
    (e: 'update:open', value: boolean): void
    (e: 'name-change', value: string): void
    (e: 'slug-change', value: string): void
    (e: 'color-change', value: string): void
    (e: 'close'): void
    (e: 'save'): void
}>()

const { t } = useI18n()

const handleColorEvent = (e: Event) => {
    const target = e.target as HTMLInputElement
    emit('color-change', target.value)
}
</script>
