import { createI18n, type Composer, type I18nMode } from 'vue-i18n'
import { nextTick } from 'vue'
import axios from 'axios'

import en from './en.json'
import zhCN from './zh-CN.json'
import zhTW from './zh-TW.json'
import frFR from './fr-FR.json'
import ru from './ru.json'
import jaJP from './ja-JP.json'
import es from './es.json'
import ptBR from './pt-BR.json'
import de from './de.json'
import ko from './ko.json'

// 1. Definition of supported languages and messages
const messages = {
    'en': en,
    'zh-CN': zhCN,
    'zh-TW': zhTW,
    'fr-FR': frFR,
    'ru': ru,
    'ja-JP': jaJP,
    'es': es,
    'pt-BR': ptBR,
    'de': de,
    'ko': ko,
}

export type LocaleType = keyof typeof messages

// List of supported locales (e.g., ['en', 'zh-CN', 'zh-TW', ...])
const SUPPORTED_LOCALES = Object.keys(messages) as LocaleType[]
const DEFAULT_LOCALE: LocaleType = 'en'
const STORAGE_KEY = 'language'

/**
 * Normalize language code to a standard format for comparison
 * e.g., 'zh-cn' -> 'zh-CN', 'en-us' -> 'en-US' (standard matches)
 * But for our keys, we just want to match case-insensitively usually,
 * or strict exact match against our keys.
 */
function normalizeLocale(lang: string): string {
    // Simple normalization: lower case for comparison
    return lang.toLowerCase()
}

/**
 * Get the system language with strict fallback logic
 * Priority: LocalStorage > Exact Match > Mapped Match > Prefix Match > Fallback
 */
export function getLanguage(): LocaleType {
    // 1. Local Storage
    const cached = localStorage.getItem(STORAGE_KEY)
    if (cached && SUPPORTED_LOCALES.includes(cached as LocaleType)) {
        return cached as LocaleType
    }

    // 2. Get Browser Language
    // navigator.language can be 'en-US', 'zh-CN', 'fr', etc.
    const sysLang = navigator.language || 'en'
    const normalizedSys = normalizeLocale(sysLang)

    // 3. Exact Match (Case Insensitive)
    // Check if any supported locale matches the system language exactly (ignoring case)
    const exactMatch = SUPPORTED_LOCALES.find(locale => normalizeLocale(locale) === normalizedSys)
    if (exactMatch) {
        return exactMatch
    }

    // 4. Mapped Match (Special Chinese Cases)
    // zh-hk, zh-mo -> zh-TW
    // zh-sg -> zh-CN
    if (normalizedSys === 'zh-hk' || normalizedSys === 'zh-mo') {
        return 'zh-TW'
    }
    if (normalizedSys === 'zh-sg') {
        return 'zh-CN'
    }

    // 5. Prefix Match (Fuzzy)
    // e.g., 'fr-CA' -> match 'fr-FR' via prefix 'fr'
    // Strategy: match the primary subtag
    const sysPrefix = normalizedSys.split('-')[0] // 'fr-CA' -> 'fr'

    // Special handling for 'zh' prefix if not caught by exact/mapped above
    // If system is just 'zh', or 'zh-Hans', we default to 'zh-CN' if available
    // If 'zh-Hant', we usually prefer 'zh-TW'
    if (sysPrefix === 'zh') {
        if (normalizedSys.includes('hant')) {
            return 'zh-TW'
        }
        return 'zh-CN'
    }

    // General prefix search
    // Find a supported locale that starts with the prefix or is the prefix
    // e.g., if we have 'fr-FR' and sys is 'fr', or sys is 'fr-CA'
    const fuzzyMatch = SUPPORTED_LOCALES.find(locale => {
        const localePrefix = normalizeLocale(locale).split('-')[0]
        return localePrefix === sysPrefix
    })

    if (fuzzyMatch) {
        return fuzzyMatch
    }

    // 6. Fallback
    return DEFAULT_LOCALE
}

// 2. Setup vue-i18n instance
const i18n = createI18n({
    legacy: false, // Use Composition API mode
    locale: getLanguage(),
    fallbackLocale: DEFAULT_LOCALE,
    messages,
    globalInjection: true, // Optional: if you want implicit $t in templates without setup
})

/**
 * Set the language dynamically and update all related states
 */
export async function setI18nLanguage(locale: LocaleType) {
    if (!SUPPORTED_LOCALES.includes(locale)) {
        console.warn(`[I18n] Language '${locale}' is not supported.`)
        return
    }

    // 1. Update i18n instance
    if (i18n.mode === 'legacy') {
        (i18n.global as unknown as any).locale = locale
    } else {
        // In Composition API mode (legacy: false), locale is a Ref
        (i18n.global as unknown as Composer).locale.value = locale
    }

    // 2. Update LocalStorage
    localStorage.setItem(STORAGE_KEY, locale)

    // 3. Update HTML lang attribute (for a11y)
    document.querySelector('html')?.setAttribute('lang', locale)

    // 4. Set Axios Default Header
    axios.defaults.headers.common['Accept-Language'] = locale

    // 5. Wait for Vue to update DOM
    await nextTick()

    console.log(`[I18n] Changed language to: ${locale}`)
}

// Initial setup logic (optional, to ensure headers/html attrs are set on load)
// It's good practice to run this once on initialization to sync everything
const currentLang = getLanguage()
document.querySelector('html')?.setAttribute('lang', currentLang)
axios.defaults.headers.common['Accept-Language'] = currentLang

export default i18n
