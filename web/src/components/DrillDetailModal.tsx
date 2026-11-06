import React from 'react';
import { DrillRun } from '../types/api';
import { HealthScoreBadge } from './HealthScoreBadge';
import { X, Clock, Database, CheckCircle2, AlertTriangle, ShieldCheck } from 'lucide-react';

interface Props {
  drill: DrillRun | null;
  onClose: () => void;
  onViewCert: (drill: DrillRun) => void;
}

export const DrillDetailModal: React.FC<Props> = ({ drill, onClose, onViewCert }) => {
  if (!drill) return null;

  const total = drill.timings?.total_rto_ms || 1;
  const timings = drill.timings || {
    fetch_ms: 0,
    decompress_ms: 0,
    sandbox_ms: 0,
    restore_ms: 0,
    audit_ms: 0,
    parity_ms: 0,
    teardown_ms: 0,
    total_rto_ms: 0,
  };

  const phases = [
    { name: 'Snapshot Fetch', ms: timings.fetch_ms, color: 'bg-indigo-500' },
    { name: 'Decompress Stream', ms: timings.decompress_ms, color: 'bg-cyan-500' },
    { name: 'Sandbox Boot', ms: timings.sandbox_ms, color: 'bg-blue-500' },
    { name: 'Restore Injection', ms: timings.restore_ms, color: 'bg-emerald-500' },
    { name: 'Integrity Audit', ms: timings.audit_ms, color: 'bg-purple-500' },
    { name: 'Teardown Prune', ms: timings.teardown_ms, color: 'bg-amber-500' },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-4xl max-h-[90vh] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-surface-elevated/50">
          <div className="flex items-center gap-3">
            <HealthScoreBadge grade={drill.health_grade} score={drill.health_score} size="md" />
            <div>
              <h2 className="text-lg font-bold text-slate-100 flex items-center gap-2">
                {drill.name}
              </h2>
              <span className="text-xs font-mono text-slate-400">{drill.id}</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {drill.certificate && (
              <button
                type="button"
                onClick={() => onViewCert(drill)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors"
              >
                <ShieldCheck className="w-4 h-4" />
                View Certificate
              </button>
            )}
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 text-sm">
          {/* Metadata Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 font-mono text-xs">
            <div className="bg-surface-elevated/40 border border-border/50 p-3 rounded-lg">
              <span className="text-slate-400 block text-[10px] uppercase">Engine</span>
              <span className="text-slate-200 font-bold capitalize mt-0.5 block">{drill.engine}</span>
            </div>

            <div className="bg-surface-elevated/40 border border-border/50 p-3 rounded-lg">
              <span className="text-slate-400 block text-[10px] uppercase">Measured RTO</span>
              <span className="text-blue-400 font-bold mt-0.5 block">{drill.timings?.total_rto_ms} ms</span>
            </div>

            <div className="bg-surface-elevated/40 border border-border/50 p-3 rounded-lg">
              <span className="text-slate-400 block text-[10px] uppercase">SLA Target</span>
              <span className={`font-bold mt-0.5 block ${drill.sla_met ? 'text-emerald-400' : 'text-rose-400'}`}>
                {drill.rto_sla_seconds}s ({drill.sla_met ? 'Passed' : 'Exceeded'})
              </span>
            </div>

            <div className="bg-surface-elevated/40 border border-border/50 p-3 rounded-lg">
              <span className="text-slate-400 block text-[10px] uppercase">Snapshot Size</span>
              <span className="text-slate-200 font-bold mt-0.5 block">
                {(drill.snapshot_bytes / 1024).toFixed(1)} KB
              </span>
            </div>
          </div>

          {/* Snapshot Info */}
          <div className="bg-surface-elevated/30 border border-border/40 p-4 rounded-xl font-mono text-xs space-y-1.5">
            <div className="text-slate-400">
              <span className="text-slate-500">Path:</span> {drill.snapshot_path}
            </div>
            <div className="text-slate-400 truncate">
              <span className="text-slate-500">SHA-256:</span> {drill.snapshot_hash || 'Computed on stream'}
            </div>
          </div>

          {/* Phase Timings Timeline */}
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-300 mb-3 flex items-center gap-1.5 font-sans">
              <Clock className="w-4 h-4 text-slate-400" />
              Phase Duration Breakdown
            </h4>

            {/* Combined Bar */}
            <div className="w-full h-3 rounded-full bg-slate-800 flex overflow-hidden mb-3">
              {phases.map((p, i) => {
                const widthPercent = Math.max((p.ms / total) * 100, p.ms > 0 ? 2 : 0);
                return (
                  <div
                    key={i}
                    style={{ width: `${widthPercent}%` }}
                    className={`${p.color} h-full`}
                    title={`${p.name}: ${p.ms}ms`}
                  />
                );
              })}
            </div>

            {/* Phase Pills */}
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-xs font-mono">
              {phases.map((p, i) => (
                <div key={i} className="flex items-center justify-between p-2 rounded bg-surface-elevated/40 border border-border/40">
                  <div className="flex items-center gap-2">
                    <span className={`w-2.5 h-2.5 rounded-full ${p.color}`} />
                    <span className="text-slate-300">{p.name}</span>
                  </div>
                  <span className="font-bold text-slate-200">{p.ms} ms</span>
                </div>
              ))}
            </div>
          </div>

          {/* Relational Constraint Violations */}
          {drill.audit?.constraint_violations && drill.audit.constraint_violations.length > 0 && (
            <div>
              <h4 className="text-xs font-semibold uppercase tracking-wider text-rose-400 mb-2 flex items-center gap-1.5 font-sans">
                <AlertTriangle className="w-4 h-4" />
                Relational Foreign Key Violations ({drill.audit.constraint_violations.length})
              </h4>
              <div className="bg-rose-500/10 border border-rose-500/30 rounded-xl overflow-hidden font-mono text-xs">
                <table className="w-full text-left">
                  <thead className="bg-rose-500/20 text-rose-300">
                    <tr>
                      <th className="py-2 px-3">Table</th>
                      <th className="py-2 px-3">Constraint</th>
                      <th className="py-2 px-3">Details</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-rose-500/20 text-rose-200">
                    {drill.audit.constraint_violations.map((v, i) => (
                      <tr key={i}>
                        <td className="py-2 px-3 font-bold">{v.table}</td>
                        <td className="py-2 px-3">{v.constraint}</td>
                        <td className="py-2 px-3 text-rose-300">{v.details}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Parity Report */}
          {drill.audit?.parity?.records && drill.audit.parity.records.length > 0 && (
            <div>
              <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-300 mb-2 flex items-center gap-1.5 font-sans">
                <Database className="w-4 h-4 text-slate-400" />
                Table Row-Count &amp; Schema Parity
              </h4>
              <div className="bg-surface-elevated/30 border border-border/60 rounded-xl overflow-hidden font-mono text-xs">
                <table className="w-full text-left">
                  <thead className="bg-surface-elevated text-slate-400 text-[11px]">
                    <tr>
                      <th className="py-2 px-3">Table</th>
                      <th className="py-2 px-3">Expected Min Rows</th>
                      <th className="py-2 px-3">Actual Rows</th>
                      <th className="py-2 px-3">Status</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/40">
                    {drill.audit.parity.records.map((r, i) => (
                      <tr key={i}>
                        <td className="py-2 px-3 font-bold text-slate-200">{r.table}</td>
                        <td className="py-2 px-3 text-slate-400">{r.expected_rows}</td>
                        <td className="py-2 px-3 text-slate-200">{r.actual_rows}</td>
                        <td className="py-2 px-3">
                          {r.matched ? (
                            <span className="text-emerald-400 flex items-center gap-1">
                              <CheckCircle2 className="w-3.5 h-3.5" /> Matched
                            </span>
                          ) : (
                            <span className="text-rose-400 font-bold">
                              {r.missing_table ? 'Missing Table' : `Short by ${Math.abs(r.diff)}`}
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Structural Errors */}
          {drill.audit?.structural_errors && drill.audit.structural_errors.length > 0 && (
            <div>
              <h4 className="text-xs font-semibold uppercase tracking-wider text-rose-400 mb-2 flex items-center gap-1.5 font-sans">
                <AlertTriangle className="w-4 h-4" />
                Structural Integrity Diagnostics
              </h4>
              <div className="bg-rose-500/10 border border-rose-500/30 p-3 rounded-xl font-mono text-xs text-rose-300 space-y-1">
                {drill.audit.structural_errors.map((err, i) => (
                  <div key={i}>• {err}</div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
