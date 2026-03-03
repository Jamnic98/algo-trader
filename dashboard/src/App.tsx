import { Route, Routes } from 'react-router-dom'

import { AppLayout } from 'layouts'
import { Accounts, Bots, BotInfo, Dashboard, Diagnostics, Trades } from 'pages'
import { AlertBanner, AlertProvider } from 'components'

function App() {
  return (
    <AlertProvider>
      <AlertBanner />
      <Routes>
        <Route path="/" element={<AppLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="accounts" element={<Accounts />} />
          <Route path="bots" element={<Bots />} />
          <Route path="bots/:id" element={<BotInfo />} />
          <Route path="diagnostics" element={<Diagnostics />} />
          <Route path="trades" element={<Trades />} />
        </Route>
      </Routes>
    </AlertProvider>
  )
}

export default App
