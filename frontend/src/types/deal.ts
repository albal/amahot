export interface Deal {
  id: number
  external_id: string
  title: string
  description: string
  price: string
  original_price: string
  image_url: string
  deal_url: string
  merchant: string
  temperature: number
  category: string
  scraped_at: string
  updated_at: string
  is_expired: boolean
  click_count: number
}

export interface DealsResponse {
  data: Deal[]
  total: number
  limit: number
  offset: number
  has_more: boolean
}

export interface ClickResponse {
  redirect_url: string
}
