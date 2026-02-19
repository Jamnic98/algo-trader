import { Outlet } from 'react-router-dom'

import { Footer /* Header */ } from 'components'

const AppLayout = () => {
  return (
    <div className="min-h-screen flex flex-col">
      {/* <Header /> */}
      <main className="grow mx-auto w-full">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}

export default AppLayout
