<template>
    <Sheet :open="open" @update:open="$emit('update:open', $event)">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
            <SheetHeader class="px-6 py-6 border-b">
                <SheetTitle>{{ t('nav.category') }}</SheetTitle>
            </SheetHeader>

            <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
                <div class="space-y-4">
                    <div>
                        <Label class="mb-1 block">{{ t('category.name') }} <span
                                class="text-destructive">*</span></Label>
                        <Input :model-value="form.name" @input="$emit('name-change', $event)" />
                    </div>
                    <div>
                        <Label class="mb-1 block">{{ t('category.url') }} <span
                                class="text-destructive">*</span></Label>
                        <div class="relative">
                            <span class="absolute left-3 top-2.5 text-muted-foreground text-sm">/{{ categoryPath }}/</span>
                            <Input :model-value="form.slug" class="pl-22" @input="$emit('slug-change', $event)" />
                        </div>
                    </div>
                    <div>
                        <Label class="mb-1 block">{{ t('category.description') }}</Label>
                        <Textarea v-model="form.description" rows="3" />
                    </div>
                    <div>
                        <Label class="mb-1 block">{{ t('category.cover') }}</Label>
                        <div
                            class="group/cover relative w-full aspect-[16/10] border border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors overflow-hidden bg-background"
                            @click="$emit('upload-cover')">
                            <img v-if="coverPreview" :src="coverPreview" class="w-full h-full object-cover" />
                            <div v-else class="flex flex-col items-center text-muted-foreground">
                                <i class="ri-image-add-line text-2xl mb-1"></i>
                                <span class="text-xs">{{ t('category.coverHint') }}</span>
                            </div>
                            <button
                                v-if="form.cover"
                                class="hidden group-hover/cover:flex absolute top-2 right-2 bg-destructive hover:bg-destructive/90 text-white rounded-full w-5 h-5 items-center justify-center z-10 shadow-sm border border-white transition-colors cursor-pointer"
                                :title="t('category.coverRemove')" @click.stop="$emit('remove-cover')">
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                    class="w-3.5 h-3.5">
                                    <path
                                        d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
                                </svg>
                            </button>
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

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { DEFAULT_CATEGORY_PATH } from '@/helpers/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetFooter } from '@/components/ui/sheet'

defineProps<{
    open: boolean
    form: any
    canSubmit: boolean
    coverPreview: string
}>()

defineEmits(['update:open', 'close', 'save', 'name-change', 'slug-change', 'upload-cover', 'remove-cover'])

const { t } = useI18n()
const siteStore = useSiteStore()

// 前缀提示要和实际渲染出的路径一致，否则用户按提示拼出来的链接是错的
const categoryPath = computed(() => siteStore.site.themeConfig?.categoryPath || DEFAULT_CATEGORY_PATH)
</script>
