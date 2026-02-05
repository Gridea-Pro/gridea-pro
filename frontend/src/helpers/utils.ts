import markdown from './markdown'

interface ThemeConfigItem {
  name: string
  type: string
  arrayItems?: Array<{ name: string; type: string }>
}

interface ConfigObject {
  [key: string]: unknown
}

export function formatYamlString(value: string): string {
  if (typeof value !== 'string') {
    return String(value)
  }
  return value.replace(/'/g, "''")
}

function renderMarkdownValue(value: string): string {
  return markdown.render(value)
}

function processArrayConfigItem(
  configValue: unknown[],
  configItem: ThemeConfigItem,
  currentThemeConfig: ThemeConfigItem[]
): void {
  const foundConfigItem = currentThemeConfig.find((i) => i.name === configItem.name)
  if (!foundConfigItem?.arrayItems) return

  for (const arrItem of configValue) {
    if (typeof arrItem !== 'object' || arrItem === null) continue

    const arrayItemObj = arrItem as Record<string, unknown>
    const arrayItemKeys = Object.keys(arrayItemObj)

    for (const key of arrayItemKeys) {
      const foundMarkdownField = foundConfigItem.arrayItems.find(
        (i) => i.name === key && i.type === 'markdown'
      )

      if (foundMarkdownField) {
        const fieldValue = arrayItemObj[key]
        if (fieldValue && typeof fieldValue === 'string') {
          arrayItemObj[key] = renderMarkdownValue(fieldValue)
        }
      }
    }
  }
}

export function formatThemeCustomConfigToRender(
  config: ConfigObject,
  currentThemeConfig: ThemeConfigItem[]
): ConfigObject {
  if (!config || !currentThemeConfig || !Array.isArray(currentThemeConfig)) {
    return config
  }

  for (const configItem of currentThemeConfig) {
    const configValue = config[configItem.name]

    if (configItem.type === 'markdown' && typeof configValue === 'string') {
      config[configItem.name] = renderMarkdownValue(configValue)
    } else if (configItem.type === 'array' && Array.isArray(configValue)) {
      processArrayConfigItem(configValue, configItem, currentThemeConfig)
    }
  }

  return config
}
