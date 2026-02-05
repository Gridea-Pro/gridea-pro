import { safeEventsOn } from './wailsRuntime'
import { toast as sonner } from 'vue-sonner'

export const toast = {
  success: (message: string, duration = 2000) => sonner.success(message, { duration }),
  error: (message: string, duration = 2000) => sonner.error(message, { duration }),
  warning: (message: string, duration = 2000) => sonner.warning(message, { duration }),
  info: (message: string, duration = 2000) => sonner.info(message, { duration }),
}

export const setupToastListeners = () => {
  safeEventsOn('app:toast', (data: any) => {
    if (data && data.message) {
      const type = data.type || 'info'
      const duration = data.duration || 2000

      switch (type) {
        case 'success':
          toast.success(data.message, duration)
          break
        case 'error':
          toast.error(data.message, duration)
          break
        case 'warning':
          toast.warning(data.message, duration)
          break
        default:
          toast.info(data.message, duration)
      }
    }
  })
}
