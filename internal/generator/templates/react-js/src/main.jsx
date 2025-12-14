import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './style.css'

kintone.events.on('app.record.index.show', (event) => {
  const el = kintone.app.getHeaderSpaceElement()
  if (el && !el.querySelector('#kcdev-root')) {
    const root = document.createElement('div')
    root.id = 'kcdev-root'
    el.appendChild(root)
    ReactDOM.createRoot(root).render(
      <React.StrictMode>
        <App />
      </React.StrictMode>
    )
  }
  return event
})
