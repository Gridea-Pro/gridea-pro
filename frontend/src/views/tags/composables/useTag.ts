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

    // slug 必须是 URL-safe + 跨平台文件系统安全的：小写字母 / 数字 / 中间的连字符。
    // 与后端 utils.ValidateSlug 的正则保持一致，避免后端兜底拒绝后用户才发现。
    const slugPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/

    // checkTagValid 依次检查 slug 合法性 → name 冲突 → slug 冲突。
    // 唯一性比对走 toLowerCase，与后端 EqualFold 对齐 —— macOS/Windows 文件系统
    // 大小写不敏感，`Go` 和 `GO` 不应被放行为两个不同 slug。
    const checkTagValid = (): { ok: true } | { ok: false; reason: 'slugInvalid' | 'name' | 'slug' } => {
        if (!slugPattern.test(form.slug)) {
            return { ok: false, reason: 'slugInvalid' }
        }
        const tags = [...siteStore.tags]
        if (isUpdate.value) {
            tags.splice(form.index, 1)
        }
        const nameLower = form.name.toLowerCase()
        const slugLower = form.slug.toLowerCase()
        if (tags.some((t: ITag) => t.name.toLowerCase() === nameLower)) {
            return { ok: false, reason: 'name' }
        }
        if (tags.some((t: ITag) => (t.slug || '').toLowerCase() === slugLower)) {
            return { ok: false, reason: 'slug' }
        }
        return { ok: true }
    }

    const saveTag = async () => {
        buildSlug()

        const check = checkTagValid()
        if (!check.ok) {
            const key =
                check.reason === 'slugInvalid' ? 'tag.slugInvalid'
                : check.reason === 'name' ? 'tag.nameRepeat'
                : 'tag.urlRepeat'
            toast.error(t(key))
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
            // Wails v2 把 Go error 序列化成字符串，所以 e 可能是 string 而不是 Error 对象；
            // 仅当两者都不可用时再回落到通用错误文案。
            const raw = typeof e === 'string' ? e : (e?.message || '')
            // 后端 utils.ErrInvalidSlug 的 Error() 前缀 —— 即便前端校验被绕过（MCP / 直调 API）
            // 也能按当前语言展示本地化文案，而不是原样抛出中文 backend 字符串。
            if (raw.startsWith('invalid slug')) {
                toast.error(t('tag.slugInvalid'))
            } else {
                toast.error(raw || t('tag.saveError'))
            }
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
                    const msg = typeof e === 'string' ? e : (e?.message || t('tag.deleteError'))
                    toast.error(msg)
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
            const msg = typeof e === 'string' ? e : (e?.message || t('tag.sortError'))
            toast.error(msg)
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
