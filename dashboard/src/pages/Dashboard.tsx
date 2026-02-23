import { useNavigate } from 'react-router-dom'

const Dashboard = () => {
  const navigate = useNavigate()

  return (
    <div>
      <div>Dashboard Page</div>
      <div className="flex flex-col">
        <button onClick={() => navigate('bots')} className="cursor-pointer">
          bots page
        </button>
        <button onClick={() => navigate('trades')} className="cursor-pointer">
          trades page
        </button>
      </div>
    </div>
  )
}

export default Dashboard
