import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@fontsource/dm-serif-text'
import '@fontsource/dm-serif-text/400-italic.css'
import '@fontsource-variable/plus-jakarta-sans'
import '@fontsource-variable/plus-jakarta-sans/wght-italic.css'
import './styles/main.css'
import App from './App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
