import { useEffect, useState } from 'react'
import { listPhotos, thumbURL } from '../api'

export default function CoverThumb({ owner, id, emoji }: { owner: string; id: number; emoji: string }) {
  const [photoId, setPhotoId] = useState<number | null>(null)

  useEffect(() => {
    let active = true
    listPhotos(owner, id)
      .then((p) => { if (active && p.length) setPhotoId(p[0].id) })
      .catch(() => {})
    return () => { active = false }
  }, [owner, id])

  if (photoId === null) {
    return (
      <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-lg bg-dracula-current text-2xl">
        {emoji}
      </div>
    )
  }
  return <img src={thumbURL(photoId)} alt="" className="h-14 w-14 shrink-0 rounded-lg object-cover" />
}
