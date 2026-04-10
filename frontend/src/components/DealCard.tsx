import { useState } from 'react'
import { recordClick } from '../api/dealsApi'
import type { Deal } from '../types/deal'
import { HotnessBadge } from './HotnessBadge'

interface Props {
  deal: Deal
}

const PLACEHOLDER = 'https://placehold.co/400x300/f3f4f6/9ca3af?text=No+Image'

export function DealCard({ deal }: Props) {
  const [imgSrc, setImgSrc] = useState(deal.image_url || PLACEHOLDER)
  const [clicking, setClicking] = useState(false)

  const handleGetDeal = async () => {
    if (clicking) return
    setClicking(true)
    try {
      const redirectURL = await recordClick(deal)
      window.open(redirectURL, '_blank', 'noopener,noreferrer')
    } finally {
      setClicking(false)
    }
  }

  const hasDiscount = deal.original_price && deal.original_price !== deal.price

  return (
    <article className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden flex flex-col hover:shadow-md transition-shadow duration-200">
      {/* Image */}
      <div className="relative bg-gray-50 flex items-center justify-center h-48 overflow-hidden">
        <img
          src={imgSrc}
          alt={deal.title}
          className="object-contain max-h-48 w-full p-2"
          loading="lazy"
          onError={() => setImgSrc(PLACEHOLDER)}
        />
        <div className="absolute top-2 right-2">
          <HotnessBadge temperature={deal.temperature} />
        </div>
      </div>

      {/* Content */}
      <div className="p-4 flex flex-col flex-1">
        <h2 className="text-gray-900 font-semibold text-sm leading-snug line-clamp-2 mb-2">
          {deal.title}
        </h2>

        {deal.description && (
          <p className="text-gray-500 text-xs leading-relaxed line-clamp-2 mb-3">
            {deal.description}
          </p>
        )}

        {/* Price row */}
        <div className="flex items-baseline gap-2 mt-auto mb-3">
          {deal.price && (
            <span className="text-lg font-bold text-gray-900">{deal.price}</span>
          )}
          {hasDiscount && (
            <span className="text-sm text-gray-400 line-through">{deal.original_price}</span>
          )}
        </div>

        {/* Clicks indicator */}
        {deal.click_count > 0 && (
          <p className="text-xs text-gray-400 mb-2">
            {deal.click_count} {deal.click_count === 1 ? 'person' : 'people'} clicked
          </p>
        )}

        {/* CTA */}
        <button
          onClick={handleGetDeal}
          disabled={clicking}
          className="w-full bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white font-semibold py-2.5 px-4 rounded-xl transition-colors duration-150 flex items-center justify-center gap-2"
        >
          {clicking ? (
            <>
              <span className="inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              Opening…
            </>
          ) : (
            'Get Deal on Amazon →'
          )}
        </button>
      </div>
    </article>
  )
}
