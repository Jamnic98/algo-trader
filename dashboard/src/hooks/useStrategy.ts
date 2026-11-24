import { useCallback } from 'react'

import { getStrategy } from 'api'
import { useAsync } from 'hooks'

export default (id: string) => useAsync(useCallback(() => getStrategy(id), [id]))
