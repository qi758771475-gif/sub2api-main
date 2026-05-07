<template>
  <AppLayout>
    <div class="shop-page-layout -m-4 md:-m-6 lg:-m-8">
      <div class="flex-1 min-h-0 overflow-hidden">
        <div v-if="loading" class="flex h-full items-center justify-center py-12">
          <div
            class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
          ></div>
        </div>

        <div
          v-else-if="!shopUrl"
          class="flex h-full items-center justify-center p-10 text-center"
        >
          <div class="max-w-md">
            <div
              class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
            >
              <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('purchaseShop.notConfigured') }}
            </h3>
          </div>
        </div>

        <iframe
          v-else
          :src="shopUrl"
          class="shop-embed-frame"
          allowfullscreen
        ></iframe>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)

const shopUrl = computed(() => {
  const url = appStore.cachedPublicSettings?.purchase_subscription_url
  if (!url) return ''
  return url.startsWith('http://') || url.startsWith('https://') ? url : ''
})

onMounted(async () => {
  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.shop-page-layout {
  @apply flex flex-col;
  height: calc(100vh - 64px);
}

.shop-embed-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  background: transparent;
  overscroll-behavior: contain;
}
</style>
