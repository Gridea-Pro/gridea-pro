/**
 * Handles content overflow detection and expansion toggling.
 */
import { ref, nextTick, type Ref } from 'vue'

export function useContentOverflow(threshold = 240) {
    const isExpanded = ref(false)
    const isOverflowing = ref(false)

    // We need to keep track of the element to check scrollHeight vs threshold
    const contentRef = ref<HTMLElement | null>(null)

    const checkOverflow = async () => {
        await nextTick()
        if (!contentRef.value) return

        // If we're already checking or in a state where we might misread, be careful.
        // However, usually we just check if scrollHeight > threshold.
        if (contentRef.value.scrollHeight > threshold) {
            isOverflowing.value = true
        } else {
            isOverflowing.value = false
        }
    }

    const toggleExpand = () => {
        isExpanded.value = !isExpanded.value
    }

    return {
        isExpanded,
        isOverflowing,
        contentRef,
        checkOverflow,
        toggleExpand
    }
}
