import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)
app.use(router)
app.mount('#app')

if (!localStorage.getItem('clientId')) {
  localStorage.setItem('clientId', 'user_' + Math.random().toString(36).substr(2, 9))
}
