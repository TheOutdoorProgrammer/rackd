import { ammoseekURL } from '../ammoseek'

export default function AmmoSeekButton({ caliber, label }: { caliber: string; label: string }) {
  const url = ammoseekURL(caliber)
  if (!url) return null
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-1 rounded-lg bg-dracula-orange/90 px-3 py-2 text-sm font-medium text-dracula-bg transition hover:bg-dracula-orange"
    >
      {label} <span aria-hidden>↗</span>
    </a>
  )
}
