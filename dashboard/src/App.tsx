import { Route, Routes } from 'react-router-dom'

import { AppLayout } from 'layouts'
import { Dashboard } from 'pages'

function App() {
  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route index element={<Dashboard />} />
      </Route>
    </Routes>
  )
}

export default App
