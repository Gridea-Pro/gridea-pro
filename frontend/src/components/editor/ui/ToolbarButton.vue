<template>
  <button
    type="button"
    class="tb-btn"
    :class="{ 'is-active': active }"
    :disabled="disabled"
    :title="title"
    @mousedown.prevent
    @click="emit('click')"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{ title?: string; active?: boolean; disabled?: boolean }>()
const emit = defineEmits<{ click: [] }>()
</script>

<style scoped>
.tb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 6px;
  color: var(--editor-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}
.tb-btn :deep(svg) {
  width: 17px;
  height: 17px;
}
.tb-btn:hover:not(:disabled) {
  background: var(--editor-hover);
  color: var(--editor-fg);
}
.tb-btn.is-active {
  background: color-mix(in srgb, var(--editor-accent) 16%, transparent);
  color: var(--editor-accent);
}
.tb-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
