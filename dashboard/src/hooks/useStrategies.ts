import { getStrategies } from 'api'
import { useAsync } from 'hooks'

export default () => useAsync(getStrategies)
