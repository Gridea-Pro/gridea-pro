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

    // slug 必须是 URL-safe + 跨平台文件系统安全的：小写字母 / 数字 / 中间的连字符。
    // 与后端 utils.ValidateSlug 的正则保持一致。
    const slugPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/

    // checkCategoryValid 依次检查 slug 合法性 → name 冲突 → slug 冲突。
    // 唯一性比对走 toLowerCase，与后端 EqualFold 对齐。
    const checkCategoryValid = (): { ok: true } | { ok: false; reason: 'slugInvalid' | 'name' | 'slug' } => {
        if (!slugPattern.test(form.slug)) {
            return { ok: false, reason: 'slugInvalid' }
        }
        const categories = [...siteStore.categories]
        if (isUpdate.value) {
            categories.splice(form.index, 1)
        }
        const nameLower = form.name.toLowerCase()
        const slugLower = form.slug.toLowerCase()
        if (categories.some((c: ICategory) => c.name.toLowerCase() === nameLower)) {
            return { ok: false, reason: 'name' }
        }
        if (categories.some((c: ICategory) => (c.slug || '').toLowerCase() === slugLower)) {
            return { ok: false, reason: 'slug' }
        }
        return { ok: true }
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

        const check = checkCategoryValid()
        if (!check.ok) {
            const key =
                check.reason === 'slugInvalid' ? 'category.slugInvalid'
                : check.reason === 'name' ? 'category.nameRepeat'
                : 'category.urlRepeat'
            toast.error(t(key))
            return
        }

        try {
            const result = await SaveCategoryFromFrontend({
                id: form.id,
                name: form.name,
                slug: form.slug,
                description: form.description,
                originalSlug: '',
            })

            if (result) {
                siteStore.categories = result.categories as ICategory[]
                categoryList.value = [...result.categories as ICategory[]]
                if (result.posts) {
                    siteStore.posts = result.posts
                }
                toast.success(t('category.saved'))
                visible.value = false
            }
        } catch (e: any) {
            // Wails v2 把 Go error 序列化成字符串，优先取字符串本体
            const raw = typeof e === 'string' ? e : (e?.message || '')
            if (raw.startsWith('invalid slug')) {
                toast.error(t('category.slugInvalid'))
            } else {
                toast.error(raw || t('category.saveError'))
            }
        }
    }

    const confirmDelete = (id: string) => {
        categoryToDelete.value = id
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (categoryToDelete.value) {
            try {
                const result = await DeleteCategoryFromFrontend(categoryToDelete.value)
                if (result) {
                    siteStore.categories = result.categories as ICategory[]
                    categoryList.value = [...result.categories as ICategory[]]
                    if (result.posts) {
                        siteStore.posts = result.posts
                    }
                    toast.success(t('category.deleted'))
                }
            } catch (e: any) {
                const msg = typeof e === 'string' ? e : (e?.message || t('category.deleteError'))
                toast.error(msg)
            }
        }
        deleteModalVisible.value = false
        categoryToDelete.value = null
    }

    const handleCategorySort = async () => {
        try {
            await SaveCategories(JSON.parse(JSON.stringify(categoryList.value)))
        } catch (e: any) {
            const msg = typeof e === 'string' ? e : (e?.message || t('category.sortError'))
            toast.error(msg)
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
