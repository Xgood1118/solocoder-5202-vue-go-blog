<template>
  <div :class="{ dark: isDark }">
    <div class="progress-bar" :style="{ width: progress + '%' }" v-if="showProgress"></div>
    <header class="header">
      <div class="container header-content">
        <router-link to="/" class="logo">技术博客</router-link>
        <nav>
          <router-link to="/">首页</router-link>
          <router-link to="/edit">写文章</router-link>
          <router-link to="/categories">分类</router-link>
          <a href="/rss.xml" target="_blank">RSS</a>
        </nav>
        <div class="header-right">
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="搜索文章..." 
            @keyup.enter="goSearch"
            class="search-input"
          >
          <button @click="toggleDark" class="theme-btn">{{ isDark ? '☀️' : '🌙' }}</button>
        </div>
      </div>
    </header>
    <main class="container main-content">
      <router-view />
    </main>
    <footer class="footer">
      <div class="container">
        <p>© 2024 内部技术博客 - 团队知识分享平台</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const isDark = ref(localStorage.getItem('darkMode') === 'true')
const searchQuery = ref('')
const progress = ref(0)
const showProgress = ref(false)

const toggleDark = () => {
  isDark.value = !isDark.value
  localStorage.setItem('darkMode', isDark.value)
  document.documentElement.classList.toggle('dark', isDark.value)
}

const goSearch = () => {
  if (searchQuery.value.trim()) {
    router.push(`/?search=${encodeURIComponent(searchQuery.value)}`)
  }
}

const updateProgress = () => {
  const scrollTop = window.scrollY
  const docHeight = document.documentElement.scrollHeight - window.innerHeight
  progress.value = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0
}

onMounted(() => {
  if (isDark.value) {
    document.documentElement.classList.add('dark')
  }
  window.addEventListener('scroll', updateProgress)
})

onUnmounted(() => {
  window.removeEventListener('scroll', updateProgress)
})

watch(() => route.path, (path) => {
  showProgress.value = path.startsWith('/article/')
  progress.value = 0
})
</script>

<style scoped>
.header {
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: var(--text-color) !important;
  text-decoration: none !important;
}

nav a {
  margin-right: 20px;
  color: var(--text-color);
  text-decoration: none;
}

nav a:hover {
  color: var(--primary-color);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input {
  width: 200px;
}

.theme-btn {
  background: none;
  font-size: 18px;
  padding: 5px 10px;
}

.main-content {
  min-height: calc(100vh - 140px);
  padding-top: 20px;
}

.footer {
  background: var(--card-bg);
  border-top: 1px solid var(--border-color);
  padding: 20px 0;
  text-align: center;
  margin-top: 40px;
}
</style>
