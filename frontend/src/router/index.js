import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import DataSource from '../views/DataSource.vue'
import Task from '../views/Task.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/datasource',
    name: 'DataSource',
    component: DataSource
  },
  {
    path: '/task',
    name: 'Task',
    component: Task
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router