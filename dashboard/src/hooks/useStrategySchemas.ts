import { useCallback } from 'react'

import { getStrategySchemas } from 'api'
import { useAsync } from 'hooks'

export default () => useAsync(useCallback(getStrategySchemas, []))
