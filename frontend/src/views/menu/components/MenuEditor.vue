<template>
    <Sheet :open="open" @update:open="$emit('update:open', $event)">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
            <SheetHeader class="px-6 py-6 border-b">
                <SheetTitle>
                    <span v-if="parentName" class="text-muted-foreground font-normal">
                        {{ parentName }} /
                    </span>
                    {{ parentName ? t('siteMenu.submenu') : t('nav.menu') }}
                </SheetTitle>
            </SheetHeader>

            <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
                <div class="space-y-4">
                    <div>
                        <Label class="mb-1 block">{{ t('siteMenu.name') }}</Label>
                        <Input :model-value="form.name" @update:model-value="$emit('name-change', $event as string)" />
                    </div>
                    <div>
                        <Label class="mb-1 block">{{ t('siteMenu.openType') }}</Label>
                        <Select
:model-value="form.openType"
                            @update:model-value="$emit('open-type-change', $event as string)">
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem v-for="item in menuTypes" :key="item" :value="item">
                                    {{ item }}
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div>
                        <Label class="mb-1 block">{{ t('siteMenu.link') }}</Label>
                        <div class="space-y-2">
                            <div class="relative">
                                <Input
:model-value="form.link" :placeholder="t('siteMenu.inputOrSelect')"
                                    @update:model-value="$emit('link-change', $event as string)" />
                            </div>
                            <Select
:model-value="form.link"
                                @update:model-value="$emit('link-change', $event as string)">
                                <SelectTrigger>
                                    <SelectValue :placeholder="t('siteMenu.selectInternalLink')" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem v-for="item in menuLinks" :key="item.value" :value="item.value">
                                        <span class="truncate max-w-[300px] block" :title="item.text">{{ item.text
                                            }}</span>
                                    </SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
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

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'

defineProps<{
    open: boolean
    form: {
        name: string
        openType: string
        link: string
    }
    menuTypes: any
    menuLinks: Array<{ text: string, value: string }>
    canSubmit: boolean
    parentName?: string
}>()

defineEmits<{
    (e: 'update:open', value: boolean): void
    (e: 'name-change', value: string): void
    (e: 'open-type-change', value: string): void
    (e: 'link-change', value: string): void
    (e: 'close'): void
    (e: 'save'): void
}>()

const { t } = useI18n()
</script>
