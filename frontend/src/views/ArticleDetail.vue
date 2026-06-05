<template>
  <div class="article-detail" v-if="article">
    <div class="main-column">
      <div class="card">
        <div class="article-header">
          <h1>{{ article.title }}</h1>
          <div class="article-meta">
            <span>{{ article.author }}</span>
            <span>{{ formatDate(article.createdAt) }}</span>
            <span>👁️ {{ article.viewCount }}</span>
            <span>❤️ {{ article.likeCount }}</span>
            <span>💬 {{ article.commentCount }}</span>
            <span>{{ article.wordCount }}字</span>
            <span>{{ article.readTime }}分钟阅读</span>
          </div>
          <div class="article-actions">
            <button @click="likeArticle" :disabled="liked">
              {{ liked ? '已点赞' : '点赞' }} ({{ article.likeCount }})
            </button>
            <button @click="exportPDF">导出PDF</button>
            <router-link :to="'/edit/' + article.id">
              <button>编辑</button>
            </router-link>
          </div>
          <div class="article-tags">
            <span v-for="tag in article.tags" :key="tag" class="tag">{{ tag }}</span>
          </div>
          <div class="pin-rule">
            📌 置顶规则：按置顶时间倒序排在最前，同置顶时间内的按发布时间倒序
          </div>
        </div>
        <hr>
        <div id="article-content" class="markdown-content" v-html="article.contentHTML"></div>
      </div>

      <div class="card" v-if="related.length > 0">
        <h3>相关推荐</h3>
        <div class="related-list">
          <router-link 
            v-for="item in related" 
            :key="item.id" 
            :to="'/article/' + item.id"
            class="related-item"
          >
            <span>{{ item.title }}</span>
            <span class="meta">{{ item.author }} · {{ item.viewCount }}阅</span>
          </router-link>
        </div>
      </div>

      <div class="card">
        <h3>评论 ({{ comments.length }})</h3>
        <div class="comment-form">
          <input 
            type="text" 
            v-model="newComment.author" 
            placeholder="你的名字"
            :disabled="replyingTo"
          >
          <textarea 
            v-model="newComment.content" 
            placeholder="写下你的评论..."
            rows="3"
          ></textarea>
          <div v-if="replyingTo" class="replying-to">
            回复 @{{ replyingTo.author }} 
            <button class="cancel-btn" @click="cancelReply">取消</button>
          </div>
          <button @click="submitComment">发表评论</button>
        </div>

        <div v-for="comment in topLevelComments" :key="comment.id" class="comment">
          <div class="comment-header">
            <strong>{{ comment.author }}</strong>
            <span>{{ formatDate(comment.createdAt) }}</span>
          </div>
          <div class="comment-content">{{ comment.content }}</div>
          <div class="comment-actions">
            <button class="action-btn" @click="likeComment(comment)">
              👍 {{ comment.likeCount }}
            </button>
            <button class="action-btn" @click="startReply(comment)">回复</button>
            <button class="action-btn report" @click="reportComment(comment)">举报</button>
          </div>
          <div v-if="getReplies(comment.id).length > 0" class="replies">
            <div v-for="reply in getReplies(comment.id)" :key="reply.id" class="comment reply">
              <div class="comment-header">
                <strong>{{ reply.author }}</strong>
                <span>回复</span>
                <span>{{ formatDate(reply.createdAt) }}</span>
              </div>
              <div class="comment-content">{{ reply.content }}</div>
              <div class="comment-actions">
                <button class="action-btn" @click="likeComment(reply)">
                  👍 {{ reply.likeCount }}
                </button>
                <button class="action-btn report" @click="reportComment(reply)">举报</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <aside class="sidebar">
      <div class="card toc-card" v-if="toc.length > 0">
        <h3>目录</h3>
        <div class="toc-list">
          <div 
            v-for="item in toc" 
            :key="item.id"
            :class="['toc-item', 'level-' + item.level, { active: activeToc === item.id }]"
            @click="scrollTo(item.id)"
          >{{ item.text }}</div>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'
import { 
  getArticle, likeArticle as apiLikeArticle, 
  createComment, likeComment as apiLikeComment, 
  reportComment as apiReportComment 
} from '../api'

const route = useRoute()
const article = ref(null)
const comments = ref([])
const related = ref([])
const toc = ref([])
const activeToc = ref('')
const liked = ref(false)
const newComment = ref({ author: '', content: '' })
const replyingTo = ref(null)

const fetchArticle = async () => {
  const res = await getArticle(route.params.id)
  article.value = res.data.article
  comments.value = res.data.comments
  related.value = res.data.related
  await nextTick()
  generateTOC()
}

const generateTOC = () => {
  const content = document.getElementById('article-content')
  if (!content) return
  
  const headings = content.querySelectorAll('h1, h2, h3')
  toc.value = []
  headings.forEach((h, i) => {
    const id = 'heading-' + i
    h.id = id
    toc.value.push({
      id,
      level: parseInt(h.tagName[1]),
      text: h.textContent
    })
  })
}

const scrollTo = (id) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}

const handleScroll = () => {
  const headings = document.querySelectorAll('#article-content h1, #article-content h2, #article-content h3')
  for (let i = headings.length - 1; i >= 0; i--) {
    const rect = headings[i].getBoundingClientRect()
    if (rect.top <= 100) {
      activeToc.value = headings[i].id
      break
    }
  }
}

const likeArticle = async () => {
  try {
    await apiLikeArticle(article.value.id)
    liked.value = true
    article.value.likeCount++
  } catch (e) {
    alert('你已经点过赞了')
  }
}

const topLevelComments = () => comments.value.filter(c => !c.parentId)
const getReplies = (parentId) => comments.value.filter(c => c.parentId === parentId)

const startReply = (comment) => {
  replyingTo.value = comment
  newComment.value.author = ''
}

const cancelReply = () => {
  replyingTo.value = null
}

const submitComment = async () => {
  if (!newComment.value.content.trim()) return
  
  await createComment(article.value.id, {
    author: newComment.value.author || '匿名',
    content: newComment.value.content,
    parentId: replyingTo.value?.id
  })
  
  newComment.value = { author: '', content: '' }
  replyingTo.value = null
  fetchArticle()
}

const likeComment = async (comment) => {
  try {
    await apiLikeComment(comment.id)
    comment.likeCount++
  } catch (e) {
    alert('你已经点过赞了')
  }
}

const reportComment = async (comment) => {
  const reason = prompt('请输入举报原因：')
  if (reason) {
    await apiReportComment(comment.id, reason)
    alert('举报已提交')
  }
}

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

const exportPDF = async () => {
  const element = document.getElementById('article-content')
  const canvas = await html2canvas(element, {
    scale: 2,
    useCORS: true,
    backgroundColor: '#ffffff'
  })
  
  const imgData = canvas.toDataURL('image/png')
  const pdf = new jsPDF('p', 'mm', 'a4')
  const imgWidth = 210
  const pageHeight = 297
  const imgHeight = (canvas.height * imgWidth) / canvas.width
  let heightLeft = imgHeight
  let position = 0

  pdf.addImage(imgData, 'PNG', 0, position, imgWidth, imgHeight)
  heightLeft -= pageHeight

  while (heightLeft > 0) {
    position = heightLeft - imgHeight
    pdf.addPage()
    pdf.addImage(imgData, 'PNG', 0, position, imgWidth, imgHeight)
    heightLeft -= pageHeight
  }

  pdf.save(`${article.value.title}.pdf`)
}

onMounted(() => {
  fetchArticle()
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

watch(() => route.params.id, () => {
  fetchArticle()
})
</script>

<style scoped>
.article-detail {
  display: grid;
  grid-template-columns: 1fr 250px;
  gap: 20px;
}

.article-header h1 {
  margin: 0 0 15px;
  font-size: 28px;
}

.article-meta {
  display: flex;
  gap: 15px;
  font-size: 14px;
  color: #888;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.article-actions {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.article-tags {
  margin-bottom: 15px;
}

.pin-rule {
  background: var(--code-bg);
  padding: 10px 15px;
  border-radius: 4px;
  font-size: 13px;
  color: #888;
  margin-bottom: 15px;
}

.comment-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}

.replying-to {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--code-bg);
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 13px;
}

.cancel-btn {
  background: none;
  color: var(--danger-color);
  padding: 0;
}

.comment {
  padding: 15px 0;
  border-bottom: 1px solid var(--border-color);
}

.comment:last-child {
  border-bottom: none;
}

.comment-header {
  display: flex;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 13px;
  color: #888;
}

.comment-header strong {
  color: var(--text-color);
}

.comment-content {
  margin-bottom: 10px;
}

.comment-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  background: none;
  color: var(--text-color);
  padding: 4px 8px;
  font-size: 12px;
}

.action-btn.report {
  color: var(--danger-color);
}

.replies {
  margin-left: 30px;
  margin-top: 10px;
  border-left: 2px solid var(--border-color);
  padding-left: 15px;
}

.related-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.related-item {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  background: var(--code-bg);
  border-radius: 4px;
  text-decoration: none;
  color: var(--text-color);
}

.related-item:hover {
  background: var(--border-color);
}

.related-item .meta {
  font-size: 12px;
  color: #888;
}

.toc-card {
  position: sticky;
  top: 80px;
}

.toc-list {
  max-height: 500px;
  overflow-y: auto;
}

.toc-item {
  padding: 6px 0;
  cursor: pointer;
  font-size: 14px;
  color: #888;
}

.toc-item:hover {
  color: var(--primary-color);
}

.toc-item.level-2 {
  padding-left: 15px;
}

.toc-item.level-3 {
  padding-left: 30px;
}

.toc-item.active {
  color: var(--primary-color);
  font-weight: bold;
}
</style>
