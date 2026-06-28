import { useEffect, useState } from 'react'
import { listPhotos, thumbURL } from '../api'
import type { Attachment } from '../types'

export default function CoverThumb({ owner, id, emoji }: { owner: string; id: number; emoji: string }) {
  const [cover, setCover] = useState<Attachment | null>(null)

  useEffect(() => {
    let active = true
    listPhotos(owner, id)
      .then((p) => {
        if (!active) return
        setCover(p.find((x) => x.cover) ?? p[0] ?? null)
      })
      .catch(() => {})
    return () => { active = false }
  }, [owner, id])

  if (!cover) {
    return (
      <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-lg bg-dracula-current text-2xl">
        {emoji}
      </div>
    )
  }
  return <img src={thumbURL(cover.id, cover.updatedAt)} alt="" className="h-14 w-14 shrink-0 rounded-lg object-cover" />
}
