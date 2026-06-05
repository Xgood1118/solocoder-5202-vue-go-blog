<template>
  <div class="categories">
    <div class="card">
      <h2>分类管理</h2>
      
      <div class="create-form">
        <input type="text" v-model="newCategory.name" placeholder="分类名称">
        <input type="text" v-model="newCategory.slug" placeholder="分类标识 (slug)">
        <button @click="handleCreateCategory">新建分类</button>
      </div>

      <table class="category-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>标识</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cat in categories" :key="cat.id">
            <td>{{ cat.id.slice(0, 8) }}</td>
            <td>
              <input 
                v-if="editingId === cat.id" 
                v-model="editForm.name"
              >
              <span v-else>{{ cat.name }}</span>
            </td>
            <td>
              <input 
                v-if="editingId === cat.id" 
                v-model="editForm.slug"
              >
              <span v-else>{{ cat.slug }}</span>
            </td>
            <td class="actions">
              <template v-if="editingId === cat.id">
                <button class="small-btn" @click="saveEdit(cat.id)">保存</button>
                <button class="small-btn secondary" @click="cancelEdit">取消</button>
              </template>
              <template v-else>
                <button class="small-btn" @click="startEdit(cat)">编辑</button>
                <button class="small-btn danger" @click="handleDeleteCategory(cat)">删除</button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { 
  getCategories, 
  createCategory as apiCreateCategory, 
  updateCategory as apiUpdateCategory, 
  deleteCategory as apiDeleteCategory 
} from '../api'

const categories = ref([])
const newCategory = ref({ name: '', slug: '' })
const editingId = ref(null)
const editForm = ref({ name: '', slug: '' })

const fetchCategories = async () => {
  const res = await getCategories()
  categories.value = res.data
}

const handleCreateCategory = async () => {
  if (!newCategory.value.name || !newCategory.value.slug) {
    alert('请填写完整信息')
    return
  }
  try {
    await apiCreateCategory(newCategory.value)
    newCategory.value = { name: '', slug: '' }
    fetchCategories()
  } catch (e) {
    alert('创建失败，分类可能已存在')
  }
}

const startEdit = (cat) => {
  editingId.value = cat.id
  editForm.value = { name: cat.name, slug: cat.slug }
}

const cancelEdit = () => {
  editingId.value = null
}

const saveEdit = async (id) => {
  try {
    await apiUpdateCategory(id, editForm.value)
    editingId.value = null
    fetchCategories()
  } catch (e) {
    alert('更新失败')
  }
}

const handleDeleteCategory = async (cat) => {
  if (!confirm(`确定要删除分类 "${cat.name}" 吗？`)) return
  try {
    await apiDeleteCategory(cat.id)
    fetchCategories()
  } catch (e) {
    alert('删除失败，该分类下可能还有文章')
  }
}

onMounted(() => {
  fetchCategories()
})
</script>

<style scoped>
.categories {
  max-width: 800px;
  margin: 0 auto;
}

.create-form {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.create-form input {
  flex: 1;
}

.category-table {
  width: 100%;
  border-collapse: collapse;
}

.category-table th, .category-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

.category-table th {
  background: var(--code-bg);
  font-weight: 600;
}

.category-table input {
  width: 100%;
}

.actions {
  display: flex;
  gap: 8px;
}

.small-btn {
  padding: 4px 10px;
  font-size: 12px;
}

.small-btn.secondary {
  background: #ccc;
  color: #333;
}

.small-btn.danger {
  background: var(--danger-color);
}
</style>
