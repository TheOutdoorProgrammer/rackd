import { useEffect, useState, type FormEvent } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'
import { createItem, getItem, listItems, updateItem, uploadPhoto } from '../api'
import { RESOURCES, type Field, type ResourceConfig } from '../resources'
import { centsToDollars, dollarsToCents } from '../format'
import SpecLookup from '../components/SpecLookup'
import PhotoStager, { type StagedPhoto } from '../components/PhotoStager'
import type { Firearm, Item } from '../types'

// A <select> renders its first <option> but never fires onChange until tapped,
// so an untouched field would save empty. Seed the first option's value so what's
// shown is what's saved (nfaType's first option is "", so it correctly stays empty).
function selectDefault(f: Field): string {
  return f.options?.[0]?.value ?? ''
}

function defaultValues(cfg: ResourceConfig): Record<string, any> {
  const v: Record<string, any> = {}
  for (const f of cfg.fields) {
    if (f.type === 'select') v[f.name] = selectDefault(f)
  }
  return v
}

function seedValues(cfg: ResourceConfig, item: Record<string, any>): Record<string, any> {
  const v = { ...item }
  for (const f of cfg.fields) {
    if (f.type === 'money') v[f.name] = centsToDollars(item[f.name] ?? 0)
    else if (f.type === 'select' && (v[f.name] === undefined || v[f.name] === '')) v[f.name] = selectDefault(f)
  }
  return v
}

function buildPayload(cfg: ResourceConfig, values: Record<string, any>): Record<string, any> {
  const p = { ...values }
  for (const f of cfg.fields) {
    if (f.type === 'money') p[f.name] = dollarsToCents(String(values[f.name] ?? ''))
    else if (f.type === 'number') p[f.name] = Number(values[f.name]) || 0
    else if (f.type === 'bool') p[f.name] = !!values[f.name]
    else if (f.type === 'multiCheck') p[f.name] = Array.isArray(values[f.name]) ? values[f.name] : []
  }
  return p
}

// uploadStaged best-effort uploads photos staged on the New form to a just-created
// item. It deliberately never throws: the item already exists, so a failed photo
// must not bubble up and tempt a re-submit (that would create a duplicate item).
// Returns how many uploads failed.
async function uploadStaged(resource: string, id: number, staged: StagedPhoto[]): Promise<number> {
  let failed = 0
  for (const p of staged) {
    try {
      await uploadPhoto(resource, id, p.blob, p.name)
    } catch {
      failed++
    }
  }
  return failed
}

export default function ResourceForm() {
  const { resource, id } = useParams()
  const cfg = resource ? RESOURCES[resource] : undefined
  const editing = id != null
  const nav = useNavigate()
  const [values, setValues] = useState<Record<string, any>>({})
  const [firearms, setFirearms] = useState<Firearm[]>([])
  const [staged, setStaged] = useState<StagedPhoto[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    if (!resource || !RESOURCES[resource]) return
    const c = RESOURCES[resource]
    if (editing && id) getItem<Item>(resource, Number(id)).then((it) => setValues(seedValues(c, it))).catch(() => {})
    else setValues(defaultValues(c))
    if (c.fields.some((f) => f.type === 'firearmRef')) {
      listItems<Firearm>('firearms').then(setFirearms).catch(() => {})
    }
  }, [resource, id, editing])

  if (!cfg) return <Navigate to="/" replace />

  const set = (name: string, v: any) => setValues((prev) => ({ ...prev, [name]: v }))

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setErr(null)
    try {
      const payload = buildPayload(cfg, values)
      if (editing && id) {
        await updateItem(resource!, Number(id), payload)
        nav(`/${resource}`)
        return
      }
      // New item: create it first (photos attach by id), then upload any staged
      // photos against the new record. Land on its detail page afterward so the
      // photos are visible — and recoverable there if an upload failed.
      const created = await createItem<Item>(resource!, payload)
      if (staged.length > 0) {
        const failed = await uploadStaged(resource!, created.id, staged)
        if (failed > 0) {
          alert(
            `Saved, but ${failed} photo${failed === 1 ? '' : 's'} couldn't be uploaded. ` +
              `You can add ${failed === 1 ? 'it' : 'them'} from the item's page.`,
          )
        }
        nav(`/${resource}/${created.id}`)
        return
      }
      nav(`/${resource}`)
    } catch (e2) {
      setErr(e2 instanceof Error ? e2.message : 'Save failed')
      setBusy(false)
    }
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <h2 className="text-xl font-semibold text-dracula-fg">
        {editing ? 'Edit' : 'New'} {cfg.singular}
      </h2>
      {cfg.key === 'firearms' && (
        <SpecLookup onFill={(fields) => setValues((prev) => ({ ...prev, ...fields }))} />
      )}
      {cfg.fields
        .filter((f) => !f.showIf || f.showIf(values))
        .map((f) => (
          <FieldInput key={f.name} field={f} value={values[f.name]} firearms={firearms} onChange={(v) => set(f.name, v)} />
        ))}
      {/* Photos can be attached straight from the New form. On Edit they're managed
          on the item's detail page, where the record already exists. */}
      {!editing && (
        <div className="rounded-2xl border border-dracula-current p-4">
          <PhotoStager photos={staged} onChange={setStaged} />
        </div>
      )}
      {err && <p className="text-sm text-dracula-red">{err}</p>}
      <div className="flex gap-2 pt-2">
        <button type="submit" disabled={busy} className="rounded-lg bg-dracula-purple px-4 py-2 font-medium text-dracula-bg disabled:opacity-50">
          {busy ? 'Saving…' : 'Save'}
        </button>
        <button type="button" onClick={() => nav(-1)} className="rounded-lg border border-dracula-current px-4 py-2 text-dracula-comment">
          Cancel
        </button>
      </div>
    </form>
  )
}

const inputCls =
  'w-full rounded-lg border border-dracula-current bg-dracula-current/40 px-3 py-2 text-dracula-fg outline-none focus:border-dracula-purple'

function FieldInput({
  field,
  value,
  firearms,
  onChange,
}: {
  field: Field
  value: any
  firearms: Firearm[]
  onChange: (v: any) => void
}) {
  const label = <span className="mb-1 block text-xs uppercase tracking-wider text-dracula-comment">{field.label}</span>

  switch (field.type) {
    case 'bool':
      return (
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={!!value} onChange={(e) => onChange(e.target.checked)} className="h-4 w-4" />
          <span className="text-dracula-fg">{field.label}</span>
        </label>
      )
    case 'textarea':
      return <label className="block">{label}<textarea value={value ?? ''} onChange={(e) => onChange(e.target.value)} rows={3} className={inputCls} /></label>
    case 'select':
      return (
        <label className="block">
          {label}
          <select value={value ?? ''} onChange={(e) => onChange(e.target.value)} className={inputCls}>
            {field.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        </label>
      )
    case 'firearmRef':
      return (
        <label className="block">
          {label}
          <select value={value ?? ''} onChange={(e) => onChange(e.target.value === '' ? null : Number(e.target.value))} className={inputCls}>
            <option value="">— none —</option>
            {firearms.map((f) => (
              <option key={f.id} value={f.id}>{f.nickname || `${f.manufacturer} ${f.model}`.trim() || `#${f.id}`}</option>
            ))}
          </select>
        </label>
      )
    case 'money':
      return <label className="block">{label}<input inputMode="decimal" value={value ?? ''} onChange={(e) => onChange(e.target.value)} placeholder="0.00" className={inputCls} /></label>
    case 'number':
      return <label className="block">{label}<input type="number" value={value ?? ''} onChange={(e) => onChange(e.target.value === '' ? '' : Number(e.target.value))} className={inputCls} /></label>
    case 'date':
      return <label className="block">{label}<input type="date" value={value ?? ''} onChange={(e) => onChange(e.target.value)} className={inputCls} /></label>
    case 'combo':
      return <ComboField field={field} value={value} onChange={onChange} />
    case 'multiCheck':
      return <MultiCheckField field={field} value={value} onChange={onChange} />
    default:
      return <label className="block">{label}<input type="text" value={value ?? ''} onChange={(e) => onChange(e.target.value)} className={inputCls} /></label>
  }
}

// ComboField is a dropdown of common values plus an "Other…" option that reveals
// a free-text input — so caliber/manufacturer/etc. are quick to pick but never
// locked to the list.
function ComboField({ field, value, onChange }: { field: Field; value: any; onChange: (v: any) => void }) {
  const options = field.options ?? []
  const inList = (v: any) => typeof v === 'string' && options.some((o) => o.value === v)
  const [other, setOther] = useState<boolean>(!!value && !inList(value))

  useEffect(() => {
    if (value && !inList(value)) setOther(true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value])

  return (
    <label className="block">
      <span className="mb-1 block text-xs uppercase tracking-wider text-dracula-comment">{field.label}</span>
      <select
        value={other ? '__other__' : (value ?? '')}
        onChange={(e) => {
          if (e.target.value === '__other__') {
            setOther(true)
            onChange('')
          } else {
            setOther(false)
            onChange(e.target.value)
          }
        }}
        className={inputCls}
      >
        <option value="">— select —</option>
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
        <option value="__other__">Other…</option>
      </select>
      {other && (
        <input
          value={value ?? ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Type it in"
          className={`${inputCls} mt-2`}
        />
      )}
    </label>
  )
}

// MultiCheckField is a set of toggle chips for picking several values (e.g. the
// shell lengths a shotgun's chamber supports). Stores a string[].
function MultiCheckField({ field, value, onChange }: { field: Field; value: any; onChange: (v: any) => void }) {
  const selected: string[] = Array.isArray(value) ? value : []
  const toggle = (v: string) =>
    onChange(selected.includes(v) ? selected.filter((x) => x !== v) : [...selected, v])

  return (
    <div className="block">
      <span className="mb-1 block text-xs uppercase tracking-wider text-dracula-comment">{field.label}</span>
      <div className="flex flex-wrap gap-2">
        {(field.options ?? []).map((o) => {
          const on = selected.includes(o.value)
          return (
            <button
              type="button"
              key={o.value}
              onClick={() => toggle(o.value)}
              className={`rounded-lg border px-3 py-1.5 text-sm ${
                on ? 'border-dracula-purple bg-dracula-purple/20 text-dracula-fg' : 'border-dracula-current text-dracula-comment'
              }`}
            >
              {o.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}
