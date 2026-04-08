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

interface IForm {
    name: any
    index: any
    openType: string
    link: string
    // 子菜单上下文：null 表示顶级操作
    parentIndex: number | null
    childIndex: number | null
}

export function useMenu() {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    const menuList = ref<IMenu[]>([])
    const visible = ref(false)
    const menuTypes = MenuTypes
    const deleteModalVisible = ref(false)
    const deleteTarget = ref<{ parentIndex: number | null; index: number } | null>(null)
    const parentName = ref('')

    const form = reactive<IForm>({
        name: '',
        index: '',
        openType: MenuTypes.Internal,
        link: '',
        parentIndex: null,
        childIndex: null,
    })

    const handleNameChange = (val: string) => { form.name = val }
    const handleOpenTypeChange = (val: string) => { form.openType = val }
    const handleLinkChange = (val: string) => { form.link = val }

    const menuLinks = computed(() => {
        const { themeConfig } = siteStore.site
        const domain = siteStore.currentDomain || ''
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
        form.parentIndex = null
        form.childIndex = null
        parentName.value = ''
        visible.value = true

        ga('Menu', 'Menu - new', siteStore.currentDomain)
    }

    const newSubMenu = (pIndex: number) => {
        form.name = null
        form.index = null
        form.openType = MenuTypes.Internal
        form.link = ''
        form.parentIndex = pIndex
        form.childIndex = null
        parentName.value = menuList.value[pIndex]?.name || ''
        visible.value = true
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
        form.parentIndex = null
        form.childIndex = null
        parentName.value = ''
    }

    const editSubMenu = (pIndex: number, _parent: IMenu, cIndex: number, child: IMenu) => {
        visible.value = true
        form.parentIndex = pIndex
        form.childIndex = cIndex
        form.name = child.name
        form.openType = child.openType
        form.link = child.link
        form.index = null
        parentName.value = menuList.value[pIndex]?.name || ''
    }

    const saveMenu = async () => {
        // 子菜单保存：直接修改 menuList 并调用 SaveMenus
        if (form.parentIndex !== null) {
            const parent = menuList.value[form.parentIndex]
            if (!parent) return
            if (!parent.children) parent.children = []

            const childMenu: IMenu = {
                id: String(Date.now()),
                name: form.name,
                openType: form.openType,
                link: form.link,
            }

            if (form.childIndex !== null) {
                // 更新子菜单
                parent.children[form.childIndex] = {
                    ...parent.children[form.childIndex],
                    name: form.name,
                    openType: form.openType,
                    link: form.link,
                }
            } else {
                // 新增子菜单
                parent.children.push(childMenu)
            }

            try {
                const menus = menuList.value.map(m => new domain.Menu(m))
                await SaveMenus(menus)
                siteStore.menus = [...menuList.value]
                toast.success(t('siteMenu.saved'))
                visible.value = false
            } catch (e: any) {
                toast.error(e.message || 'Error saving submenu')
            }
            return
        }

        // 顶级菜单保存
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
        deleteTarget.value = { parentIndex: null, index }
        deleteModalVisible.value = true
    }

    const confirmDeleteChild = (parentIndex: number, childIndex: number) => {
        deleteTarget.value = { parentIndex, index: childIndex }
        deleteModalVisible.value = true
    }

    const handleDelete = async () => {
        if (!deleteTarget.value) {
            deleteModalVisible.value = false
            return
        }

        const { parentIndex, index } = deleteTarget.value

        if (parentIndex !== null) {
            // 删除子菜单
            const parent = menuList.value[parentIndex]
            if (parent?.children) {
                parent.children.splice(index, 1)
                try {
                    const menus = menuList.value.map(m => new domain.Menu(m))
                    await SaveMenus(menus)
                    siteStore.menus = [...menuList.value]
                    toast.success(t('siteMenu.deleted'))
                } catch (e: any) {
                    toast.error(e.message || 'Error deleting submenu')
                }
            }
        } else {
            // 删除顶级菜单
            try {
                const menus = await DeleteMenuFromFrontend(index)
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
        deleteTarget.value = null
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
        parentName,
        newMenu,
        newSubMenu,
        closeSheet,
        editMenu,
        editSubMenu,
        saveMenu,
        confirmDelete,
        confirmDeleteChild,
        handleDelete,
        handleMenuSort,
        handleNameChange,
        handleOpenTypeChange,
        handleLinkChange,
    }
}
