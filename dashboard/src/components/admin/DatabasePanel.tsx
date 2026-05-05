import { useState, useEffect } from 'react'
import { Play, AlertTriangle } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { TableInfo, TableResult, QueryResult, ConfirmState } from '../../types'

interface Props { apiKey: string }

const WRITE_RE = /^\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|REPLACE|TRUNCATE)\b/i

export function DatabasePanel({ apiKey }: Props) {
  const [tab, setTab] = useState<'tables' | 'sql'>('tables')
  const [tables, setTables] = useState<TableInfo[]>([])
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [tableData, setTableData] = useState<TableResult | null>(null)
  const [offset, setOffset] = useState(0)
  const [sql, setSql] = useState('')
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null)
  const [queryError, setQueryError] = useState('')
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.listTables(apiKey).then(setTables).catch(() => setError('Failed to load tables'))
  }, [apiKey])

  function loadTable(name: string, off = 0) {
    setSelectedTable(name)
    setOffset(off)
    setTableData(null)
    api.admin.getTableRows(apiKey, name, 200, off).then(setTableData).catch(() => setError('Failed to load rows'))
  }

  function runQuery() {
    const trimmed = sql.trim()
    if (!trimmed) return
    if (WRITE_RE.test(trimmed)) {
      setConfirm({ message: 'This query will modify the database.', detail: trimmed, onConfirm: execQuery })
    } else {
      execQuery()
    }
  }

  async function execQuery() {
    setConfirm(null)
    setQueryError('')
    setQueryResult(null)
    try {
      setQueryResult(await api.admin.runQuery(apiKey, sql))
    } catch (err) {
      setQueryError(err instanceof Error ? err.message : 'Query failed')
    }
  }

  const isWrite = WRITE_RE.test(sql.trim())

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div className="text-[14px] font-bold" style={{ color: 'var(--text)' }}>Database</div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      <div className="flex flex-shrink-0 border-b" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
        {(['tables', 'sql'] as const).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className="text-xs px-5 py-3 border-b-2 font-medium capitalize transition-colors"
            style={
              tab === t
                ? { color: 'var(--coral)', borderBottomColor: 'var(--coral)' }
                : { color: 'var(--text-2)', borderBottomColor: 'transparent' }
            }
          >
            {t === 'sql' ? 'SQL' : 'Tables'}
          </button>
        ))}
      </div>

      {tab === 'tables' && (
        <div className="flex flex-1 overflow-hidden">
          <div className="w-40 overflow-y-auto flex-shrink-0 py-2 border-r" style={{ borderColor: 'var(--border)' }}>
            <p className="text-[9px] font-bold tracking-[1.5px] uppercase px-3 pb-2" style={{ color: 'var(--text-3)' }}>Tables</p>
            {tables.map(t => (
              <button
                key={t.name}
                onClick={() => loadTable(t.name, 0)}
                className="w-full text-left px-3 py-[6px] text-[11px] flex justify-between items-center transition-colors"
                style={{
                  color: selectedTable === t.name ? 'var(--coral)' : 'var(--text-2)',
                  background: selectedTable === t.name ? 'var(--selected-bg)' : 'transparent',
                }}
              >
                <span>{t.name}</span>
                <span style={{ color: 'var(--text-3)' }}>{t.row_count}</span>
              </button>
            ))}
          </div>

          <div className="flex-1 flex flex-col overflow-hidden">
            {tableData ? (
              <>
                <div className="flex-1 overflow-auto">
                  <table className="w-full border-collapse">
                    <thead className="sticky top-0">
                      <tr style={{ background: 'var(--surface)' }}>
                        {tableData.columns.map(c => (
                          <th key={c} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b whitespace-nowrap" style={{ color: 'var(--text-3)', borderColor: 'var(--border)' }}>
                            {c}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {tableData.rows.map((row, i) => (
                        <tr key={i} style={{ borderBottom: '1px solid var(--border)' }}>
                          {row.map((cell, j) => (
                            <td key={j} title={String(cell ?? '')} className="px-4 py-2 text-[10px] font-mono max-w-[200px] truncate" style={{ color: 'var(--text-2)' }}>
                              {String(cell ?? '')}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="px-5 py-2 flex items-center gap-3 flex-shrink-0 border-t" style={{ borderColor: 'var(--border)' }}>
                  <span className="text-[10px]" style={{ color: 'var(--text-3)' }}>{tableData.rows.length} rows</span>
                  {offset > 0 && (
                    <button onClick={() => loadTable(selectedTable!, offset - 200)} className="text-[10px]" style={{ color: 'var(--text-2)' }}>← prev</button>
                  )}
                  {tableData.rows.length === 200 && (
                    <button onClick={() => loadTable(selectedTable!, offset + 200)} className="text-[10px]" style={{ color: 'var(--text-2)' }}>next →</button>
                  )}
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-full text-xs" style={{ color: 'var(--text-3)' }}>Select a table</div>
            )}
          </div>
        </div>
      )}

      {tab === 'sql' && (
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="p-4 border-b flex-shrink-0" style={{ borderColor: 'var(--border)' }}>
            <textarea
              value={sql}
              onChange={e => setSql(e.target.value)}
              rows={4}
              placeholder="SELECT * FROM users LIMIT 10"
              className="w-full rounded-lg px-3 py-2 text-xs font-mono outline-none resize-none"
              style={{ background: 'var(--code-bg)', border: '1px solid var(--code-border)', color: 'var(--text)' }}
            />
            <div className="flex items-center gap-2 mt-2">
              <button
                onClick={runQuery}
                className="flex items-center gap-[6px] rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity hover:opacity-90"
                style={{ background: 'var(--coral)', color: 'white' }}
              >
                <Play size={11} fill="white" strokeWidth={0} />
                Run
              </button>
              {isWrite && (
                <span
                  className="flex items-center gap-1 text-[9px] px-2 py-[2px] rounded"
                  style={{ color: 'var(--orange)', background: 'rgba(255,163,73,0.1)', border: '1px solid rgba(255,163,73,0.3)' }}
                >
                  <AlertTriangle size={10} /> write operation
                </span>
              )}
            </div>
          </div>
          <div className="flex-1 overflow-auto">
            {queryError && <div className="p-4 text-xs font-mono" style={{ color: 'var(--err-text)' }}>{queryError}</div>}
            {queryResult && (
              queryResult.columns.length > 0 ? (
                <table className="w-full border-collapse">
                  <thead className="sticky top-0">
                    <tr style={{ background: 'var(--surface)' }}>
                      {queryResult.columns.map(c => (
                        <th key={c} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b whitespace-nowrap" style={{ color: 'var(--text-3)', borderColor: 'var(--border)' }}>
                          {c}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {(queryResult.rows ?? []).map((row, i) => (
                      <tr key={i} style={{ borderBottom: '1px solid var(--border)' }}>
                        {row.map((cell, j) => (
                          <td key={j} className="px-4 py-2 text-[10px] font-mono max-w-[200px] truncate" style={{ color: 'var(--text-2)' }}>
                            {String(cell ?? '')}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-4 text-xs" style={{ color: 'var(--text-2)' }}>No results</div>
              )
            )}
          </div>
        </div>
      )}
    </div>
  )
}
