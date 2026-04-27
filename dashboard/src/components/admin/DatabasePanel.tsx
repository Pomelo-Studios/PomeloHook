import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { TableInfo, TableResult, QueryResult } from '../../types'

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
  const [confirm, setConfirm] = useState<{ sql: string; onConfirm: () => void } | null>(null)
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
      setConfirm({ sql: trimmed, onConfirm: execQuery })
    } else {
      execQuery()
    }
  }

  async function execQuery() {
    setConfirm(null)
    setQueryError('')
    setQueryResult(null)
    try {
      const result = await api.admin.runQuery(apiKey, sql)
      setQueryResult(result)
    } catch (err) {
      setQueryError(err instanceof Error ? err.message : 'Query failed')
    }
  }

  const tabBtn = (id: 'tables' | 'sql', label: string) => (
    <button onClick={() => setTab(id)}
      className={`text-xs px-4 py-2 border-b-2 transition-colors ${tab === id ? 'text-emerald-400 border-emerald-500' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}>
      {label}
    </button>
  )

  return (
    <div className="flex flex-col h-full">
      {confirm && (
        <ConfirmDialog
          message="This query will modify the database."
          detail={confirm.sql}
          onConfirm={confirm.onConfirm}
          onCancel={() => setConfirm(null)}
        />
      )}
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Database</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      <div className="flex border-b border-zinc-800 flex-shrink-0 bg-zinc-900/20">
        {tabBtn('tables', 'Tables')}
        {tabBtn('sql', 'SQL')}
      </div>

      {tab === 'tables' && (
        <div className="flex flex-1 overflow-hidden">
          <div className="w-40 border-r border-zinc-800 overflow-y-auto flex-shrink-0 py-2">
            <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-3 pb-2">Tables</p>
            {tables.map(t => (
              <button key={t.name} onClick={() => loadTable(t.name, 0)}
                className={`w-full text-left px-3 py-1.5 text-[10px] flex justify-between items-center hover:bg-zinc-800/50 ${selectedTable === t.name ? 'text-emerald-400 bg-zinc-800/50' : 'text-zinc-500'}`}>
                <span>{t.name}</span>
                <span className="text-zinc-600">{t.row_count}</span>
              </button>
            ))}
          </div>
          <div className="flex-1 flex flex-col overflow-hidden">
            {tableData ? (
              <>
                <div className="flex-1 overflow-auto">
                  <table className="w-full border-collapse">
                    <thead className="sticky top-0">
                      <tr className="bg-zinc-900/80">
                        {tableData.columns.map(c => (
                          <th key={c} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold whitespace-nowrap">{c}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {tableData.rows.map((row, i) => (
                        <tr key={i} className="hover:bg-zinc-900/40">
                          {row.map((cell, j) => (
                            <td key={j} className="px-3 py-1.5 text-[10px] text-zinc-400 font-mono border-b border-zinc-800/50 max-w-[200px] truncate">{String(cell ?? '')}</td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="border-t border-zinc-800 px-4 py-2 flex items-center gap-3 flex-shrink-0">
                  <span className="text-[10px] text-zinc-600">{tableData.rows.length} rows</span>
                  {offset > 0 && <button onClick={() => loadTable(selectedTable!, offset - 200)} className="text-[10px] text-zinc-500 hover:text-zinc-300">← prev</button>}
                  {tableData.rows.length === 200 && <button onClick={() => loadTable(selectedTable!, offset + 200)} className="text-[10px] text-zinc-500 hover:text-zinc-300">next →</button>}
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-full text-zinc-600 text-xs">Select a table</div>
            )}
          </div>
        </div>
      )}

      {tab === 'sql' && (
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="p-3 border-b border-zinc-800 flex-shrink-0">
            <textarea
              value={sql}
              onChange={e => setSql(e.target.value)}
              rows={4}
              placeholder="SELECT * FROM users LIMIT 10"
              className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-200 font-mono outline-none focus:border-zinc-600 resize-none placeholder-zinc-700"
            />
            <div className="flex items-center gap-2 mt-2">
              <button onClick={runQuery} className="text-[10px] px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">▶ Run</button>
              {WRITE_RE.test(sql.trim()) && (
                <span className="text-[9px] text-amber-500 bg-amber-950/50 border border-amber-900/50 px-2 py-0.5 rounded">⚠ write operation</span>
              )}
            </div>
          </div>
          <div className="flex-1 overflow-auto">
            {queryError && <div className="p-3 text-red-400 text-xs font-mono">{queryError}</div>}
            {queryResult && (
              queryResult.columns.length > 0 ? (
                <table className="w-full border-collapse">
                  <thead className="sticky top-0">
                    <tr className="bg-zinc-900/80">
                      {queryResult.columns.map(c => (
                        <th key={c} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold whitespace-nowrap">{c}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {(queryResult.rows ?? []).map((row, i) => (
                      <tr key={i} className="hover:bg-zinc-900/40">
                        {row.map((cell, j) => (
                          <td key={j} className="px-3 py-1.5 text-[10px] text-zinc-400 font-mono border-b border-zinc-800/50 max-w-[200px] truncate">{String(cell ?? '')}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-3 text-zinc-500 text-xs">{queryResult.affected} row(s) affected</div>
              )
            )}
          </div>
        </div>
      )}
    </div>
  )
}
