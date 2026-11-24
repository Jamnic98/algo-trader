// hooks/useIsMobile.ts
import { useEffect, useState } from 'react'

const MOBILE_BREAKPOINT = 768

const useIsMobile = () => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < MOBILE_BREAKPOINT)

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    })

    observer.observe(document.body)
    return () => observer.disconnect()
  }, [])

  return isMobile
}

export default useIsMobile
