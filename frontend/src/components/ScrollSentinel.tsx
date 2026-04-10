import { useEffect, useRef } from 'react'

interface Props {
  onVisible: () => void
}

export function ScrollSentinel({ onVisible }: Props) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          onVisible()
        }
      },
      { rootMargin: '200px' },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [onVisible])

  return <div ref={ref} style={{ height: 1 }} aria-hidden="true" />
}
