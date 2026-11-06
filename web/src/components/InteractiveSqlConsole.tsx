import React, { useState, useEffect } from 'react';
import { ActiveSandboxInfo, QueryResult } from '../types/api';
import { Terminal, Play, Trash2, Clock, Database, AlertCircle, CheckCircle2 } from 'lucide-react';

interface Props {
  sandboxes?: ActiveSandboxInfo[];
  onRefreshSandboxes?: () => void;
  sandboxId?: string;
  engine?: string;
  onClose?: () => void;
}

export const InteractiveSqlConsole: React.FC<Props> = ({
  sandboxes = [],
  onRefreshSandboxes,
  sandboxId,
  engine = 'sqlite',
  onClose,
}) => {
  const [selectedId, setSelectedId] = useState<string>(sandboxId || (sandboxes[0]?.id ?? ''));
  const [query, setQuery] = useState<string>('SELECT name FROM sqlite_master WHERE type="table";');
  const [result, setResult] = useState<QueryResult | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (sandboxId) {
      setSelectedId(sandboxId);
    } else if (sandboxes.length > 0 && (!selectedId || !sandboxes.some((s) => s.id === selectedId))) {
      setSelectedId(sandboxes[0].id);
    }
  }, [sandboxes, selectedId, sandboxId]);

  const activeSandbox = sandboxes.find((s) => s.id === selectedId);

  const handleExecute = async () => {
    const targetId = sandboxId || selectedId;
    if (!targetId || !query.trim()) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/sandboxes/${targetId}/query`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Query execution failed');
      }
      const data: QueryResult = await res.json();
      setResult(data);
    } catch (err: any) {
      setError(err.message);
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  const handleDestroy = async () => {
    const targetId = sandboxId || selectedId;
    if (!targetId) return;
    try {
      await fetch(`/api/sandboxes/${targetId}`, { method: 'DELETE' });
      onRefreshSandboxes?.();
      onClose?.();
      setResult(null);
    } catch (err) {
      console.error('Failed to prune sandbox:', err);
    }
  };

  const currentEngine = sandboxId ? engine : (activeSandbox?.engine || 'sqlite');

  const quickQueries = currentEngine === 'sqlite' ? [
    { label: 'List Tables', sql: 'SELECT name FROM sqlite_master WHERE type="table";' },
    { label: 'Sample Users', sql: 'SELECT * FROM users LIMIT 10;' },
    { label: 'Sample Orders', sql: 'SELECT * FROM orders LIMIT 10;' },
    { label: 'Foreign Key Check', sql: 'PRAGMA foreign_key_check;' },
    { label: 'Integrity Check', sql: 'PRAGMA integrity_check;' },
  ] : [
    { label: 'List Tables', sql: 'SELECT table_name FROM information_schema.tables WHERE table_schema="public";' },
    { label: 'Sample Rows', sql: 'SELECT * FROM users LIMIT 10;' },
  ];

  const currentId = sandboxId || selectedId;

  if (!currentId && sandboxes.length === 0) {
    return (
      <div className="bg-surface border border-border/70 rounded-xl p-8 text-center text-slate-400">
        <Database className="w-10 h-10 mx-auto text-slate-600 mb-3" />
        <h4 className="text-sm font-semibold text-slate-300">No Ephemeral Sandboxes Held Open</h4>
        <p className="text-xs text-slate-500 mt-1 max-w-md mx-auto">
          To debug a restored database interactively, enable "Hold Sandbox for Debugging" when launching a drill.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-surface border border-border/80 rounded-xl overflow-hidden shadow-xl font-mono text-xs">
      {/* Header & Sandbox Picker */}
      <div className="p-4 bg-surface-elevated/60 border-b border-border flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Terminal className="w-4 h-4 text-blue-400" />
          <span className="font-bold text-slate-200 uppercase tracking-wider text-xs font-sans">
            Interactive SQL Console
          </span>

          {sandboxes.length > 0 && !sandboxId ? (
            <select
              value={selectedId}
              onChange={(e) => {
                setSelectedId(e.target.value);
                setResult(null);
              }}
              className="bg-slate-900 border border-slate-700 rounded px-2.5 py-1 text-slate-200 text-xs focus:outline-none focus:border-blue-500"
            >
              {sandboxes.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.id} ({s.engine})
                </option>
              ))}
            </select>
          ) : (
            <span className="bg-slate-900 border border-slate-700 rounded px-2.5 py-1 text-blue-300">
              {currentId} ({currentEngine})
            </span>
          )}
        </div>

        <div className="flex items-center gap-3">
          {activeSandbox && (
            <div className="flex items-center gap-1.5 text-amber-400 bg-amber-500/10 border border-amber-500/30 px-2 py-0.5 rounded text-[11px]">
              <Clock className="w-3 h-3" />
              <span>Expires {new Date(activeSandbox.hold_until).toLocaleTimeString()}</span>
            </div>
          )}
          <button
            type="button"
            onClick={handleDestroy}
            className="inline-flex items-center gap-1 text-[11px] px-2.5 py-1 rounded bg-rose-500/10 text-rose-300 hover:bg-rose-500/20 border border-rose-500/30 transition-colors"
          >
            <Trash2 className="w-3 h-3" />
            Prune Sandbox
          </button>
        </div>
      </div>

      {/* Query Bar */}
      <div className="p-4 space-y-3">
        {/* Quick Query Pills */}
        <div className="flex flex-wrap items-center gap-1.5 font-sans">
          <span className="text-[11px] text-slate-400 font-mono mr-1">Presets:</span>
          {quickQueries.map((q, i) => (
            <button
              key={i}
              type="button"
              onClick={() => setQuery(q.sql)}
              className="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 text-[11px] transition-colors"
            >
              {q.label}
            </button>
          ))}
        </div>

        {/* Input & Run */}
        <div className="flex items-start gap-2">
          <textarea
            rows={2}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter read-only SQL query (SELECT, PRAGMA, EXPLAIN)..."
            className="flex-1 bg-slate-950 border border-slate-700 rounded-lg p-2.5 text-slate-200 placeholder-slate-600 focus:outline-none focus:border-blue-500 font-mono text-xs resize-none"
          />
          <button
            type="button"
            onClick={handleExecute}
            disabled={loading}
            className="px-4 py-3 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-bold transition-colors flex items-center gap-1.5 disabled:opacity-50 h-full font-sans text-xs"
          >
            <Play className="w-3.5 h-3.5 fill-current" />
            {loading ? 'Running...' : 'Run Query'}
          </button>
        </div>

        {/* Error Notification */}
        {error && (
          <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-300 flex items-center gap-2">
            <AlertCircle className="w-4 h-4 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Results Table */}
        {result && (
          <div className="border border-border/70 rounded-xl overflow-hidden bg-slate-950/60 mt-3">
            <div className="px-3 py-1.5 bg-slate-900 border-b border-border/60 flex items-center justify-between text-[11px] text-slate-400">
              <span className="flex items-center gap-1.5 text-emerald-400">
                <CheckCircle2 className="w-3.5 h-3.5" />
                {result.row_count} rows returned
              </span>
              <span>Execution Time: {result.duration_ms} ms</span>
            </div>

            <div className="overflow-x-auto max-h-60">
              <table className="w-full text-left text-[11px]">
                <thead className="bg-slate-900 text-slate-300 border-b border-border/60">
                  <tr>
                    <th className="py-2 px-3 text-slate-500 w-10">#</th>
                    {result.columns.map((col, i) => (
                      <th key={i} className="py-2 px-3 font-semibold">
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/30 text-slate-300">
                  {result.rows.map((row, rIdx) => (
                    <tr key={rIdx} className="hover:bg-slate-800/40">
                      <td className="py-1.5 px-3 text-slate-600 select-none">{rIdx + 1}</td>
                      {row.map((cell, cIdx) => (
                        <td key={cIdx} className="py-1.5 px-3 truncate max-w-xs">
                          {cell === 'NULL' ? <span className="text-slate-600 italic">NULL</span> : cell}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
