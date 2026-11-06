import React, { useState } from 'react';
import { BisectReport } from '../types/api';
import { X, Play, Clock, AlertTriangle, CheckCircle2, GitCommit, Search, ShieldCheck } from 'lucide-react';
import { HealthScoreBadge } from './HealthScoreBadge';

interface Props {
  isOpen: boolean;
  onClose: () => void;
}

export const TimeTravelBisectModal: React.FC<Props> = ({ isOpen, onClose }) => {
  const [directory, setDirectory] = useState('./test-backups');
  const [engine, setEngine] = useState('sqlite');
  const [loading, setLoading] = useState(false);
  const [report, setReport] = useState<BisectReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  if (!isOpen) return null;

  const handleRunBisect = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setReport(null);

    try {
      const res = await fetch('/api/drills/bisect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ directory, engine }),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Bisect operation failed');
      }

      const data: BisectReport = await res.json();
      setReport(data);
    } catch (err: any) {
      setError(err.message || 'Failed to execute snapshot bisect');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-3xl max-h-[90vh] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-surface-elevated/60">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
              <Clock className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-sm font-bold uppercase tracking-wider text-slate-100 font-mono">
                Time-Travel Snapshot Bisect Engine
              </h2>
              <p className="text-xs text-slate-400">
                Pinpoint the exact chronological snapshot where corruption or regression was introduced
              </p>
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 text-sm">
          {/* Form */}
          <form onSubmit={handleRunBisect} className="bg-surface-elevated/30 border border-border/50 p-4 rounded-xl space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="sm:col-span-2">
                <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 font-mono block mb-1">
                  Snapshots Directory Path
                </label>
                <input
                  type="text"
                  value={directory}
                  onChange={(e) => setDirectory(e.target.value)}
                  placeholder="./test-backups or /data/backups"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 font-mono block mb-1">
                  Engine
                </label>
                <select
                  value={engine}
                  onChange={(e) => setEngine(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-indigo-500 capitalize"
                >
                  <option value="sqlite">SQLite</option>
                  <option value="postgres">PostgreSQL</option>
                  <option value="mysql">MySQL</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition-all shadow-md shadow-indigo-500/20 disabled:opacity-50"
              >
                {loading ? (
                  <>
                    <Search className="w-4 h-4 animate-spin" />
                    Executing Chronological Bisect...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 fill-current" />
                    Launch Chronological Bisect
                  </>
                )}
              </button>
            </div>
          </form>

          {/* Error */}
          {error && (
            <div className="p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 font-mono text-xs flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Report Results */}
          {report && (
            <div className="space-y-5 animate-in fade-in duration-300">
              {/* Inflection Alert */}
              {report.first_broken_snapshot ? (
                <div className="p-4 rounded-xl bg-rose-500/15 border border-rose-500/40 text-rose-200 font-mono text-xs space-y-2">
                  <div className="flex items-center gap-2 text-rose-400 font-bold uppercase tracking-wider">
                    <AlertTriangle className="w-5 h-5" />
                    <span>Corruption / Failure Point Identified</span>
                  </div>
                  <div>
                    <span className="text-slate-400">First Broken Snapshot:</span>{' '}
                    <strong className="text-white text-sm bg-rose-950/60 px-2 py-0.5 rounded border border-rose-500/40">
                      {report.first_broken_snapshot}
                    </strong>
                  </div>
                  {report.first_failure_reason && (
                    <div className="text-slate-300 text-[11px] pt-1 border-t border-rose-500/30">
                      Failure Reason: {report.first_failure_reason}
                    </div>
                  )}
                </div>
              ) : (
                <div className="p-4 rounded-xl bg-emerald-500/15 border border-emerald-500/40 text-emerald-200 font-mono text-xs flex items-center gap-2">
                  <ShieldCheck className="w-5 h-5 text-emerald-400" />
                  <span>All {report.tested_snapshots} evaluated snapshots are verified clean and structurally sound!</span>
                </div>
              )}

              {/* Metrics Summary */}
              <div className="grid grid-cols-4 gap-3 font-mono text-xs text-center">
                <div className="p-3 bg-surface-elevated/40 border border-border/50 rounded-lg">
                  <span className="text-slate-400 text-[10px] uppercase block">Total Discovered</span>
                  <span className="text-slate-100 font-bold text-base mt-0.5 block">{report.total_snapshots}</span>
                </div>
                <div className="p-3 bg-surface-elevated/40 border border-border/50 rounded-lg">
                  <span className="text-slate-400 text-[10px] uppercase block">Tested Runs</span>
                  <span className="text-blue-400 font-bold text-base mt-0.5 block">{report.tested_snapshots}</span>
                </div>
                <div className="p-3 bg-surface-elevated/40 border border-border/50 rounded-lg">
                  <span className="text-slate-400 text-[10px] uppercase block">Healthy</span>
                  <span className="text-emerald-400 font-bold text-base mt-0.5 block">{report.healthy_count}</span>
                </div>
                <div className="p-3 bg-surface-elevated/40 border border-border/50 rounded-lg">
                  <span className="text-slate-400 text-[10px] uppercase block">Broken</span>
                  <span className="text-rose-400 font-bold text-base mt-0.5 block">{report.broken_count}</span>
                </div>
              </div>

              {/* Timeline Sequence */}
              <div>
                <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 font-mono mb-3 flex items-center gap-2">
                  <GitCommit className="w-4 h-4 text-indigo-400" />
                  Chronological Evaluation Timeline
                </h3>
                <div className="space-y-2">
                  {report.timeline.map((point, i) => {
                    const isFirstBroken = point.snapshot_path === report.first_broken_snapshot;
                    return (
                      <div
                        key={i}
                        className={`p-3 rounded-xl border flex items-center justify-between font-mono text-xs transition-colors ${
                          isFirstBroken
                            ? 'bg-rose-500/20 border-rose-500 text-rose-100 shadow-md shadow-rose-950/50'
                            : point.integrity_ok
                            ? 'bg-surface-elevated/30 border-border/50 text-slate-300'
                            : 'bg-rose-500/10 border-rose-500/30 text-rose-300'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          <div className="flex items-center justify-center w-6 h-6 rounded-full bg-slate-800 text-[11px] font-bold text-slate-400">
                            #{i + 1}
                          </div>
                          <div>
                            <div className="font-semibold flex items-center gap-2">
                              {point.snapshot_path}
                              {isFirstBroken && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-rose-500 text-white font-bold uppercase">
                                  Break Point
                                </span>
                              )}
                            </div>
                            {point.error_message && (
                              <div className="text-[11px] text-rose-300 mt-0.5 truncate max-w-md">
                                {point.error_message}
                              </div>
                            )}
                          </div>
                        </div>

                        <div className="flex items-center gap-4">
                          <div className="text-right">
                            <span className="text-slate-400 text-[10px] block">RTO</span>
                            <span className="text-slate-200 font-bold">{point.rto_ms} ms</span>
                          </div>

                          <HealthScoreBadge grade={point.health_grade} score={point.health_score} size="sm" />

                          {point.integrity_ok ? (
                            <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
                          ) : (
                            <AlertTriangle className="w-5 h-5 text-rose-400 shrink-0" />
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
