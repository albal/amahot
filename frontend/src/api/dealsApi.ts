import type { ClickResponse, Deal, DealsResponse } from '../types/deal'

export async function fetchDeals(params: {
  limit: number
  offset: number
}): Promise<DealsResponse> {
  const url = `/api/deals?limit=${params.limit}&offset=${params.offset}`
  const res = await fetch(url)
  if (!res.ok) {
    throw new Error(`Failed to fetch deals: ${res.status}`)
  }
  const data: DealsResponse = await res.json()
  // Normalise null data array to empty array
  if (!data.data) {
    data.data = []
  }
  return data
}

export async function recordClick(deal: Deal): Promise<string> {
  const res = await fetch(`/api/clicks/${deal.id}`, { method: 'POST' })
  if (!res.ok) {
    // Fall back to the stored deal_url directly
    return deal.deal_url
  }
  const data: ClickResponse = await res.json()
  return data.redirect_url || deal.deal_url
}
