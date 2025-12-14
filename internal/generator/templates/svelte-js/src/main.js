import App from './App.svelte'
import './style.css'

kintone.events.on('app.record.index.show', (event) => {
  const el = kintone.app.getHeaderSpaceElement()
  if (el && !el.querySelector('#kcdev-root')) {
    const root = document.createElement('div')
    root.id = 'kcdev-root'
    el.appendChild(root)
    new App({ target: root })
  }
  return event
})
