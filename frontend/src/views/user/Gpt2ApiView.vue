<template>
  <iframe :src="gpt2apiUrl" class="h-screen w-full border-0" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const gpt2apiUrl = computed(() => {
  const token = localStorage.getItem('auth_token')
  return `https://gptimage.oneapi.cam?token=${token || ''}`
})
</script>
