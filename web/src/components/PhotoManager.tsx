import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { deletePhoto, listPhotos, photoURL, rotatePhoto, setCover, thumbURL, uploadPhoto } from '../api'
import { prepareImageUpload } from '../images'
import type { Attachment } from '../types'

export default function PhotoManager({ owner, id }: { owner: string; id: number }) {
  const [photos, setPhotos] = useState<Attachment[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [viewer, setViewer] = useState<number | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  const refresh = () => listPhotos(owner, id).then(setPhotos).catch(() => {})
  useEffect(() => { refresh() }, [owner, id])

  const onFile = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setBusy(true)
    setErr(null)
    try {
      const { blob, name } = await prepareImageUpload(file)
      await uploadPhoto(owner, id, blob, name)
      await refresh()
    } catch (e2) {
      setErr(e2 instanceof Error ? e2.message : 'Upload failed')
    } finally {
      setBusy(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  const remove = async (pid: number) => {
    await deletePhoto(pid)
    if (viewer === pid) setViewer(null)
    await refresh()
  }

  const makeCover = async (pid: number) => {
    await setCover(pid)
    await refresh()
  }

  // Rotate rewrites the file server-side and bumps updatedAt; refreshing pulls the
  // new stamp, which changes the image URL so the browser fetches the rotated bytes.
  const rotate = async (pid: number) => {
    await rotatePhoto(pid)
    await refresh()
  }

  const viewed = viewer === null ? null : photos.find((p) => p.id === viewer)

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-sm font-medium text-dracula-comment">Photos</span>
        <label className="cursor-pointer rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-fg">
          {busy ? 'Uploading…' : '+ Photo'}
          <input ref={fileRef} type="file" accept="image/*,.heic,.heif" className="hidden" onChange={onFile} disabled={busy} />
        </label>
      </div>
      {err && <p className="mb-2 text-sm text-dracula-red">{err}</p>}
      {photos.length === 0 ? (
        <p className="text-sm text-dracula-comment">No photos yet.</p>
      ) : (
        <div className="grid grid-cols-3 gap-2">
          {photos.map((p) => (
            <div key={p.id} className="relative">
              <button type="button" onClick={() => setViewer(p.id)} className="block w-full" title="Tap to enlarge">
                <img
                  src={thumbURL(p.id, p.updatedAt)}
                  alt={p.filename}
                  className={`aspect-square w-full rounded-lg object-cover ${p.cover ? 'ring-2 ring-dracula-purple' : ''}`}
                />
              </button>
              <button
                onClick={() => makeCover(p.id)}
                title={p.cover ? 'Cover photo' : 'Set as cover'}
                className={`absolute left-1 top-1 rounded-md bg-dracula-bg/80 px-1.5 py-0.5 text-xs ${p.cover ? 'text-dracula-yellow' : 'text-dracula-fg'}`}
              >
                {p.cover ? '★' : '☆'}
              </button>
              <div className="absolute right-1 top-1 flex gap-1">
                <button
                  onClick={() => rotate(p.id)}
                  title="Rotate 90°"
                  className="rounded-md bg-dracula-bg/80 px-1.5 py-0.5 text-xs text-dracula-fg"
                >
                  ⟳
                </button>
                <button
                  onClick={() => remove(p.id)}
                  title="Delete"
                  className="rounded-md bg-dracula-bg/80 px-1.5 py-0.5 text-xs text-dracula-red"
                >
                  ✕
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
      {viewed && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-4"
          onClick={() => setViewer(null)}
        >
          <img src={photoURL(viewed.id, viewed.updatedAt)} alt="" className="max-h-full max-w-full rounded-lg object-contain" />
        </div>
      )}
    </div>
  )
}
