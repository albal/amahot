import { DealGrid } from './components/DealGrid'

export default function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-2xl" aria-hidden="true">🔥</span>
            <div>
              <h1 className="text-xl font-bold text-gray-900 leading-tight">Hot Amazon Deals</h1>
              <p className="text-xs text-gray-500 leading-none">Sorted by community hotness</p>
            </div>
          </div>
          <span className="hidden sm:inline-flex items-center gap-1 text-xs text-gray-400 bg-gray-100 px-3 py-1.5 rounded-full">
            ♨️ Over 100° only
          </span>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <DealGrid />
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 mt-8 py-6 text-center text-xs text-gray-400">
        <p>
          Deals sourced from{' '}
          <a
            href="https://www.hotukdeals.com"
            target="_blank"
            rel="noopener noreferrer"
            className="underline hover:text-gray-600"
          >
            hotukdeals.com
          </a>
          . As an Amazon Associate we earn from qualifying purchases.
        </p>
      </footer>
    </div>
  )
}
