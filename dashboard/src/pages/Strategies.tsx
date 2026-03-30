import { useNavigate } from 'react-router-dom'
// import { Plus } from 'lucide-react'

import { BarLoader, Heading } from 'components'
import { useStrategies } from 'hooks'

const Strategies = () => {
  const navigate = useNavigate()
  const { data: strategies, loading: loadingStrategies } = useStrategies()

  if (loadingStrategies) return <BarLoader fullscreen />

  return (
    <div className="space-y-8">
      <Heading title="Strategies" />
      {/* <button
        onClick={() => navigate('/strategies/create')}
        className="flex flex-row flex-nowrap gap-2 px-2 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all"
      >
        New <Plus size={18} />
      </button> */}
      <div className="space-y-4">
        {!strategies ||
        strategies.length === 0 ? //     onClick={() => navigate('/strategies/create')} //   <button //   No strategies yet.{' '} // <p className="text-content-secondary text-sm">
        //     className="text-accent hover:opacity-80 transition-opacity cursor-pointer"
        //   >
        //     Create one →
        //   </button>
        // </p>
        null : (
          <ul>
            {strategies.map((strategy) => (
              <li
                className="text-white cursor-pointer hover:text-gray-500"
                key={strategy.slug}
                onClick={() => navigate(`/strategies/${strategy.slug}`)}
              >
                {strategy.display_name}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

export default Strategies
