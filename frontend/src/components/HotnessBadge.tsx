interface Props {
  temperature: number
}

function badgeClass(temp: number): string {
  if (temp >= 500) return 'bg-red-600 text-white'
  if (temp >= 200) return 'bg-orange-500 text-white'
  return 'bg-green-600 text-white'
}

function flameIcon(temp: number): string {
  if (temp >= 500) return '🔥'
  if (temp >= 200) return '♨️'
  return '✅'
}

export function HotnessBadge({ temperature }: Props) {
  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-sm font-bold ${badgeClass(temperature)}`}
      title={`${temperature}° hotness`}
    >
      {flameIcon(temperature)} {temperature}°
    </span>
  )
}
