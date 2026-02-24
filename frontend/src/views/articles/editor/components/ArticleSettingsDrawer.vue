<template>
    <Sheet v-model:open="openModel">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
            <SheetHeader class="px-6 py-6 border-b">
                <SheetTitle>{{ $t('article.settings') }}</SheetTitle>
            </SheetHeader>

            <div class="relative flex-1 px-6 py-6 space-y-6 overflow-y-auto">
                <!-- URL -->
                <div class="space-y-2">
                    <Label>URL</Label>
                    <Input v-model="form.fileName" @change="(e: any) => $emit('fileNameChange', e)" />
                </div>

                <!-- Categories -->
                <div class="space-y-2">
                    <Label>{{ $t('nav.category') }}</Label>
                    <Select v-model="form.category">
                        <SelectTrigger class="w-full">
                            <SelectValue :placeholder="$t('selectCategory')" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="_none_">{{ $t('none') }}</SelectItem>
                            <SelectItem v-for="c in availableCategories" :key="c" :value="c">{{ c }}</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <!-- Tags -->
                <div class="space-y-2">
                    <Label>{{ $t('nav.tag') }}</Label>
                    <div>
                        <div class="flex flex-wrap gap-2 p-2 border rounded-md bg-background min-h-[32px] mb-2">
                            <span v-for="tag in form.tags" :key="tag"
                                class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary/10 border border-primary/20 text-xs text-primary/80">
                                {{ tag }}
                                <button @click="$emit('removeTag', tag)"
                                    class="ml-1 text-primary/60 hover:text-destructive">
                                    <XMarkIcon class="size-3" />
                                </button>
                            </span>
                            <input :value="tagInput"
                                @input="$emit('update:tagInput', ($event.target as HTMLInputElement).value)"
                                @keydown.enter.prevent="$emit('addTag')"
                                class="flex-1 min-w-[80px] bg-transparent outline-none text-foreground text-sm px-1"
                                placeholder="Add tag..." />
                        </div>
                        <div class="flex flex-wrap gap-2 max-h-[120px] overflow-y-auto p-1 border rounded-md">
                            <span v-for="t in availableTags" :key="t" @click="$emit('selectTag', t)"
                                class="cursor-pointer text-xs px-2 py-1 rounded-full bg-primary/5 hover:bg-primary/15 border border-primary/10 transition-colors select-none text-muted-foreground">
                                {{ t }}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- Date -->
                <div class="space-y-2">
                    <Label>{{ $t('article.createAt') }}</Label>
                    <Popover>
                        <PopoverTrigger as-child>
                            <Button variant="outline" :class="cn(
                                'w-full justify-start text-left font-normal hover:bg-primary/5 hover:text-primary border-primary/20 cursor-pointer',
                                !dateValue && 'text-muted-foreground',
                            )">
                                <CalendarIcon class="mr-2 h-4 w-4" />
                                {{ (form.date && form.date.isValid()) ? form.date.format('YYYY-MM-DD HH:mm:ss') :
                                    $t('pickDate') }}
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent class="w-auto p-0" align="start">
                            <Calendar :model-value="(dateValue as any)"
                                @update:model-value="(val: any) => $emit('update:dateValue', val)" show-week-number />
                            <div class="border-t p-3">
                                <Label class="text-xs text-muted-foreground mb-2 block capitalize">{{ $t('time')
                                    }}</Label>
                                <div class="relative">
                                    <ClockIcon class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground z-10" />
                                    <Input type="time" step="1" :model-value="timeValue"
                                        @update:model-value="(val: any) => $emit('update:timeValue', val as string)"
                                        class="h-9 pl-9 accent-primary selection:bg-primary selection:text-primary-foreground" />
                                </div>
                            </div>
                        </PopoverContent>
                    </Popover>
                </div>

                <!-- Feature Image -->
                <div class="space-y-2">
                    <Label>{{ $t('article.featureImage') }}</Label>
                    <div class="space-y-2">
                        <Input :model-value="featureDisplayValue"
                            @update:model-value="(val: any) => $emit('update:featureDisplayValue', val as string)"
                            :placeholder="$t('article.featureImagePlaceholder') || 'Image URL or Local Path'" />

                        <div class="feature-uploader cursor-pointer border border-dashed rounded-md p-4 text-center hover:border-primary transition-colors bg-background"
                            @click="$emit('selectFeatureImage')">
                            <div v-if="featureImagePreviewSrc">
                                <img class="feature-image mx-auto max-h-[150px] object-cover rounded-md"
                                    :src="featureImagePreviewSrc" />
                            </div>
                            <div v-else>
                                <img src="@/assets/images/image_upload.svg" class="upload-img mx-auto w-20">
                                <i class="ri-upload-2-line upload-icon text-lg mt-2 block text-muted-foreground"></i>
                                <div class="text-xs text-muted-foreground mt-2">点击选择本地图片</div>
                            </div>
                        </div>

                        <Button v-if="featureDisplayValue" variant="destructive" size="sm" class="mt-2 w-full"
                            @click.stop="$emit('clearFeatureImage')">
                            <template #icon>
                                <TrashIcon class="size-4 mr-2" />
                            </template>
                            {{ $t('common.delete') }}
                        </Button>
                    </div>
                </div>

                <!-- Hide in List -->
                <div class="flex items-center justify-between">
                    <Label>{{ $t('article.hideInList') }}</Label>
                    <Switch size="sm" v-model:checked="form.hideInList" />
                </div>

                <!-- Top Article -->
                <div class="flex items-center justify-between">
                    <Label>{{ $t('article.top') }}</Label>
                    <Switch size="sm" v-model:checked="form.isTop" />
                </div>
            </div>

            <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
                <Button variant="outline"
                    class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                    @click="openModel = false">
                    {{ $t('common.cancel') }}
                </Button>
                <Button variant="default"
                    class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                    @click="$emit('confirmPublish')">
                    {{ $t('article.publish') }}
                </Button>
            </SheetFooter>
        </SheetContent>
    </Sheet>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Calendar } from '@/components/ui/calendar'
import { Sheet, SheetContent, SheetTitle, SheetHeader, SheetFooter } from '@/components/ui/sheet'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CalendarIcon, ClockIcon, TrashIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import type { DateValue } from '@internationalized/date'
import type { ArticleFormState } from '../composables/useArticleForm'

const props = defineProps<{
    open: boolean
    form: ArticleFormState
    tagInput: string
    availableTags: string[]
    availableCategories: string[]
    dateValue: DateValue
    timeValue: string
    featureDisplayValue: string
    featureImagePreviewSrc: string
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
    'update:tagInput': [value: string]
    'update:dateValue': [value: DateValue]
    'update:timeValue': [value: string]
    'update:featureDisplayValue': [value: string]
    addTag: []
    removeTag: [tag: string]
    selectTag: [tag: string]
    fileNameChange: [event: Event]
    selectFeatureImage: []
    clearFeatureImage: []
    confirmPublish: []
}>()

const openModel = computed({
    get: () => props.open,
    set: (val: boolean) => emit('update:open', val),
})
</script>
