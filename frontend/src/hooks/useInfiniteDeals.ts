import { useCallback, useEffect, useRef, useState } from 'react'
import { fetchDeals } from '../api/dealsApi'
import type { Deal } from '../types/deal'

const PAGE_SIZE = 20

export function useInfiniteDeals() {
  const [deals, setDeals] = useState<Deal[]>([])
  const [offset, setOffset] = useState(0)
  const [hasMore, setHasMore] = useState(true)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const loadingRef = useRef(false)

  const loadMore = useCallback(async () => {
    // Use ref to prevent double-loads from StrictMode double-effect
    if (loadingRef.current || !hasMore) return
    loadingRef.current = true
    setLoading(true)
    setError(null)

    try {
      const result = await fetchDeals({ limit: PAGE_SIZE, offset })
      setDeals(prev => {
        // Deduplicate by id in case of overlap
        const existing = new Set(prev.map(d => d.id))
        const newDeals = result.data.filter(d => !existing.has(d.id))
        return [...prev, ...newDeals]
      })
      setOffset(prev => prev + result.data.length)
      setHasMore(result.has_more)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load deals')
      setHasMore(false)
    } finally {
      setLoading(false)
      loadingRef.current = false
    }
  }, [hasMore, offset])

  // Load first page on mount
  useEffect(() => {
    loadMore()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const retry = useCallback(() => {
    setError(null)
    setHasMore(true)
    loadMore()
  }, [loadMore])

  return { deals, hasMore, loading, error, loadMore, retry }
}
