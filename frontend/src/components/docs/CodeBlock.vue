<template>
  <div :class="className">
    <div class="flex items-center gap-3 px-4 py-2.5 bg-slate-800 dark:bg-slate-900 rounded-t-lg">
      <div class="flex gap-1.5">
        <span class="w-2.5 h-2.5 rounded-full bg-red-500" />
        <span class="w-2.5 h-2.5 rounded-full bg-yellow-500" />
        <span class="w-2.5 h-2.5 rounded-full bg-green-500" />
      </div>
      <span class="text-[10px] text-slate-500 uppercase tracking-wider flex-1">{{ language }}</span>
      <button type="button" class="flex items-center gap-1 text-xs text-slate-400 hover:text-slate-200 transition-colors" @click="copyCode">
        <svg v-if="!copied" class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
        <svg v-else class="w-3 h-3 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5"/></svg>
        {{ copied ? t('docsGuide.copied') : t('docsGuide.copy') }}
      </button>
    </div>
    <pre class="m-0 p-4 bg-slate-900 dark:bg-[#06060f] rounded-b-lg border-x border-b border-slate-700 dark:border-cyan-500/20 overflow-x-auto"><code class="text-[13px] font-mono text-slate-200 dark:text-cyan-100 leading-relaxed whitespace-pre">{{ code }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{ code: string; language: string; className?: string }>()

const copied = ref(false)

async function copyCode() {
  try {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = props.code
    ta.style.cssText = 'position:fixed;opacity:0'
    document.body.appendChild(ta)
    ta.select(); document.execCommand('copy'); document.body.removeChild(ta)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}
</script>
