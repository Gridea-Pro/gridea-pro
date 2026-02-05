import { EventsEmit } from 'wailsjs/runtime'

export default function ga(eventCategory: string, eventAction: string, eventLabel: string) {
  EventsEmit('ga', {
    eventCategory,
    eventAction,
    eventLabel,
  })
}
