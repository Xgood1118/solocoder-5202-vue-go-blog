import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import ArticleDetail from '../views/ArticleDetail.vue'
import ArticleEdit from '../views/ArticleEdit.vue'
import Categories from '../views/Categories.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/article/:id', component: ArticleDetail },
  { path: '/edit/:id?', component: ArticleEdit },
  { path: '/categories', component: Categories }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
