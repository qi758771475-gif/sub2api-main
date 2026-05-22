<template>
  <iframe :src="gpt2apiUrl" class="h-screen w-full border-0" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const gpt2apiUrl = computed(() => {
  const base = import.meta.env.VITE_GPT2API_URL || 'http://127.0.0.1:17200'
  const token = localStorage.getItem('auth_token')
  return `${base}?token=${token || ''}`
})
</script>
