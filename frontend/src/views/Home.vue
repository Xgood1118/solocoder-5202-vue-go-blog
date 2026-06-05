<template>
  <div class="home">
    <div class="main-column">
      <div class="filter-bar">
        <select v-model="filters.category" @change="fetchArticles">
          <option value="">全部分类</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </select>
        <select v-model="filters.sortBy" @change="fetchArticles">
          <option value="time">最新发布</option>
          <option value="views">最多浏览</option>
        </select>
        <input 
          type="text" 
          v-model="authorFilter" 
          placeholder="按作者筛选..." 
          @keyup.enter="filterByAuthor"
        >
      </div>

      <div v-if="articles.length === 0" class="empty">
        暂无文章，<router-link to="/edit">去写第一篇</router-link>
      </div>

      <div v-for="article in articles" :key="article.id" class="article-card card">
        <div class="article-header">
          <h3>
            <router-link :to="'/article/' + article.id">
              <span v-if="article.isPinned" class="tag pinned">置顶</span>
              <span v-html="highlightText(article.title)"></span>
            </router-link>
          </h3>
          <div class="article-meta">
            <span class="author" @click="filterAuthor(article.author)">{{ article.author }}</span>
            <span>{{ formatDate(article.createdAt) }}</span>
            <span>👁️ {{ article.viewCount }}</span>
            <span>❤️ {{ article.likeCount }}</span>
            <span>💬 {{ article.commentCount }}</span>
            <span>{{ article.wordCount }}字</span>
            <span>{{ article.readTime }}分钟阅读</span>
          </div>
        </div>
        <div class="article-tags">
          <span 
            v-for="tag in article.tags" 
            :key="tag" 
            class="tag"
            @click="filterByTag(tag)"
          >{{ tag }}</span>
        </div>
      </div>

      <div class="pagination" v-if="total > pageSize">
        <button @click="prevPage" :disabled="page === 1">上一页</button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button @click="nextPage" :disabled="page === totalPages">下一页</button>
      </div>
    </div>

    <aside class="sidebar">
      <div class="card">
        <h3>热门标签云</h3>
        <div class="tag-cloud">
          <span 
            v-for="tag in tagCloud" 
            :key="tag.name"
            :style="getTagStyle(tag)"
            @click="filterByTag(tag.name)"
          >{{ tag.name }}</span>
        </div>
      </div>

      <div class="card">
        <h3>近期活跃作者</h3>
        <div class="author-list">
          <div v-for="(author, index) in activeAuthors" :key="author.name" class="author-item">
            <span class="rank">{{ index + 1 }}</span>
            <span class="name" @click="filterAuthor(author.name)">{{ author.name }}</span>
            <span class="stats">{{ author.posts }}篇 / {{ author.views }}阅</span>
          </div>
        </div>
      </div>

      <div class="card">
        <h3>归档</h3>
        <div class="archive-list">
          <div v-for="item in archive" :key="item.month" class="archive-item">
            <span @click="filterByMonth(item.month)">{{ item.month }}</span>
            <span class="count">({{ item.articles.length }})</span>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getArticles, getCategories, getTagCloud, getActiveAuthors, getArchive } from '../api'

const route = useRoute()
const articles = ref([])
const categories = ref([])
const tagCloud = ref([])
const activeAuthors = ref([])
const archive = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filters = ref({
  category: '',
  tag: '',
  search: '',
  sortBy: 'time',
  author: '',
  month: ''
})
const authorFilter = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const fetchArticles = async () => {
  const params = {
    page: page.value,
    size: pageSize.value,
    tag: filters.value.tag,
    category: filters.value.category,
    search: filters.value.search,
    author: filters.value.author,
    sortBy: filters.value.sortBy
  }
  const res = await getArticles(params)
  articles.value = res.data.list
  total.value = res.data.total
}

const fetchCategories = async () => {
  const res = await getCategories()
  categories.value = res.data
}

const fetchTagCloud = async () => {
  const res = await getTagCloud()
  tagCloud.value = res.data
}

const fetchActiveAuthors = async () => {
  const res = await getActiveAuthors()
  activeAuthors.value = res.data
}

const fetchArchive = async () => {
  const res = await getArchive()
  archive.value = res.data
}

const filterByTag = (tag) => {
  filters.value.tag = tag
  filters.value.month = ''
  filters.value.author = ''
  page.value = 1
  fetchArticles()
}

const filterByAuthor = () => {
  filters.value.author = authorFilter.value
  filters.value.tag = ''
  filters.value.month = ''
  page.value = 1
  fetchArticles()
}

const filterAuthor = (author) => {
  filters.value.author = author
  authorFilter.value = author
  filters.value.tag = ''
  filters.value.month = ''
  page.value = 1
  fetchArticles()
}

const filterByMonth = (month) => {
  const monthArticles = archive.value.find(a => a.month === month)?.articles || []
  articles.value = monthArticles
  total.value = monthArticles.length
  filters.value.month = month
  filters.value.tag = ''
  filters.value.author = ''
}

const prevPage = () => {
  if (page.value > 1) {
    page.value--
    fetchArticles()
  }
}

const nextPage = () => {
  if (page.value < totalPages.value) {
    page.value++
    fetchArticles()
  }
}

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const highlightText = (text) => {
  if (!filters.value.search) return text
  const regex = new RegExp(`(${filters.value.search})`, 'gi')
  return text.replace(regex, '<span class="highlight">$1</span>')
}

const getTagStyle = (tag) => {
  const maxCount = Math.max(...tagCloud.value.map(t => t.count))
  const size = 12 + (tag.count / maxCount) * 12
  const opacity = 0.5 + (tag.count / maxCount) * 0.5
  return {
    fontSize: size + 'px',
    opacity: opacity,
    cursor: 'pointer',
    margin: '4px',
    display: 'inline-block'
  }
}

onMounted(() => {
  filters.value.search = route.query.search || ''
  fetchArticles()
  fetchCategories()
  fetchTagCloud()
  fetchActiveAuthors()
  fetchArchive()
})
</script>

<style scoped>
.home {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 20px;
}

.filter-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.filter-bar select, .filter-bar input {
  flex: 1;
  max-width: 200px;
}

.article-card h3 {
  margin: 0 0 10px;
}

.article-card h3 a {
  color: var(--text-color);
  text-decoration: none;
}

.article-card h3 a:hover {
  color: var(--primary-color);
}

.article-meta {
  display: flex;
  gap: 15px;
  font-size: 12px;
  color: #888;
  margin-bottom: 10px;
}

.article-meta .author {
  cursor: pointer;
  color: var(--primary-color);
}

.article-tags {
  margin-top: 10px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 15px;
  margin-top: 20px;
}

.empty {
  text-align: center;
  padding: 60px;
  color: #888;
}

.sidebar .card h3 {
  margin-bottom: 15px;
  font-size: 16px;
}

.tag-cloud {
  line-height: 2;
}

.author-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.author-item:last-child {
  border-bottom: none;
}

.rank {
  width: 20px;
  height: 20px;
  background: var(--primary-color);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.name {
  flex: 1;
  cursor: pointer;
}

.name:hover {
  color: var(--primary-color);
}

.stats {
  font-size: 12px;
  color: #888;
}

.archive-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  cursor: pointer;
}

.archive-item:hover {
  color: var(--primary-color);
}

.count {
  color: #888;
}
</style>
