<script setup lang="ts">
import { onMounted } from 'vue'

onMounted(() => {
  const token = localStorage.getItem('auth_token')
  window.location.href = `https://gptimage.oneapi.cam?token=${token || ''}`
})
</script>
