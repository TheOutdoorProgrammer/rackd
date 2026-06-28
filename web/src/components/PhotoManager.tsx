import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { deletePhoto, listPhotos, thumbURL, uploadPhoto } from '../api'
import type { Attachment } from '../types'

export default function PhotoManager({ owner, id }: { owner: string; id: number }) {
  const [photos, setPhotos] = useState<Attachment[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  const refresh = () => listPhotos(owner, id).then(setPhotos).catch(() => {})
  useEffect(() => { refresh() }, [owner, id])

  const onFile = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setBusy(true)
    setErr(null)
    try {
      let blob: Blob = file
      let name = file.name
      // iPhone HEIC/HEIF → convert to JPEG client-side (server stays CGO-free).
      if (/heic|heif/i.test(file.type) || /\.hei[cf]$/i.test(file.name)) {
        const heic2any = (await import('heic2any')).default
        blob = (await heic2any({ blob: file, toType: 'image/jpeg', quality: 0.9 })) as Blob
        name = name.replace(/\.\w+$/, '.jpg')
      }
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
    await refresh()
  }

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
            <div key={p.id} className="group relative">
              <img src={thumbURL(p.id)} alt={p.filename} className="aspect-square w-full rounded-lg object-cover" />
              <button
                onClick={() => remove(p.id)}
                className="absolute right-1 top-1 rounded-md bg-dracula-bg/80 px-1.5 text-xs text-dracula-red opacity-0 transition group-hover:opacity-100"
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
