<template>
    <Dialog v-model:open="openModel">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>{{ t('common.warning') }}</DialogTitle>
            </DialogHeader>
            <div class="flex items-center gap-3 text-destructive">
                <ExclamationTriangleIcon class="size-6" />
                <p class="text-sm text-foreground">{{ t('article.unsavedWarning') }}</p>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="openModel = false">{{ t('common.cancel') }}</Button>
                <Button @click="$emit('confirmClose')">{{ t('article.noSaveAndBack') }}</Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
    open: boolean
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
    confirmClose: []
}>()

const { t } = useI18n()

const openModel = computed({
    get: () => props.open,
    set: (val: boolean) => emit('update:open', val),
})
</script>
