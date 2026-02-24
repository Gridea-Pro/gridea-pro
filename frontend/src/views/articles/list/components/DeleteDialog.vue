<template>
    <DeleteConfirmDialog v-model:open="openModel" :confirm-text="t('common.delete')" @confirm="$emit('confirm')" />
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'

const props = defineProps<{
    open: boolean
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
    confirm: []
}>()

const { t } = useI18n()

const openModel = computed({
    get: () => props.open,
    set: (val: boolean) => emit('update:open', val),
})
</script>
