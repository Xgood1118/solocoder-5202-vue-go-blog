<template>
  <div class="article-edit">
    <div class="card">
      <h2>{{ isEdit ? '编辑文章' : '写文章' }}</h2>
      
      <div class="form-group">
        <label>标题</label>
        <input type="text" v-model="form.title" placeholder="请输入文章标题">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>作者</label>
          <input type="text" v-model="form.author" placeholder="你的名字">
        </div>
        <div class="form-group">
          <label>分类</label>
          <select v-model="form.categoryId">
            <option value="">请选择分类</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
          <button class="small-btn" @click="showNewCategory = true">+ 新建分类</button>
        </div>
        <div class="form-group">
          <label>置顶</label>
          <input type="checkbox" v-model="form.isPinned">
        </div>
      </div>

      <div class="form-group">
        <label>标签（用逗号分隔）</label>
        <input type="text" v-model="tagsInput" placeholder="Vue, JavaScript, 前端">
      </div>

      <div class="editor-container">
        <div class="editor-header">
          <span>Markdown 编辑器</span>
          <span class="auto-save" v-if="lastSaved">自动保存于 {{ lastSaved }}</span>
        </div>
        <div class="editor-wrapper">
          <textarea 
            v-model="form.content" 
            class="editor"
            placeholder="开始写作，支持 Markdown 语法..."
            @input="onContentInput"
          ></textarea>
          <div class="preview">
            <div class="markdown-content" v-html="renderedContent"></div>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button @click="save" :disabled="saving">
          {{ saving ? '保存中...' : '保存文章' }}
        </button>
        <button class="secondary" @click="restoreDraft" v-if="hasDraft">
          恢复草稿
        </button>
        <router-link to="/">
          <button class="secondary">取消</button>
        </router-link>
      </div>

      <div class="word-count">
        字数：{{ wordCount }} | 预计阅读时间：{{ readTime }} 分钟
      </div>
    </div>

    <div class="modal" v-if="showNewCategory" @click.self="showNewCategory = false">
      <div class="modal-content card">
        <h3>新建分类</h3>
        <input type="text" v-model="newCategory.name" placeholder="分类名称">
        <input type="text" v-model="newCategory.slug" placeholder="分类标识 (slug)">
        <div class="modal-actions">
          <button @click="createCategory">创建</button>
          <button class="secondary" @click="showNewCategory = false">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import { 
  createArticle, updateArticle, getArticle, 
  getCategories, createCategory as apiCreateCategory 
} from '../api'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const categories = ref([])
const showNewCategory = ref(false)
const newCategory = ref({ name: '', slug: '' })
const lastSaved = ref('')
const autoSaveTimer = ref(null)

const form = ref({
  title: '',
  author: '',
  categoryId: '',
  content: '',
  isPinned: false,
  tags: []
})

const tagsInput = computed({
  get: () => form.value.tags.join(', '),
  set: (val) => form.value.tags = val.split(',').map(t => t.trim()).filter(Boolean)
})

const renderedContent = computed(() => marked(form.value.content || ''))
const wordCount = computed(() => form.value.content.replace(/\s/g, '').length)
const readTime = computed(() => Math.ceil(wordCount.value / 200) || 1)

const draftKey = computed(() => {
  const clientId = localStorage.getItem('clientId') || 'guest'
  const articleId = route.params.id || 'new'
  return `draft_${clientId}_${articleId}`
})

const hasDraft = computed(() => !!localStorage.getItem(draftKey.value))

const fetchCategories = async () => {
  const res = await getCategories()
  categories.value = res.data
}

const fetchArticle = async () => {
  if (!isEdit.value) return
  const res = await getArticle(route.params.id)
  form.value = {
    title: res.data.article.title,
    author: res.data.article.author,
    categoryId: res.data.article.categoryId,
    content: res.data.article.content,
    isPinned: res.data.article.isPinned,
    tags: res.data.article.tags
  }
}

const save = async () => {
  if (!form.value.title || !form.value.author || !form.value.content) {
    alert('请填写完整信息')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await updateArticle(route.params.id, form.value)
    } else {
      const res = await createArticle(form.value)
      localStorage.removeItem(draftKey.value)
      router.push('/article/' + res.data.id)
      return
    }
    localStorage.removeItem(draftKey.value)
    router.push('/article/' + route.params.id)
  } catch (e) {
    alert('保存失败')
  }
  saving.value = false
}

const createCategory = async () => {
  if (!newCategory.value.name || !newCategory.value.slug) {
    alert('请填写完整信息')
    return
  }
  try {
    const res = await apiCreateCategory(newCategory.value)
    categories.value.push(res.data)
    form.value.categoryId = res.data.id
    showNewCategory.value = false
    newCategory.value = { name: '', slug: '' }
  } catch (e) {
    alert('创建失败，分类可能已存在')
  }
}

const autoSave = () => {
  if (form.value.title || form.value.content) {
    localStorage.setItem(draftKey.value, JSON.stringify(form.value))
    lastSaved.value = new Date().toLocaleTimeString('zh-CN')
  }
}

const restoreDraft = () => {
  const draft = localStorage.getItem(draftKey.value)
  if (draft && confirm('确定要恢复草稿吗？当前内容将被覆盖。')) {
    Object.assign(form.value, JSON.parse(draft))
  }
}

const onContentInput = () => {
  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
  }
  autoSaveTimer.value = setTimeout(autoSave, 30000)
}

watch(() => [form.value.title, form.value.content], () => {
  onContentInput()
}, { deep: true })

onMounted(() => {
  fetchCategories()
  fetchArticle()
  
  const existingDraft = localStorage.getItem(draftKey.value)
  if (existingDraft && !isEdit.value && confirm('发现未保存的草稿，要恢复吗？')) {
    Object.assign(form.value, JSON.parse(existingDraft))
  }
})

onUnmounted(() => {
  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
  }
})
</script>

<style scoped>
.article-edit {
  max-width: 1000px;
  margin: 0 auto;
}

.form-group {
  margin-bottom: 15px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr 100px;
  gap: 15px;
  align-items: end;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

.form-group input, .form-group select {
  width: 100%;
}

.small-btn {
  margin-top: 5px;
  padding: 6px 12px;
  font-size: 12px;
}

.editor-container {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  padding: 10px 15px;
  background: var(--code-bg);
  font-size: 14px;
}

.auto-save {
  color: var(--success-color);
  font-size: 12px;
}

.editor-wrapper {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 500px;
}

.editor, .preview {
  padding: 15px;
  min-height: 500px;
}

.editor {
  border: none;
  border-right: 1px solid var(--border-color);
  resize: none;
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.preview {
  overflow-y: auto;
  max-height: 600px;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.secondary {
  background: #ccc;
  color: #333;
}

.word-count {
  text-align: right;
  color: #888;
  font-size: 13px;
  margin-top: 10px;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 400px;
}

.modal-content h3 {
  margin-bottom: 15px;
}

.modal-content input {
  width: 100%;
  margin-bottom: 10px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
  justify-content: flex-end;
}
</style>
