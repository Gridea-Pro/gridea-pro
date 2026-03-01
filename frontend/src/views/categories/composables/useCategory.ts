import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ICategory } from '@/stores/site'
import { generateId } from '@/utils/id'
import slug from '@/helpers/slug'
import { toast } from '@/helpers/toast'
import { SaveCategoryFromFrontend, DeleteCategoryFromFrontend, SaveCategories } from '@/wailsjs/go/facade/CategoryFacade'

export function useCategory() {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    const visible = ref(false)
    const isUpdate = ref(false)
    const slugChanged = ref(false)
    const deleteModalVisible = ref(false)
    const categoryToDelete = ref<string | null>(null)

    // Note: We use a separate local list for draggable
    const categoryList = ref<ICategory[]>([])

    interface IForm {
        id: string         // 不可变 UUID（新建时为空）
        name: string
        slug: string
        description: string
        index: number
    }

    const form = reactive<IForm>({
        id: '',
        name: '',
        slug: '',
        description: '',
        index: -1,
    })

    const canSubmit = computed(() => {
        return !!(form.name && form.slug)
    })

    // Initialize list from store
    onMounted(() => {
        categoryList.value = [...siteStore.categories]
    })

    // Make list reactive to store updates if needed, though we mainly update it manually after save
    // We can watch siteStore.categories if we want two-way sync but manual update is safer for draggable

    const handleNameChange = (e: any) => {
        const val = e.target ? e.target.value : e
        form.name = val
        if (!slugChanged.value) {
            form.slug = slug(val)
        }
    }

    const handleSlugChange = (e: any) => {
        const val = e.target ? e.target.value : e
        form.slug = val
        slugChanged.value = !!val
    }

    const buildSlug = () => {
        if (form.slug === '') {
            form.slug = slug(form.name) || generateId()
        }
    }

    const checkCategoryValid = () => {
        const categories = [...siteStore.categories]
        if (isUpdate.value) {
            categories.splice(form.index, 1)
        }
        const foundIndex = categories.findIndex((c: ICategory) => c.slug === form.slug)
        return foundIndex === -1
    }

    const openCreateSheet = () => {
        form.id = ''
        form.name = ''
        form.slug = ''
        form.description = ''
        form.index = -1
        slugChanged.value = false
        visible.value = true
        isUpdate.value = false
    }

    const openEditSheet = (category: ICategory, index: number) => {
        visible.value = true
        isUpdate.value = true
        form.id = category.id || ''  // 加载不可变 UUID
        form.name = category.name
        form.slug = category.slug
        form.description = category.description || ''
        form.index = index
        slugChanged.value = true
    }

    const closeSheet = () => {
        visible.value = false
    }

    const saveCategory = async () => {
        buildSlug()

        const valid = checkCategoryValid()
        if (!valid) {
            toast.error(t('category.urlRepeat'))
            return
        }

        try {
            const categories = await SaveCategoryFromFrontend({
                id: form.id,        // 空则新建，非空则按 ID 更新
                name: form.name,
                slug: form.slug,
                description: form.description,
                originalSlug: '',   // 已废弃，保留防能老客户端
            })

            if (categories) {
                siteStore.categories = categories
                categoryList.value = [...categories]
                toast.success(t('category.saved'))
                visible.value = false
            }
        } catch (e: any) {
            toast.error(e.message || 'Error saving category')
        }
    }

    const confirmDelete = (id: string) => {
        categoryToDelete.value = id
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (categoryToDelete.value) {
            try {
                const categories = await DeleteCategoryFromFrontend(categoryToDelete.value)
                if (categories) {
                    siteStore.categories = categories
                    categoryList.value = [...categories]
                    toast.success(t('category.deleted'))
                }
            } catch (e: any) {
                toast.error(e.message || 'Error deleting category')
            }
        }
        deleteModalVisible.value = false
        categoryToDelete.value = null
    }

    const handleCategorySort = async () => {
        try {
            await SaveCategories(JSON.parse(JSON.stringify(categoryList.value)))
        } catch (e: any) {
            toast.error(e.message || 'Error sorting categories')
        }
    }

    return {
        visible,
        form,
        categoryList,
        deleteModalVisible,
        canSubmit,
        openCreateSheet,
        openEditSheet,
        closeSheet,
        saveCategory,
        confirmDelete,
        handleDelete,
        handleCategorySort,
        handleNameChange,
        handleSlugChange,
    }
}
