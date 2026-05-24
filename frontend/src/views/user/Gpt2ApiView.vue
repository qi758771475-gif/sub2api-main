<template>
  <div class="flex h-screen items-center justify-center bg-gray-50 dark:bg-dark-900">
    <div class="text-center">
      <p class="mb-4 text-lg text-gray-600 dark:text-gray-400">
        {{ locale === 'zh' ? '正在跳转到 GPT2API...' : 'Redirecting to GPT2API...' }}
      </p>
      <a :href="gpt2apiUrl" class="btn btn-primary" target="_blank">
        {{ locale === 'zh' ? '如果未自动跳转，请点击这里' : 'Click here if not redirected' }}
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()
const token = localStorage.getItem('auth_token')
const gpt2apiUrl = computed(() => `https://gptimage.oneapi.cam?token=${token || ''}`)

onMounted(() => {
  window.open(gpt2apiUrl.value, '_blank')
})
</script>
