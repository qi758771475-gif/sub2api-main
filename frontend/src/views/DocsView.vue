<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0f]">
    <!-- Top-right controls -->
    <div class="fixed top-4 right-4 z-50 flex gap-2">
      <button class="flex items-center gap-1.5 px-3 py-1.5 bg-white dark:bg-slate-800 border border-slate-300 dark:border-cyan-500/30 rounded-lg text-sm text-slate-700 dark:text-slate-200 shadow-sm hover:border-blue-500 dark:hover:border-cyan-400 transition-colors" @click="toggleLang" :title="locale === 'zh' ? 'Switch to English' : '切换到中文'">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m10.5 21 5.25-11.25L21 21m-9-3h7.5M3 5.621a48.474 48.474 0 016-.371m0 0c1.12 0 2.233.038 3.334.114M9 5.25V3m3.334 2.364C11.176 10.658 7.69 15.08 3 17.502m9.334-12.138c.896.061 1.784.154 2.666.282m0 0c1.096.173 2.164.396 3.2.672"/></svg>
        <span class="text-xs font-medium">{{ locale === 'zh' ? 'EN' : '中文' }}</span>
      </button>
      <button class="flex items-center gap-1.5 px-3 py-1.5 bg-white dark:bg-slate-800 border border-slate-300 dark:border-cyan-500/30 rounded-lg text-sm text-slate-700 dark:text-slate-200 shadow-sm hover:border-blue-500 dark:hover:border-cyan-400 transition-colors" @click="toggleDark">
        <svg v-if="isDark" class="w-4 h-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z"/></svg>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z"/></svg>
      </button>
    </div>

    <!-- Header -->
    <div class="bg-white dark:bg-[#0d0d1a] border-b border-slate-200 dark:border-cyan-500/20 relative overflow-hidden">
      <div class="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-cyan-400 to-purple-500 hidden dark:block"></div>
      <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <button
          @click="$router.back()"
          class="inline-flex items-center gap-1.5 mb-6 px-3 py-1.5 bg-white dark:bg-slate-800 border border-slate-300 dark:border-cyan-500/30 rounded-lg text-sm text-slate-600 dark:text-slate-300 shadow-sm hover:border-blue-500 dark:hover:border-cyan-400 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
          </svg>
          {{ locale === 'zh' ? '返回' : 'Back' }}
        </button>
        <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100">
          {{ t('docsGuide.title') }}
        </h1>
        <p class="mt-2 text-lg text-slate-500 dark:text-slate-400">
          {{ t('docsGuide.subtitle') }}
        </p>
      </div>
    </div>

    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <!-- Overview -->
      <section class="mb-12">
        <h2 class="flex items-center gap-2 text-xl font-semibold text-slate-800 dark:text-slate-100 mb-4">
          <span class="w-2 h-2 rounded-full bg-blue-500"></span>
          {{ t('docsGuide.overview') }}
        </h2>
        <div class="bg-white dark:bg-[#0d0d1a] border border-slate-200 dark:border-cyan-500/15 rounded-xl p-6">
          <p class="text-slate-600 dark:text-slate-300 leading-relaxed">
            {{ t('docsGuide.overviewText') }}
          </p>
        </div>
      </section>

      <!-- Quick Start -->
      <section class="mb-12">
        <h2 class="flex items-center gap-2 text-xl font-semibold text-slate-800 dark:text-slate-100 mb-4">
          <span class="w-2 h-2 rounded-full bg-green-500 dark:shadow-[0_0_8px_#22c55e]"></span>
          {{ t('docsGuide.quickStart') }}
        </h2>
        <div class="bg-white dark:bg-[#0d0d1a] border border-slate-200 dark:border-cyan-500/15 rounded-xl p-7">
          <div class="grid grid-cols-[1fr_auto_1fr] items-center gap-3">
            <div class="flex flex-col items-center gap-2.5 p-5 bg-slate-50 dark:bg-slate-800 border border-dashed border-slate-300 dark:border-slate-600 rounded-xl text-center">
              <span class="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-600 text-slate-500 dark:text-slate-200 text-sm font-bold flex items-center justify-center">1</span>
              <span class="text-sm text-slate-600 dark:text-slate-200 font-medium">{{ t('docsGuide.step1Text') }}</span>
              <div class="flex gap-2 flex-wrap justify-center">
                <router-link to="/login" class="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-sky-500 hover:bg-sky-600 text-white rounded-lg text-xs font-medium transition-colors">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9"/></svg>
                  {{ t('docsGuide.step1Login') }}
                </router-link>
                <router-link to="/register" class="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-green-500 hover:bg-green-600 text-white rounded-lg text-xs font-medium transition-colors">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM4 19.235v-.11a6.375 6.375 0 0112.75 0v.109A12.318 12.318 0 0110.374 21c-2.331 0-4.512-.645-6.374-1.766z"/></svg>
                  {{ t('docsGuide.step1Register') }}
                </router-link>
              </div>
            </div>
            <div class="text-slate-300 dark:text-slate-600 hidden sm:block">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"/></svg>
            </div>
            <div class="flex flex-col items-center gap-2.5 p-5 bg-slate-50 dark:bg-slate-800 border border-dashed border-slate-300 dark:border-slate-600 rounded-xl text-center">
              <span class="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-600 text-slate-500 dark:text-slate-200 text-sm font-bold flex items-center justify-center">2</span>
              <span class="text-sm text-slate-600 dark:text-slate-200 font-medium">{{ t('docsGuide.step2Text') }}</span>
              <router-link to="/keys" class="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-sky-500 hover:bg-sky-600 text-white rounded-lg text-xs font-medium transition-colors">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z"/></svg>
                {{ t('docsGuide.step2Link') }}
              </router-link>
            </div>
          </div>
          <div class="flex items-center gap-2.5 py-4 my-2 border-y border-slate-200 dark:border-slate-600">
            <span class="w-7 h-7 rounded-full bg-green-500 text-white text-sm font-bold flex items-center justify-center">3</span>
            <span class="text-sm text-slate-600 dark:text-slate-200 font-medium">{{ t('docsGuide.step3Text') }}</span>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="bg-slate-900 dark:bg-[#06060f] border border-slate-700 dark:border-cyan-500/20 rounded-xl p-4">
              <div class="flex items-center gap-1.5 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418"/></svg>
                {{ t('docsGuide.apiBaseUrl') }}
              </div>
              <code class="text-sm font-mono text-slate-200 break-all select-all">{{ baseUrlExample }}</code>
            </div>
            <div class="bg-slate-900 dark:bg-[#06060f] border border-slate-700 dark:border-cyan-500/20 rounded-xl p-4">
              <div class="flex items-center gap-1.5 mb-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z"/></svg>
                {{ t('docsGuide.apiKey') }}
              </div>
              <code class="text-sm font-mono text-green-400 break-all select-all">sk-your-api-key-here</code>
            </div>
          </div>
        </div>
      </section>

      <!-- Tool Configuration -->
      <section class="mb-12">
        <h2 class="flex items-center gap-2 text-xl font-semibold text-slate-800 dark:text-slate-100 mb-1">
          <span class="w-2 h-2 rounded-full bg-cyan-500 dark:shadow-[0_0_8px_#06b6d4]"></span>
          {{ locale === 'zh' ? '工具配置' : 'Tool Setup' }}
        </h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mb-6">{{ locale === 'zh' ? '点击展开对应工具的配置说明' : 'Click to expand configuration for each tool' }}</p>

        <div class="space-y-3">
          <!-- CC Switch -->
          <ToolSection :title="t('docsGuide.tools.ccSwitch.name')" :description="t('docsGuide.tools.ccSwitch.desc')">
            <a :href="t('docsGuide.tools.ccSwitch.downloadUrl')" target="_blank" rel="noopener" class="inline-flex items-center gap-2 mb-3 px-3 py-2 bg-slate-800 dark:bg-slate-700 border border-slate-700 dark:border-cyan-500/30 text-white rounded-lg hover:bg-slate-700 dark:hover:border-cyan-400 dark:hover:shadow-[0_0_12px_rgba(6,182,212,.3)] transition-all text-sm font-medium">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              {{ t('docsGuide.tools.ccSwitch.download') }}: GitHub Releases
            </a>
            <div class="flex items-start gap-2 mb-4 p-3 bg-green-50 dark:bg-green-500/8 border border-green-200 dark:border-green-500/30 rounded-lg text-sm text-green-800 dark:text-green-300">
              <span>💡</span>
              <span>{{ locale === 'zh' ? '推荐：登录后在 API 密钥页面点"导入到 CCS"一键完成配置' : 'Tip: Login → API Keys → click "Import to CCS" for one-click setup' }}</span>
            </div>
            <CodeBlock :code="ccSwitchSetup" language="bash" />
            <p class="mt-4 mb-2 text-sm text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.ccSwitch.manualSetup') }}</p>
            <CodeBlock :code="ccSwitchManualSetup" language="json" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.ccSwitch.note') }}</p>
          </ToolSection>

          <ToolSection title="Claude Code" :description="t('docsGuide.tools.claudeCode.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.claudeCode.setup') }}</p>
            <CodeBlock :code="claudeCodeSetup" language="bash" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.claudeCode.note') }}</p>
          </ToolSection>

          <ToolSection title="OpenCode" :description="t('docsGuide.tools.opencode.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.opencode.setup') }}</p>
            <CodeBlock :code="openCodeSetup" language="json" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.opencode.note') }}</p>
          </ToolSection>

          <ToolSection title="Codex" :description="t('docsGuide.tools.codex.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.codex.setup') }}</p>
            <CodeBlock :code="codexSetup" language="bash" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.codex.note') }}</p>
          </ToolSection>

          <ToolSection title="Cursor IDE" :description="t('docsGuide.tools.cursor.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.cursor.setup') }}</p>
            <ol class="step-list">
              <li>{{ t('docsGuide.tools.cursor.step1') }}</li>
              <li>{{ t('docsGuide.tools.cursor.step2') }}</li>
              <li>{{ t('docsGuide.tools.cursor.step3') }}</li>
            </ol>
            <CodeBlock :code="cursorSetup" language="bash" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.cursor.note') }}</p>
          </ToolSection>

          <ToolSection title="Continue.dev" :description="t('docsGuide.tools.continue.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.continue.setup') }}</p>
            <CodeBlock :code="continueSetup" language="json" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.continue.note') }}</p>
          </ToolSection>

          <ToolSection title="Aider" :description="t('docsGuide.tools.aider.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.aider.setup') }}</p>
            <CodeBlock :code="aiderSetup" language="bash" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.aider.note') }}</p>
          </ToolSection>

          <ToolSection :title="t('docsGuide.tools.terminal.name')" :description="t('docsGuide.tools.terminal.desc')">
            <h4 class="text-sm font-medium text-cyan-400 mb-2">{{ t('docsGuide.tools.terminal.chatCompletions') }}</h4>
            <CodeBlock :code="curlChatSetup" language="bash" />
            <h4 class="text-sm font-medium text-green-400 mb-2 mt-4">{{ t('docsGuide.tools.terminal.responses') }}</h4>
            <CodeBlock :code="curlResponsesSetup" language="bash" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.terminal.note') }}</p>
          </ToolSection>

          <ToolSection :title="t('docsGuide.tools.python.name')" :description="t('docsGuide.tools.python.desc')">
            <p class="text-slate-600 dark:text-slate-300 mb-3">{{ t('docsGuide.tools.python.setup') }}</p>
            <CodeBlock :code="pythonSetup" language="bash" />
            <CodeBlock :code="pythonCodeSetup" language="python" class="mt-3" />
            <p class="mt-3 text-xs text-slate-500 dark:text-slate-400">{{ t('docsGuide.tools.python.note') }}</p>
          </ToolSection>
        </div>
      </section>

      <!-- Platform Notes -->
      <section class="mb-12">
        <h2 class="flex items-center gap-2 text-xl font-semibold text-slate-800 dark:text-slate-100 mb-1">
          <span class="w-2 h-2 rounded-full bg-purple-500 dark:shadow-[0_0_8px_#a855f7]"></span>
          {{ t('docsGuide.platformSpecific.title') }}
        </h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mb-6">{{ locale === 'zh' ? '不同操作系统设置环境变量的方法' : 'How to set environment variables on each OS' }}</p>
        <div class="space-y-4">
          <div v-for="platform in platforms" :key="platform.key" class="bg-white dark:bg-[#0d0d1a] border border-slate-200 dark:border-cyan-500/10 hover:dark:border-cyan-500/30 rounded-xl p-6 transition-colors">
            <div class="flex items-center gap-3 mb-3">
              <span class="text-2xl">{{ platform.icon }}</span>
              <h3 class="text-lg font-semibold" :class="platform.color">{{ platform.title }}</h3>
            </div>
            <p class="text-slate-600 dark:text-slate-300 text-sm mb-3">{{ platform.desc }}</p>
            <CodeBlock :code="platform.code" language="bash" />
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ToolSection from '@/components/docs/ToolSection.vue'
import CodeBlock from '@/components/docs/CodeBlock.vue'
import { setLocale } from '@/i18n'

const { t, locale } = useI18n()

// Dark mode
const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleDark() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Language
function toggleLang() {
  const next = locale.value === 'zh' ? 'en' : 'zh'
  setLocale(next)
}

const baseUrlExample = 'https://oneapi.cam/v1'
const baseUrl = baseUrlExample

// ── Code snippets ──

const ccSwitchSetup = computed(() => `# ${locale.value === 'zh' ? '方式一：网页端一键导入（推荐）' : 'Method 1: One-Click Import (Recommended)'}
# ${locale.value === 'zh' ? '登录 → API 密钥 → 点击「导入到 CCS」' : 'Login → API Keys → Click "Import to CCS"'}

# ${locale.value === 'zh' ? '方式二：手动设置环境变量' : 'Method 2: Set environment variables'}
export OPENAI_BASE_URL="${baseUrl}"
export OPENAI_API_KEY="sk-your-api-key-here"`)

const ccSwitchManualSetup = computed(() => `{
  "providers": {
    "oneapi-claude":  { "type": "anthropic", "base_url": "${baseUrl}", "api_key": "sk-..." },
    "oneapi-codex":   { "type": "openai",    "base_url": "${baseUrl}", "api_key": "sk-..." },
    "oneapi-gemini":  { "type": "gemini",    "base_url": "${baseUrl}", "api_key": "sk-..." }
  }
}`)

const claudeCodeSetup = computed(() => `# ${locale.value === 'zh' ? '推荐：添加到 ~/.zshrc 或 ~/.bashrc' : 'Add to ~/.zshrc or ~/.bashrc'}
export ANTHROPIC_BASE_URL="${baseUrl}"
export ANTHROPIC_API_KEY="sk-your-api-key-here"

# ${locale.value === 'zh' ? '或者：在项目目录创建 .env 文件' : 'Or: .env in project root'}
ANTHROPIC_BASE_URL=${baseUrl}
ANTHROPIC_API_KEY=sk-your-api-key-here`)

const openCodeSetup = computed(() => `{
  "providers": {
    "oneapi": {
      "base_url": "${baseUrl}",
      "api_key": "sk-your-api-key-here"
    }
  },
  "default_provider": "oneapi"
}`)

const codexSetup = computed(() => `# ${locale.value === 'zh' ? '环境变量（推荐）' : 'Env vars (recommended)'}
export OPENAI_BASE_URL="${baseUrl}"
export OPENAI_API_KEY="sk-your-api-key-here"

# ${locale.value === 'zh' ? '或命令' : 'Or via CLI'}
codex config set openai_base_url "${baseUrl}"
codex config set openai_api_key "sk-your-api-key-here"`)

const cursorSetup = computed(() => `# ${locale.value === 'zh' ? '在 Cursor → Settings → Models 中添加模型名' : 'In Cursor → Settings → Models, add:'}
gpt-4o
claude-sonnet-4-5
claude-opus-4-7
gemini-2.5-pro`)

const continueSetup = computed(() => `{
  "models": [{
    "title": "OneApi",
    "provider": "openai",
    "model": "gpt-4o",
    "apiBase": "${baseUrl}",
    "apiKey": "sk-your-api-key-here"
  }]
}`)

const aiderSetup = computed(() => `export OPENAI_API_BASE="${baseUrl}"
export OPENAI_API_KEY="sk-your-api-key-here"
aider --model openai/gpt-4o`)

const curlChatSetup = computed(() => `curl ${baseUrl}/chat/completions \\
  -H "Authorization: Bearer sk-your-api-key-here" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'`)

const curlResponsesSetup = computed(() => `curl ${baseUrl}/responses \\
  -H "Authorization: Bearer sk-your-api-key-here" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4o","input":"Hello","stream":true}'`)

const pythonSetup = 'pip install openai'

const pythonCodeSetup = computed(() => `from openai import OpenAI

client = OpenAI(
    base_url="${baseUrl}",
    api_key="sk-your-api-key-here",
)
r = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role":"user","content":"Hello"}],
)
print(r.choices[0].message.content)`)

const platforms = computed(() => [
  {
    key: 'windows', icon: '🪟', color: 'text-blue-400',
    title: t('docsGuide.platformSpecific.windows.title'),
    desc: t('docsGuide.platformSpecific.windows.desc'),
    code: `# ${locale.value === 'zh' ? 'CMD（管理员）' : 'CMD (Admin)'}
setx ANTHROPIC_BASE_URL "${baseUrl}"
setx ANTHROPIC_API_KEY "sk-your-api-key-here"

# PowerShell
[Environment]::SetEnvironmentVariable('ANTHROPIC_BASE_URL', '${baseUrl}', 'User')
[Environment]::SetEnvironmentVariable('ANTHROPIC_API_KEY', 'sk-your-api-key-here', 'User')`,
  },
  {
    key: 'macos', icon: '🍎', color: 'dark:text-slate-300 text-slate-700',
    title: t('docsGuide.platformSpecific.macos.title'),
    desc: t('docsGuide.platformSpecific.macos.desc'),
    code: `# ${locale.value === 'zh' ? '添加到 ~/.zshrc 并生效' : 'Add to ~/.zshrc & source'}
echo 'export ANTHROPIC_BASE_URL="${baseUrl}"' >> ~/.zshrc
echo 'export ANTHROPIC_API_KEY="sk-your-api-key-here"' >> ~/.zshrc
source ~/.zshrc`,
  },
  {
    key: 'linux', icon: '🐧', color: 'text-yellow-400',
    title: t('docsGuide.platformSpecific.linux.title'),
    desc: t('docsGuide.platformSpecific.linux.desc'),
    code: `# ${locale.value === 'zh' ? '添加到 ~/.bashrc 并生效' : 'Add to ~/.bashrc & source'}
echo 'export ANTHROPIC_BASE_URL="${baseUrl}"' >> ~/.bashrc
echo 'export ANTHROPIC_API_KEY="sk-your-api-key-here"' >> ~/.bashrc
source ~/.bashrc`,
  },
])
</script>

<style scoped>
/* ── Step list ── */
.step-list { list-style: none; counter-reset: step; padding: 0; margin: 0 0 1rem; }
.step-list li {
  counter-increment: step; padding: .5rem 0 .5rem 2rem; position: relative;
  font-size: .875rem;
}
.step-list li::before {
  content: counter(step);
  position: absolute; left: 0; top: .5rem;
  width: 1.25rem; height: 1.25rem;
  border-radius: 50%;
  font-size: .75rem; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  background: #3b82f6; color: #fff;
}
</style>
