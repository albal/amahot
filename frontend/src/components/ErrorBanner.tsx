interface Props {
  message: string
  onRetry: () => void
}

export function ErrorBanner({ message, onRetry }: Props) {
  return (
    <div className="rounded-xl bg-red-50 border border-red-200 p-6 text-center">
      <p className="text-red-700 font-medium mb-3">{message}</p>
      <button
        onClick={onRetry}
        className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium"
      >
        Try again
      </button>
    </div>
  )
}
