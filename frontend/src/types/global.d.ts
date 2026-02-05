export interface ElectronAPI {
  send: (channel: string, data?: unknown) => void
  once: (channel: string, callback: (event: unknown, ...args: unknown[]) => void) => void
  removeAllListeners: (channel: string) => void
  getLocale?: () => string
}

declare global {
  interface Window {
    electronAPI: ElectronAPI
  }
}

export {}
