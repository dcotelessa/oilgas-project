import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'

// Import views
import Dashboard from './views/Dashboard.vue'
import Customers from './views/Customers.vue'
import Inventory from './views/Inventory.vue'
import Grades from './views/Grades.vue'

// Router setup
const routes = [
  { path: '/', component: Dashboard },
  { path: '/customers', component: Customers },
  { path: '/inventory', component: Inventory },
  { path: '/grades', component: Grades },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// App setup
const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
