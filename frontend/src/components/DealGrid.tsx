import { useInfiniteDeals } from '../hooks/useInfiniteDeals'
import { DealCard } from './DealCard'
import { ErrorBanner } from './ErrorBanner'
import { LoadingSkeleton } from './LoadingSkeleton'
import { ScrollSentinel } from './ScrollSentinel'

export function DealGrid() {
  const { deals, hasMore, loading, error, loadMore, retry } = useInfiniteDeals()

  return (
    <div>
      {deals.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
          {deals.map(deal => (
            <DealCard key={deal.id} deal={deal} />
          ))}
        </div>
      )}

      {loading && (
        <div className={deals.length > 0 ? 'mt-6' : ''}>
          <LoadingSkeleton />
        </div>
      )}

      {error && (
        <div className="mt-6">
          <ErrorBanner message={error} onRetry={retry} />
        </div>
      )}

      {!loading && !error && deals.length === 0 && (
        <div className="text-center py-24 text-gray-500">
          <p className="text-4xl mb-4">🔥</p>
          <p className="text-lg font-medium">No hot deals yet</p>
          <p className="text-sm mt-1">The scraper is warming up — check back in a minute!</p>
        </div>
      )}

      {!loading && !error && hasMore && (
        <ScrollSentinel onVisible={loadMore} />
      )}

      {!hasMore && deals.length > 0 && (
        <p className="text-center text-gray-400 text-sm mt-10 pb-6">
          You've seen all {deals.length} deals
        </p>
      )}
    </div>
  )
}
