import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import urlJoin from 'url-join'
import { MenuTypes } from '@/helpers/enums'
import type { IMenu } from '@/interfaces/menu'
import type { IPost } from '@/interfaces/post'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import { SaveMenuFromFrontend, DeleteMenuFromFrontend, SaveMenus } from '@/wailsjs/go/facade/MenuFacade'
import { domain, facade } from '@/wailsjs/go/models'
import { EventsEmit } from '@/wailsjs/runtime'

interface IForm {
    name: any
    index: any
    openType: string
    link: string
}

export function useMenu() {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    const menuList = ref<IMenu[]>([])
    const visible = ref(false)
    const menuTypes = MenuTypes
    const deleteModalVisible = ref(false)
    const menuToDelete = ref<number | null>(null)

    const form = reactive<IForm>({
        name: '',
        index: '',
        openType: MenuTypes.Internal,
        link: '',
    })

    const handleNameChange = (val: string) => {
        form.name = val
    }
    const handleOpenTypeChange = (val: string) => {
        form.openType = val
    }
    const handleLinkChange = (val: string) => {
        form.link = val
    }

    const menuLinks = computed(() => {
        const { setting, themeConfig } = siteStore.site
        const domain = setting.domain || ''
        const posts = siteStore.posts.map((item: IPost) => {
            return {
                text: `📄 ${item.title}`,
                value: urlJoin(domain, themeConfig.postPath || 'post', item.fileName || ''),
            }
        })
        return [
            {
                text: '🏠 Homepage',
                value: domain,
            },
            {
                text: '📚 Archives',
                value: urlJoin(domain, 'archives'),
            },
            {
                text: '🏷️ Tags',
                value: urlJoin(domain, themeConfig.tagPath || 'tags'),
            },
            ...posts,
        ].filter((item) => typeof item.value === 'string' && item.value.trim() !== '')
    })

    const canSubmit = computed(() => {
        return !!(form.name && form.link)
    })

    const newMenu = () => {
        form.name = null
        form.index = null
        form.openType = MenuTypes.Internal
        form.link = ''
        visible.value = true

        ga('Menu', 'Menu - new', siteStore.site.setting.domain)
    }

    const closeSheet = () => {
        visible.value = false
    }

    const editMenu = (menu: IMenu, index: number) => {
        visible.value = true
        form.index = index
        form.name = menu.name
        form.openType = menu.openType
        form.link = menu.link
    }

    const saveMenu = async () => {
        try {
            const menuForm = new facade.MenuForm({
                name: form.name,
                openType: form.openType,
                link: form.link,
                index: form.index,
            })
            const menus = await SaveMenuFromFrontend(menuForm)

            if (menus) {
                siteStore.menus = menus
                menuList.value = [...menus]
                toast.success(t('siteMenu.saved'))
                visible.value = false
                ga('Menu', 'Menu - save', form.name)
            }
        } catch (e: any) {
            toast.error(e.message || 'Error saving menu')
        }
    }

    const confirmDelete = (index: number) => {
        menuToDelete.value = index
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (menuToDelete.value !== null) {
            try {
                const menus = await DeleteMenuFromFrontend(menuToDelete.value)
                if (menus) {
                    siteStore.menus = menus
                    menuList.value = [...menus]
                    toast.success(t('siteMenu.deleted'))
                }
            } catch (e: any) {
                toast.error(e.message || 'Error deleting menu')
            }
        }
        deleteModalVisible.value = false
        menuToDelete.value = null
    }

    const handleMenuSort = async () => {
        try {
            const menus = menuList.value.map(m => new domain.Menu(m))
            await SaveMenus(menus)
        } catch (e: any) {
            toast.error(e.message || 'Error sorting menu')
        }
    }

    onMounted(() => {
        menuList.value = [...siteStore.menus]
    })

    return {
        menuList,
        visible,
        menuTypes,
        deleteModalVisible,
        form,
        menuLinks,
        canSubmit,
        newMenu,
        closeSheet,
        editMenu,
        saveMenu,
        confirmDelete,
        handleDelete,
        handleMenuSort,
        handleNameChange,
        handleOpenTypeChange,
        handleLinkChange,
    }
}
