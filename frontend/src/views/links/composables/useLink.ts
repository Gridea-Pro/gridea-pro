import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore, type ILink } from '@/stores/site'
import { customAlphabet } from 'nanoid'
import { toast } from '@/helpers/toast'
import { SaveLinkFromFrontend, DeleteLinkFromFrontend, SaveLinks } from '@/wailsjs/go/facade/LinkFacade'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { facade, domain } from '@/wailsjs/go/models'

export function useLink() {
    const { t } = useI18n()
    const siteStore = useSiteStore()
    const nanoid = customAlphabet('0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz', 6)

    const visible = ref(false)
    const isUpdate = ref(false)
    const deleteModalVisible = ref(false)
    const linkToDelete = ref<string | null>(null)

    // Local list for draggable
    const linkList = ref<ILink[]>([])

    interface IForm {
        name: string
        url: string
        description: string
        avatar: string
        id: string
        index: number
    }

    const form = reactive<IForm>({
        name: '',
        url: '',
        description: '',
        avatar: '',
        id: '',
        index: -1,
    })

    const canSubmit = computed(() => {
        return !!(form.name && form.url)
    })

    // Initialize list from store
    onMounted(() => {
        linkList.value = [...siteStore.links]
    })

    // Form Handlers
    const handleNameChange = (e: any) => {
        const val = e.target ? e.target.value : e
        form.name = val
    }

    const handleUrlChange = (e: any) => {
        const val = e.target ? e.target.value : e
        form.url = val
    }

    const handleAvatarChange = (e: any) => {
        const val = e.target ? e.target.value : e
        form.avatar = val
    }

    // Actions
    const openCreateSheet = () => {
        form.name = ''
        form.url = ''
        form.description = ''
        form.avatar = ''
        form.id = ''
        form.index = -1
        visible.value = true
        isUpdate.value = false
    }

    const openEditSheet = (link: ILink, index: number) => {
        visible.value = true
        isUpdate.value = true
        form.name = link.name
        form.url = link.url
        form.description = link.description || ''
        form.avatar = link.avatar || ''
        form.id = link.id
        form.index = index
    }

    const closeSheet = () => {
        visible.value = false
    }

    const saveLink = async () => {
        try {
            const linkForm = new facade.LinkForm({
                id: form.id,
                name: form.name,
                url: form.url,
                avatar: form.avatar,
                description: form.description,
            })
            const links = await SaveLinkFromFrontend(linkForm)

            if (links) {
                siteStore.links = links
                linkList.value = [...links]
                toast.success(t('link.saved'))
                visible.value = false
            }
        } catch (e: any) {
            toast.error(e.message || 'Error saving link')
        }
    }

    const confirmDelete = (id: string) => {
        linkToDelete.value = id
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (linkToDelete.value) {
            try {
                const links = await DeleteLinkFromFrontend(linkToDelete.value)
                if (links) {
                    siteStore.links = links
                    linkList.value = [...links]
                    toast.success(t('link.deleted'))
                }
            } catch (e: any) {
                toast.error(e.message || 'Error deleting link')
            }
        }
        deleteModalVisible.value = false
        linkToDelete.value = null
    }

    const openLink = (url: string) => {
        BrowserOpenURL(url)
    }

    const handleLinkSort = async () => {
        try {
            const links = linkList.value.map(l => new domain.Link(l))
            await SaveLinks(links)
        } catch (e: any) {
            toast.error(e.message || 'Error sorting links')
        }
    }

    return {
        visible,
        form,
        linkList,
        deleteModalVisible,
        canSubmit,
        openCreateSheet,
        openEditSheet,
        closeSheet,
        saveLink,
        confirmDelete,
        handleDelete,
        openLink,
        handleLinkSort,
        handleNameChange,
        handleUrlChange,
        handleAvatarChange,
    }
}
