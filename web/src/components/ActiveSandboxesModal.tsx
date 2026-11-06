import React, { useEffect, useState, useCallback } from 'react';
import { ActiveSandboxInfo } from '../types/api';
import { InteractiveSqlConsole } from './InteractiveSqlConsole';
import { X, Terminal, Trash2, Clock, Database, RefreshCw, AlertCircle } from 'lucide-react';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSandboxCountChange?: (count: number) => void;
}

export const ActiveSandboxesModal: React.FC<Props> = ({
  isOpen,
  onClose,
  onSandboxCountChange,
}) => {
  const [sandboxes, setSandboxes] = useState<ActiveSandboxInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeConsoleSandbox, setActiveConsoleSandbox] = useState<ActiveSandboxInfo | null>(null);

  const fetchSandboxes = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/sandboxes');
      if (res.ok) {
        const data: ActiveSandboxInfo[] = await res.json();
        setSandboxes(data);
        onSandboxCountChange?.(data.length);
        if (activeConsoleSandbox && !data.find((s) => s.id === activeConsoleSandbox.id)) {
          setActiveConsoleSandbox(null);
        }
      }
    } catch (e) {
      console.error('Failed to fetch sandboxes:', e);
    } finally {
      setLoading(false);
    }
  }, [activeConsoleSandbox, onSandboxCountChange]);

  useEffect(() => {
    if (isOpen) {
      fetchSandboxes();
      const interval = setInterval(fetchSandboxes, 5000);
      return () => clearInterval(interval);
    }
  }, [isOpen, fetchSandboxes]);

  const handleTerminate = async (id: string) => {
    try {
      await fetch(`/api/sandboxes/${id}`, { method: 'DELETE' });
      fetchSandboxes();
    } catch (e) {
      console.error('Failed to terminate sandbox:', e);
    }
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-4xl max-h-[90vh] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-surface-elevated/60">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Terminal className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-bold uppercase tracking-wider text-slate-100 font-mono">
                  Live Ephemeral Sandboxes
                </h2>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                  {sandboxes.length} Active
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Inspect isolated databases in-flight without affecting production state
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={fetchSandboxes}
              className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
              title="Refresh sandboxes"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {activeConsoleSandbox ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => setActiveConsoleSandbox(null)}
                  className="text-xs text-blue-400 hover:text-blue-300 font-mono flex items-center gap-1"
                >
                  ← Back to Sandboxes List
                </button>
                <span className="text-xs font-mono text-slate-400">
                  Target: <strong className="text-slate-200">{activeConsoleSandbox.id}</strong> ({activeConsoleSandbox.engine})
                </span>
              </div>

              <InteractiveSqlConsole
                sandboxId={activeConsoleSandbox.id}
                engine={activeConsoleSandbox.engine}
                onClose={() => setActiveConsoleSandbox(null)}
              />
            </div>
          ) : sandboxes.length === 0 ? (
            <div className="p-10 text-center rounded-2xl border border-dashed border-border/80 bg-surface-elevated/20 space-y-3 font-mono">
              <AlertCircle className="w-8 h-8 text-slate-500 mx-auto" />
              <div className="text-sm font-bold text-slate-300">No Ephemeral Sandboxes Currently Held</div>
              <p className="text-xs text-slate-400 max-w-md mx-auto">
                By default, recovery sandboxes are torn down immediately after verification to conserve resources.
                To keep a sandbox alive for querying, check <strong>"Hold Sandbox for Debugging"</strong> when launching a drill,
                or set <code className="text-blue-300">hold_minutes: 30</code> in your drill configuration.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {sandboxes.map((sb) => {
                const holdUntil = new Date(sb.hold_until);
                const minutesLeft = Math.max(0, Math.round((holdUntil.getTime() - Date.now()) / 60000));

                return (
                  <div
                    key={sb.id}
                    className="p-4 rounded-xl bg-surface-elevated/40 border border-border/60 hover:border-border transition-colors flex flex-col md:flex-row md:items-center justify-between gap-4 font-mono text-xs"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <Database className="w-4 h-4 text-blue-400" />
                        <span className="font-bold text-slate-100">{sb.id}</span>
                        <span className="text-[10px] uppercase px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-300 border border-blue-500/30">
                          {sb.engine}
                        </span>
                        <span className="text-[10px] uppercase px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-border">
                          {sb.type}
                        </span>
                      </div>

                      <div className="text-slate-400 text-[11px] flex items-center gap-2">
                        <Clock className="w-3.5 h-3.5 text-amber-400" />
                        <span>Held until: {holdUntil.toLocaleTimeString()} ({minutesLeft}m remaining)</span>
                      </div>

                      {sb.file_path && (
                        <div className="text-slate-500 text-[11px] truncate max-w-md">
                          Mount: {sb.file_path}
                        </div>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => setActiveConsoleSandbox(sb)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition-colors shadow-sm"
                      >
                        <Terminal className="w-3.5 h-3.5" />
                        Interactive SQL
                      </button>

                      <button
                        type="button"
                        onClick={() => handleTerminate(sb.id)}
                        className="p-2 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/30 transition-colors"
                        title="Teardown and purge sandbox immediately"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
