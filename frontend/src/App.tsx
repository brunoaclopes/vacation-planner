import { Routes, Route } from 'react-router-dom'
import { Box } from '@mui/material'
import Layout from './components/Layout'
import CalendarPage from './pages/CalendarPage'
import SettingsPage from './pages/SettingsPage'
import { CalendarProvider } from './context/CalendarContext'
import { I18nProvider } from './i18n'

function App() {
  return (
    <I18nProvider>
      <CalendarProvider>
        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
          <Layout>
            <Routes>
              <Route path="/" element={<CalendarPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </Layout>
        </Box>
      </CalendarProvider>
    </I18nProvider>
  )
}

export default App
