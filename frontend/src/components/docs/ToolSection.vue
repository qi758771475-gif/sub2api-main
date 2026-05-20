<template>
  <div class="bg-white dark:bg-[#0d0d1a] border border-slate-200 dark:border-cyan-500/15 hover:dark:border-cyan-500/30 rounded-xl overflow-hidden transition-colors">
    <div class="px-6 pt-5">
      <h3 class="flex items-center gap-2 text-base font-semibold text-slate-800 dark:text-slate-100">
        <span class="text-cyan-500 font-mono text-sm dark:text-[#22d3ee]">&gt;</span>
        {{ title }}
      </h3>
      <p class="mt-1 text-sm text-slate-500 dark:text-slate-400 leading-relaxed">{{ description }}</p>
    </div>
    <div class="px-6 pb-5 pt-4">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>
