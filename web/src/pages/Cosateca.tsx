import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../api'
import { usePermissions } from '../hooks/usePermissions'
import { Plus, Edit2, Trash2, Check, X, Clock, Loader2, Search, Filter, Tool, CreditCard, RefreshCw } from 'lucide-react'
import { fmtTQ } from '../lib/format'

type CosatecaObject = {
  id: string
  name: string
  description?: string
  category: string
  condition: string
  status: string
  location?: string
  design_ref?: string
  design_version?: string
  tq_value: number
  contributed_by: string
  contributor_name?: string
  federated_visible: boolean
  current_loan_id?: string | null
  current_borrower?: string | null
  loan_due_at?: string | null
  loan_state?: string | null
}

type CosatecaLoan = {
  id: string
  object_id: string
  object_name: string
  borrower_id: string
  borrower_name: string
  loaned_at: string
  due_at?: string | null
  state: string
}

const STATUS_LABELS: Record<string, string> = {
  disponible: 'Disponible',
  prestado: 'Prestado',
  mantenimiento: 'Mantenimiento',
  retirado: 'Retirado',
}

const STATUS_COLORS: Record<string, string> = {
  disponible: 'bg-trueque-100 text-trueque-800',
  prestado: 'bg-amber-100 text-amber-800',
  mantenimiento: 'bg-rose-100 text-rose-800',
  retirado: 'bg-slate-100 text-slate-800',
}

export default function Cosateca() {
  const { t } = useTranslation(['cosateca', 'common'])
  const { hasPermission } = usePermissions()
  const [tab, setTab] = useState<'catalogo' | 'prestamos' | 'aportar' | 'mantenimiento'>('catalogo')
  const [objects, setObjects] = useState<CosatecaObject[]>([])
  const [loans, setLoans] = useState<CosatecaLoan[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  // --- Formulario: aportar objeto ---
  const [showAportar, setShowAportar] = useState(false)
  const [form, setForm] = useState({
    name: '',
    description: '',
    category: 'herramienta',
    condition: 'bueno',
    location: '',
    design_ref: '',
    tq_value: 0,
  })

  // --- Formulario: prestamo ---
  const [showPrestar, setShowPrestar] = useState<string | null>(null)
  const [dueAt, setDueAt] = useState('')

  // --- Formulario: mantenimiento ---
  const [showMantener, setShowMantener] = useState<string | null>(null)
  const [mantTipo, setMantTipo] = useState('reparacion')
  const [mantDesc, setMantDesc] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [oRes, lRes] = await Promise.all([
        api.get<{ objects: CosatecaObject[]; count: number }>('/api/cosateca/objects'),
        api.get<{ loans: CosatecaLoan[]; count: number }>('/api/cosateca/loans?mine=1'),
      ])
      setObjects(oRes.objects)
      setLoans(lRes.loans)
    } catch (e: any) {
      setError(e.message || 'Error cargando la cosateca')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleAportar = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/api/cosateca/objects', {
        name: form.name,
        description: form.description,
        category: form.category,
        condition: form.condition,
        location: form.location,
        design_ref: form.design_ref,
        tq_value: Number(form.tq_value) || 0,
      })
      setSuccess('Objeto aportado a la cosateca')
      setShowAportar(false)
      setForm({ name: '', description: '', category: 'herramienta', condition: 'bueno', location: '', design_ref: '', tq_value: 0 })
      load()
    } catch (e: any) {
      setError(e.message || 'Error aportando objeto')
    }
  }

  const prestarObjeto = async (objectId: string) => {
    try {
      await api.post('/api/cosateca/loans', { object_id: objectId, due_at: dueAt || undefined })
      setSuccess('Prestamo registrado')
      setShowPrestar(null)
      setDueAt('')
      load()
    } catch (e: any) {
      setError(e.message || 'Error registrando prestamo')
    }
  }

  const devolverObjeto = async (loanId: string) => {
    if (!window.confirm(t('cosateca:confirm_return', { defaultMessage: '¿Confirmar devolución del objeto?' }))) return
    try {
      await api.post(`/api/cosateca/loans/${loanId}/return`, {})
      setSuccess('Objeto devuelto')
      load()
    } catch (e: any) {
      setError(e.message || 'Error devolviendo objeto')
    }
  }

  const registrarMantenimiento = async (objectId: string) => {
    try {
      await api.post('/api/cosateca/maintenance', {
        object_id: objectId,
        type: mantTipo,
        description: mantDesc,
      })
      setSuccess('Mantenimiento registrado')
      setShowMantener(null)
      setMantTipo('reparacion')
      setMantDesc('')
      load()
    } catch (e: any) {
      setError(e.message || 'Error registrando mantenimiento')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">{t('cosateca:title')}</h1>
          <p className="text-muted-foreground">{t('cosateca:subtitle')}</p>
        </div>
        <button
          onClick={() => setTab('aportar')}
          className="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-md border bg-trueque-600 text-white hover:bg-trueque-700 disabled:opacity-50"
          disabled={!hasPermission('cosateca.manage')}
        >
          <Plus size={16} />
          {t('cosateca:add_object')}
        </button>
      </div>

      {error && (
        <div className="rounded-md border border-rose-200 bg-rose-50 p-3 text-rose-800 text-sm">{error}</div>
      )}
      {success && (
        <div className="rounded-md border border-trueque-200 bg-trueque-50 p-3 text-trueque-800 text-sm">{success}</div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border">
        {['catalogo', 'prestamos', 'aportar', 'mantenimiento'].map((key) => {
          const label = t(`cosateca:tab_${key}` as any, { defaultMessage: key })
          const active = tab === key
          return (
            <button
              key={key}
              onClick={() => setTab(key as any)}
              className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium transition-colors ${
                active
                  ? 'border-trueque-600 text-trueque-700'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-muted'
              }`}
            >
              {label}
            </button>
          )
        })}
      </div>

      {/* CATALOGO */}
      {tab === 'catalogo' && (
        <section className="mt-4">
          <div className="flex flex-col sm:flex-row gap-3 items-end mb-4">
            <div className="relative w-full sm:w-64">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder={t('cosateca:search', { defaultMessage: 'Buscar objeto...' })}
                className="w-full pl-10 pr-3 py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
              />
            </div>
            <select
              className="w-full sm:w-48 py-1.5 text-sm border rounded-md bg-background"
              defaultValue=""
            >
              <option value="">{t('cosateca:all_statuses', { defaultMessage: 'Todos los estados' })}</option>
              <option value="disponible">{t('cosateca:status_disponible')}</option>
              <option value="prestado">{t('cosateca:status_prestado')}</option>
              <option value="mantenimiento">{t('cosateca:status_mantenimiento')}</option>
            </select>
          </div>
          {loading ? (
            <div className="h-64 flex items-center justify-center"><Loader2 size={24} className="animate-spin" /></div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {objects.length === 0 && (
                <div className="col-span-3 p-8 text-center text-muted-foreground">{t('cosateca:no_objects')}</div>
              )}
              {objects.map((o) => (
                <div key={o.id} className="border rounded-lg p-3">
                  <div className="flex justify-between items-start gap-2">
                    <h3 className="font-semibold">{o.name}</h3>
                    <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${STATUS_COLORS[o.status] || 'bg-muted'}`}>
                      {STATUS_LABELS[o.status] || o.status}
                    </span>
                  </div>
                  {o.description && <p className="text-sm text-muted-foreground mt-1">{o.description}</p>}
                  <div className="flex flex-wrap gap-1 mt-2">
                    <span className="text-xs border px-1.5 py-0.5 rounded">{o.category}</span>
                    <span className="text-xs border px-1.5 py-0.5 rounded">{o.condition}</span>
                    {o.design_ref && (
                      <span className="text-xs border px-1.5 py-0.5 rounded" title={o.design_ref}>
                        OAD: {o.design_ref.length > 20 ? o.design_ref.slice(0, 17) + '...' : o.design_ref}
                      </span>
                    )}
                  </div>
                  {o.contributor_name && (
                    <p className="text-xs text-muted-foreground mt-2">
                      {t('cosateca:contributed_by', { defaultMessage: 'Aportado por {name}', values: { name: o.contributor_name } })}
                    </p>
                  )}
                  {o.tq_value && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {t('cosateca:tq_value', { defaultMessage: 'Valor incorporado: {value} TQ', values: { value: fmtTQ(o.tq_value) } })}
                    </p>
                  )}
                  <div className="mt-3 flex gap-2">
                    {o.status === 'disponible' && (
                      <button
                        onClick={() => setShowPrestar(o.id)}
                        className="flex-1 px-3 py-1.5 text-sm font-medium rounded-md border bg-trueque-600 text-white hover:bg-trueque-700"
                      >
                        {t('cosateca:lend', { defaultMessage: 'Prestar' })}
                      </button>
                    )}
                    {o.current_loan_id && (
                      <button
                        onClick={() => setTab('prestamos')}
                        className="px-3 py-1.5 text-sm font-medium rounded-md border text-muted-foreground hover:bg-muted"
                      >
                        {t('cosateca:prestado_a', { defaultMessage: 'Prestado a {name}', values: { name: o.current_borrower || '?' } })}
                      </button>
                    )}
                    {o.status === 'mantenimiento' && (
                      <button
                        onClick={() => setShowMantener(o.id)}
                        className="px-3 py-1.5 text-sm font-medium rounded-md border text-muted-foreground hover:bg-muted"
                      >
                        {t('cosateca:maintain', { defaultMessage: 'Mantener' })}
                      </button>
                    )}
                  </div>
                  {/* Prestar dialogo inline */}
                  {showPrestar === o.id && (
                    <div className="mt-3 p-3 border rounded-lg bg-muted/50 space-y-2">
                      <label className="text-xs font-medium block">{t('cosateca:return_by', { defaultMessage: 'Devolver antes del' })}</label>
                      <input
                        type="datetime-local"
                        value={dueAt}
                        onChange={(e) => setDueAt(e.target.value)}
                        className="w-full py-1 text-sm border rounded-md"
                      />
                      <div className="flex gap-2">
                        <button
                          onClick={() => prestarObjeto(o.id)}
                          disabled={!dueAt}
                          className="flex-1 py-1.5 text-sm font-medium rounded-md border bg-trueque-600 text-white disabled:opacity-50"
                        >
                          {t('cosateca:confirm_lend', { defaultMessage: 'Confirmar prestamo' })}
                        </button>
                        <button
                          onClick={() => setShowPrestar(null)}
                          className="flex-1 py-1.5 text-sm font-medium rounded-md border text-muted-foreground hover:bg-muted"
                        >
                          {t('common:cancel', { defaultMessage: 'Cancelar' })}
                        </button>
                      </div>
                    </div>
                  )}
                  {/* Mantener dialogo inline */}
                  {showMantener === o.id && (
                    <div className="mt-3 p-3 border rounded-lg bg-muted/50 space-y-2">
                      <label className="text-xs font-medium block">{t('cosateca:maintenance_type')}</label>
                      <select
                        value={mantTipo}
                        onChange={(e) => setMantTipo(e.target.value)}
                        className="w-full py-1 text-sm border rounded-md"
                      >
                        <option value="reparacion">{t('cosateca:maint_repair')}</option>
                        <option value="limpieza">{t('cosateca:maint_cleaning')}</option>
                        <option value="calibracion">{t('cosateca:maint_calibration')}</option>
                        <option value="revision">{t('cosateca:maint_review')}</option>
                      </select>
                      <label className="text-xs font-medium block mt-2">{t('cosateca:description')}</label>
                      <textarea
                        value={mantDesc}
                        onChange={(e) => setMantDesc(e.target.value)}
                        rows={2}
                        className="w-full py-1 text-sm border rounded-md"
                      />
                      <div className="flex gap-2 mt-2">
                        <button
                          onClick={() => registrarMantenimiento(o.id)}
                          className="flex-1 py-1.5 text-sm font-medium rounded-md border bg-trueque-600 text-white"
                        >
                          {t('cosateca:register', { defaultMessage: 'Registrar' })}
                        </button>
                        <button
                          onClick={() => setShowMantener(null)}
                          className="flex-1 py-1.5 text-sm font-medium rounded-md border text-muted-foreground hover:bg-muted"
                        >
                          {t('common:cancel', { defaultMessage: 'Cancelar' })}
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* PRESTAMOS MÍOS */}
      {tab === 'prestamos' && (
        <section className="mt-4">
          <div className="flex gap-2">
            <button
              onClick={() => load()}
              className="px-3 py-1.5 text-sm font-medium rounded-md border bg-background text-muted-foreground hover:bg-muted"
            >
              <RefreshCw size={16} /> {t('cosateca:refresh')}
            </button>
          </div>
          <div className="mt-3">
            {loans.length === 0 ? (
              <p className="text-muted-foreground text-sm">{t('cosateca:no_loans')}</p>
            ) : (
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="text-muted-foreground border-b">
                    <th className="text-left py-2">{t('cosateca:object')}</th>
                    <th className="text-left py-2">{t('cosateca:loaned_at')}</th>
                    <th className="text-left py-2">{t('cosateca:due_at')}</th>
                    <th className="text-left py-2">{t('cosateca:state')}</th>
                    <th className="text-right py-2"></th>
                  </tr>
                </thead>
                <tbody>
                  {loans.map((l) => (
                    <tr key={l.id} className="border-b last:border-0">
                      <td className="py-2">{l.object_name}</td>
                      <td className="py-2">{new Date(l.loaned_at).toLocaleDateString()}</td>
                      <td className="py-2">{l.due_at ? new Date(l.due_at).toLocaleDateString() : '-'}</td>
                      <td className="py-2">
                        <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${
                          l.state === 'activo' ? 'bg-amber-100 text-amber-800' : 'bg-trueque-100 text-trueque-800'
                        }`}>
                          {l.state === 'activo' ? t('cosateca:active') : t('cosateca:returned')}
                        </span>
                      </td>
                      <td className="text-right py-2">
                        {l.state === 'activo' && (
                          <button
                            onClick={() => devolverObjeto(l.id)}
                            className="px-2 py-1 text-xs font-medium rounded border text-rose-700 border-rose-200 hover:bg-rose-50"
                          >
                            {t('cosateca:return_obj', { defaultMessage: 'Devolver' })}
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </section>
      )}

      {/* APORTAR OBJETO */}
      {tab === 'aportar' && (
        <section className="mt-4 max-w-md">
          <form onSubmit={handleAportar} className="space-y-3 border p-4 rounded-lg">
            <h2 className="text-lg font-semibold mb-1">{t('cosateca:add_object')}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-semibold mb-1">{t('cosateca:name')}</label>
                <input
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder={t('cosateca:placeholder_name')}
                  className="w-full py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold mb-1">{t('cosateca:category')}</label>
                <select
                  value={form.category}
                  onChange={(e) => setForm({ ...form, category: e.target.value })}
                  className="w-full py-1.5 text-sm border rounded-md"
                >
                  <option value="herramienta">{t('cosateca:cat_tool')}</option>
                  <option value="maquina">{t('cosateca:cat_machine')}</option>
                  <option value="utensilio">{t('cosateca:cat_utensil')}</option>
                  <option value="material">{t('cosateca:cat_material')}</option>
                  <option value="equipo_eventos">{t('cosateca:cat_events')}</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-semibold mb-1">{t('cosateca:condition')}</label>
                <select
                  value={form.condition}
                  onChange={(e) => setForm({ ...form, condition: e.target.value })}
                  className="w-full py-1.5 text-sm border rounded-md"
                >
                  <option value="bueno">{t('cosateca:cond_good')}</option>
                  <option value="regular">{t('cosateca:cond_fair')}</option>
                  <option value="malo">{t('cosateca:cond_bad')}</option>
                </select>
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs font-semibold mb-1">{t('cosateca:location')}</label>
                <input
                  value={form.location}
                  onChange={(e) => setForm({ ...form, location: e.target.value })}
                  placeholder={t('cosateca:placeholder_location')}
                  className="w-full py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs font-semibold mb-1">{t('cosateca:design_ref', { defaultMessage: 'Diseño OAD (referencia)' })}</label>
                <input
                  value={form.design_ref}
                  onChange={(e) => setForm({ ...form, design_ref: e.target.value })}
                  placeholder={t('cosateca:placeholder_design_ref')}
                  className="w-full py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
                />
                <p className="text-xs text-muted-foreground mt-1">{t('cosateca:design_ref_help')}</p>
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs font-semibold mb-1">{t('cosateca:description')}</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={2}
                  className="w-full py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold mb-1">{t('cosateca:tq_value_label', { defaultMessage: 'Valor incorporado (TQ)' })}</label>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    value={form.tq_value}
                    onChange={(e) => setForm({ ...form, tq_value: e.target.value })}
                    className="w-full py-1.5 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-trueque-500"
                  />
                  <span className="text-xs text-muted-foreground">TQ</span>
                </div>
                <p className="text-xs text-muted-foreground mt-1">{t('cosateca:tq_value_help')}</p>
              </div>
            </div>
            <div className="flex gap-2 mt-2 pt-2 border-t">
              <button
                type="submit"
                disabled={!form.name}
                className="flex-1 py-2 text-sm font-medium rounded-md border bg-trueque-600 text-white disabled:opacity-50"
              >
                {t('cosateca:save')}
              </button>
              <button
                type="button"
                onClick={() => setShowAportar(false)}
                className="flex-1 py-2 text-sm font-medium rounded-md border text-muted-foreground hover:bg-muted"
              >
                {t('common:cancel', { defaultMessage: 'Cancelar' })}
              </button>
            </div>
          </form>
        </section>
      )}

      {/* APORTACIONES Y CREDITO TQ */}
      {tab === 'mantenimiento' && (
        <section className="mt-4 max-w-2xl">
          <h2 className="text-lg font-semibold mb-3">{t('cosateca:contributions_title')}</h2>
          <p className="text-sm text-muted-foreground mb-4">
            {t('cosateca:contributions_help')}
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            <div className="bg-muted/50 p-3 rounded-lg">
              <h3 className="text-sm font-semibold mb-2">{t('cosateca:principle_object')}</h3>
              <p className="text-xs text-muted-foreground">{t('cosateca:principle_object_desc')}</p>
            </div>
            <div className="bg-muted/50 p-3 rounded-lg">
              <h3 className="text-sm font-semibold mb-2">{t('cosateca:principle_record')}</h3>
              <p className="text-xs text-muted-foreground">{t('cosateca:principle_record_desc')}</p>
            </div>
            <div className="bg-muted/50 p-3 rounded-lg md:col-span-2">
              <h3 className="text-sm font-semibold mb-2">{t('cosateca:principle_reciprocity')}</h3>
              <p className="text-xs text-muted-foreground">{t('cosateca:principle_reciprocity_desc')}</p>
            </div>
          </div>
        </section>
      )}
    </div>
  )
}
