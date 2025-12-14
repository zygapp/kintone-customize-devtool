import './style.css'

kintone.events.on('app.record.index.show', (event) => {
  const el = kintone.app.getHeaderSpaceElement()
  if (el && !el.querySelector('#kcdev-root')) {
    const root = document.createElement('div')
    root.id = 'kcdev-root'
    root.className = 'kcdev-app'
    root.innerHTML = `
      <h1>kintone Customize</h1>
      <button id="kcdev-counter">count is 0</button>
    `
    el.appendChild(root)

    let count = 0
    const button = root.querySelector('#kcdev-counter')
    button.addEventListener('click', () => {
      count++
      button.textContent = `count is ${count}`
    })
  }
  return event
})
