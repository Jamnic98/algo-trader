import { Route, Routes } from 'react-router-dom'

import { AppLayout } from 'layouts'
import { Bots, BotDetail, Dashboard, Trades } from 'pages'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="bots" element={<Bots />} />
        <Route path="bots/:id" element={<BotDetail />} />
        <Route path="trades" element={<Trades />} />
      </Route>
    </Routes>
  )
}

export default App
