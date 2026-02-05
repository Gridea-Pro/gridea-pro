<script lang="ts" setup>
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

onMounted(() => {
  setTimeout(() => {
    const { redirect } = route.query
    if (redirect) {
      const path = Array.isArray(redirect) ? redirect[0] : redirect
      if (path) {
        router.push(path)
      }
    }
  }, 1000)
})
</script>

<template>
  <div class="loading-page">
    <img class="img" src="@/assets/images/developing_code.svg" width="160px">
    {{ $t('inConfig') }}...
  </div>
</template>


<style lang="less" scoped>
.loading-page {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
  min-height: 100vh;
  .img {
    margin-bottom: 16px;
  }
}
</style>
