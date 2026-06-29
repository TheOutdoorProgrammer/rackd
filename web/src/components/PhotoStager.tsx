import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { prepareImageUpload } from '../images'

// A photo picked on the New form, held in the browser until the item exists and
// the bytes can be uploaded against its id. `url` is an object URL for the
// preview thumbnail and must be revoked when the photo is dropped.
export interface StagedPhoto {
  key: string
  blob: Blob
  name: string
  url: string
}

// PhotoStager is the create-time counterpart to PhotoManager: it collects photos
// before a record has an id. It's a controlled component — the parent form owns
// the staged list and uploads it after createItem returns.
export default function PhotoStager({
  photos,
  onChange,
}: {
  photos: StagedPhoto[]
  onChange: (photos: StagedPhoto[]) => void
}) {
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  // Revoke any outstanding preview URLs when the stager unmounts (e.g. the form
  // navigates away after save) so the blobs aren't leaked.
  const photosRef = useRef(photos)
  photosRef.current = photos
  useEffect(
    () => () => {
      photosRef.current.forEach((p) => URL.revokeObjectURL(p.url))
    },
    [],
  )

  const add = async (e: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? [])
    if (files.length === 0) return
    setBusy(true)
    setErr(null)
    const added: StagedPhoto[] = []
    try {
      for (const file of files) {
        const { blob, name } = await prepareImageUpload(file)
        added.push({ key: crypto.randomUUID(), blob, name, url: URL.createObjectURL(blob) })
      }
      onChange([...photos, ...added])
    } catch (e2) {
      added.forEach((p) => URL.revokeObjectURL(p.url))
      setErr(e2 instanceof Error ? e2.message : 'Could not add photo')
    } finally {
      setBusy(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  const remove = (key: string) => {
    const p = photos.find((x) => x.key === key)
    if (p) URL.revokeObjectURL(p.url)
    onChange(photos.filter((x) => x.key !== key))
  }

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-sm font-medium text-dracula-comment">Photos</span>
        <label className="cursor-pointer rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-fg">
          {busy ? 'Adding…' : '+ Photo'}
          <input
            ref={fileRef}
            type="file"
            accept="image/*,.heic,.heif"
            multiple
            className="hidden"
            onChange={add}
            disabled={busy}
          />
        </label>
      </div>
      {err && <p className="mb-2 text-sm text-dracula-red">{err}</p>}
      {photos.length === 0 ? (
        <p className="text-sm text-dracula-comment">Add photos now, or later from the item's page.</p>
      ) : (
        <div className="grid grid-cols-3 gap-2">
          {photos.map((p) => (
            <div key={p.key} className="relative">
              <img src={p.url} alt={p.name} className="aspect-square w-full rounded-lg object-cover" />
              <button
                type="button"
                onClick={() => remove(p.key)}
                title="Remove"
                className="absolute right-1 top-1 rounded-md bg-dracula-bg/80 px-1.5 py-0.5 text-xs text-dracula-red"
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
