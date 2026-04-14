import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ITag } from '@/stores/site'
import { generateId } from '@/utils/id'
import slugFn from '@/helpers/slug'
import { toast } from '@/helpers/toast'
import { GetTagColors, SaveTagFromFrontend, DeleteTagFromFrontend, SaveTags } from '@/wailsjs/go/facade/TagFacade'
import { domain, facade } from '@/wailsjs/go/models'

interface IForm {
    name: string
    slug: string
    index: number
    originalSlug?: string
    originalName?: string
    color?: string
}

export function useTag() {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    const visible = ref(false)
    const isUpdate = ref(false)
    const slugChanged = ref(false)
    const deleteModalVisible = ref(false)
    const tagToDelete = ref<string | null>(null)

    const form = reactive<IForm>({
        name: '',
        slug: '',
        index: -1,
        originalSlug: '',
        originalName: '',
        color: '#3b82f6',
    })

    const presetColors = ref<string[]>([])
    const tagList = ref<ITag[]>([])

    const canSubmit = computed(() => {
        return !!(form.name && form.slug)
    })

    onMounted(async () => {
        presetColors.value = await GetTagColors()
        tagList.value = [...siteStore.tags]
    })

    watch(() => siteStore.tags, (newTags) => {
        tagList.value = [...newTags]
    })

    const handleNameChange = (val: string) => {
        form.name = val
        if (!slugChanged.value) {
            form.slug = slugFn(val)
        }
    }

    const handleSlugChange = (val: string) => {
        form.slug = val
        slugChanged.value = !!val
    }

    const handleColorChange = (color: string) => {
        form.color = color
    }

    const openCreateSheet = () => {
        form.name = ''
        form.slug = ''
        form.index = -1
        form.originalSlug = ''
        form.originalName = ''
        if (presetColors.value.length > 0) {
            form.color = presetColors.value[Math.floor(Math.random() * presetColors.value.length)]
        } else {
            form.color = '#3b82f6'
        }
        slugChanged.value = false
        visible.value = true
        isUpdate.value = false
    }

    const editTag = (tag: ITag, index: number) => {
        visible.value = true
        isUpdate.value = true
        form.name = tag.name
        form.slug = tag.slug || ''
        form.index = index
        form.originalSlug = tag.slug || ''
        form.originalName = tag.name
        form.color = tag.color || '#3b82f6'
        slugChanged.value = true
    }

    const closeSheet = () => {
        visible.value = false
    }

    const buildSlug = () => {
        if (form.slug === '') {
            form.slug = slugFn(form.name) || generateId()
        }
    }

    const checkTagValid = () => {
        const tags = [...siteStore.tags]
        if (isUpdate.value) {
            tags.splice(form.index, 1)
        }
        const foundIndex = tags.findIndex((t: ITag) => t.slug === form.slug)
        return foundIndex === -1
    }

    const saveTag = async () => {
        buildSlug()

        const valid = checkTagValid()
        if (!valid) {
            toast.error(t('tag.urlRepeat'))
            return
        }

        try {
            const tagForm = new facade.TagForm({
                name: form.name,
                slug: form.slug,
                color: form.color || '',
                originalName: form.originalName || '',
            })
            const result = await SaveTagFromFrontend(tagForm)

            if (result) {
                siteStore.tags = result.tags
                tagList.value = [...result.tags]
                if (result.posts) {
                    siteStore.posts = result.posts
                }
                toast.success(t('tag.saved'))
                visible.value = false
            }
        } catch (e: any) {
            toast.error(e.message || 'Error saving tag')
        }
    }

    const confirmDelete = (slug: string) => {
        tagToDelete.value = slug
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (tagToDelete.value) {
            const tag = siteStore.tags.find(t => t.slug === tagToDelete.value)
            if (tag) {
                try {
                    const result = await DeleteTagFromFrontend(tag.name)
                    if (result) {
                        siteStore.tags = result.tags
                        tagList.value = [...result.tags]
                        if (result.posts) {
                            siteStore.posts = result.posts
                        }
                        toast.success(t('tag.deleted'))
                    }
                } catch (e: any) {
                    toast.error(e.message || 'Error deleting tag')
                }
            }
        }
        deleteModalVisible.value = false
        tagToDelete.value = null
    }

    const handleTagSort = async () => {
        try {
            const tags = tagList.value.map(t => new domain.Tag(t))
            await SaveTags(tags)
        } catch (e: any) {
            toast.error(e.message || 'Error sorting tags')
        }
    }

    return {
        visible,
        form,
        tagList,
        presetColors,
        deleteModalVisible,
        canSubmit,
        openCreateSheet,
        editTag,
        closeSheet,
        saveTag,
        confirmDelete,
        handleDelete,
        handleTagSort,
        handleNameChange,
        handleSlugChange,
        handleColorChange,
    }
}
