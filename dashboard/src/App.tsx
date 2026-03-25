import { Route, Routes } from 'react-router-dom'

import { AppLayout } from 'layouts'
import {
  Accounts,
  Bots,
  BotDetails,
  Dashboard,
  Diagnostics,
  Trades,
  Strategies,
  StrategyDetails,
  StrategyCreator,
} from 'pages'
import { AlertBanner, AlertProvider } from 'components'
import { BreadcrumbProvider } from 'context'

function App() {
  return (
    <AlertProvider>
      <AlertBanner />
      <BreadcrumbProvider>
        <Routes>
          <Route path="/" element={<AppLayout />}>
            <Route index element={<Dashboard />} />
            <Route path="accounts" element={<Accounts />} />
            <Route path="bots" element={<Bots />} />
            <Route path="bots/:id" element={<BotDetails />} />
            <Route path="diagnostics" element={<Diagnostics />} />

            <Route path="strategies" element={<Strategies />} />
            <Route path="strategies/:id" element={<StrategyDetails />} />
            <Route path="strategies/create" element={<StrategyCreator />} />

            <Route path="trades" element={<Trades />} />
          </Route>
        </Routes>
      </BreadcrumbProvider>
    </AlertProvider>
  )
}

export default App
