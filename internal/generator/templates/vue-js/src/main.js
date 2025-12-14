import { createApp } from 'vue'
import App from './App.vue'
import './style.css'

kintone.events.on('app.record.index.show', (event) => {
  const el = kintone.app.getHeaderSpaceElement()
  if (el && !el.querySelector('#kcdev-root')) {
    const root = document.createElement('div')
    root.id = 'kcdev-root'
    el.appendChild(root)
    createApp(App).mount(root)
  }
  return event
})
