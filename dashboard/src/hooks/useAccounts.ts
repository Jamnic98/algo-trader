import { useCallback } from 'react'

import { listAccounts } from 'api'
import useAsync from './useAsync'

export default () => useAsync(useCallback(listAccounts, []))
