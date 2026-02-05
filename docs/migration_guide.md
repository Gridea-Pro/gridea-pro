# I18n Migration Guide

We have refactored the i18n system to use standard JSON files with a **Semantic Structure**. This ensures a Single Source of Truth for both Vue frontend and Wails backend.

## 1. Directory Structure

Old structure (deprecated):
- `frontend/src/assets/locales.ts`
- `frontend/src/assets/locales-menu.ts`

**New structure:**
- `frontend/src/locales/*.json` (e.g., `en.json`, `zh-CN.json`)
- `frontend/src/locales/index.ts` (Vue i18n entry)
- `i18n_helper.go` (Go helper at project root)

## 2. New Semantic Structure

The flat structure has been replaced by nested, logical namespaces:

- **common**: Generic actions (`save`, `delete`, `edit`, `cancel`).
- **nativeMenu**: System-level menu items only (`quit`, `close`, `toggledevtools`).
- **nav**: Sidebar/Navigation items.
- **dashboard**: Version info, render/sync status.
- **article**: Blog post management (`title`, `status`, `publish`).
- **tag**: Tag management.
- **category**: Category management.
- **link**: Friend links management.
- **siteMenu**: Blog navigation menu management.
- **settings**:
    - `basic`: Site info, favicon, avatar.
    - `comment`: Waline/Giscus settings.
    - `theme`: Theme selection and config.
    - `network`: Git/Server connection settings.
    - `system`: App language, version.
- **comment**: Comment management actions.

## 3. Usage in Vue 3 Components

**Setup:**
Ensure `main.ts` imports and uses the new i18n instance:
```typescript
import { createApp } from 'vue'
import App from './App.vue'
import i18n from './locales' // Import from new location

const app = createApp(App)
app.use(i18n)
app.mount('#app')
```

**In Template:**
Use `$t` with the new nested keys.

*Example 1: Navigation Item*
```html
<!-- Old -->
<p>{{ $t('preview') }}</p>
<!-- New -->
<p>{{ $t('nav.preview') }}</p>
```

*Example 2: Settings Label*
```html
<!-- Old -->
<label>{{ $t('domain') }}</label>
<!-- New -->
<label>{{ $t('settings.network.domain') }}</label>
```

*Example 3: Common Action*
```html
<button>{{ $t('common.save') }}</button>
```

**In Script (Composition API):**
```typescript
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
console.log(t('article.publishSuccess'))
```

## 4. Usage in Go (Wails Native Menu)

A new helper `i18n_helper.go` is available at the project root.

**Example usage in `main.go` or `app.go`:**

```go
func createMenu(lang string) {
    // Load native menu translations
    // NOTE: This now looks for "nativeMenu" key in JSON
    menuTrans, err := LoadMenuTranslations(lang)
    if err != nil {
        println("Error loading translations:", err.Error())
        // Handle error (maybe fallback)
    }

    // specific string
    editTitle := menuTrans["edit"] // "Edit" or "编辑"
    
    // ... construct Wails menu ...
}
```

## 5. Adding/Editing Translations

1.  Open `frontend/src/locales/{lang}.json`.
2.  Add keys under the appropriate semantic namespace.
3.  Avoid dumping everything into `ui` or root. Ideally, find the correct module (e.g. `article` vs `settings`).
4.  Go backend will automatically pick up `nativeMenu` changes after re-compilation.
