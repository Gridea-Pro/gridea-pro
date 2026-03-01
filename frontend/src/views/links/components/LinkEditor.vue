<template>
    <Sheet :open="open" @update:open="$emit('update:open', $event)">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
            <SheetHeader class="px-6 py-6 border-b">
                <SheetTitle>{{ t('nav.link') }}</SheetTitle>
            </SheetHeader>

            <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
                <div class="space-y-4">
                    <div>
                        <Label class="mb-1.5 block">{{ t('link.name') }} <span class="text-destructive">*</span></Label>
                        <Input :model-value="form.name" @input="$emit('name-change', $event)" />
                    </div>
                    <div>
                        <Label class="mb-1.5 block">{{ t('link.url') }} <span class="text-destructive">*</span></Label>
                        <Input
:model-value="form.url" placeholder="https://example.com"
                            @input="$emit('url-change', $event)" />
                    </div>
                    <div>
                        <Label class="mb-1.5 block">{{ t('link.avatar') }}</Label>
                        <Input
:model-value="form.avatar" placeholder="https://example.com/avatar.png"
                            @input="$emit('avatar-change', $event)" />
                    </div>
                    <div>
                        <Label class="mb-1.5 block">{{ t('link.description') }}</Label>
                        <Textarea v-model="form.description" rows="3" />
                    </div>
                </div>
            </div>
            <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
                <Button
variant="outline"
                    class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                    @click="$emit('close')">{{ t('common.cancel') }}</Button>
                <Button
variant="default"
                    class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                    :disabled="!canSubmit" @click="$emit('save')">{{ t('common.save') }}</Button>
            </SheetFooter>
        </SheetContent>
    </Sheet>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'

defineProps<{
    open: boolean
    form: any
    canSubmit: boolean
}>()

defineEmits(['update:open', 'close', 'save', 'name-change', 'url-change', 'avatar-change'])

const { t } = useI18n()
</script>
