<template>
  <div class="emoji-card">
    <EmojiPicker :native="true" @select="handleSelect" />
  </div>
</template>

<script lang="ts" setup>
import EmojiPicker from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'

const emit = defineEmits(['select'])

const handleSelect = (emojiData: any) => {
  emit('select', emojiData.i)
}
</script>

<style lang="less" scoped>
.emoji-card {
  margin: -14px -17px;
  
  :deep(.v3-emoji-picker) {
    border: none;
    box-shadow: none;
    width: 100%;
  }
}
</style>
