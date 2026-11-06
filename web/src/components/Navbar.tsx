import React from 'react';
import { SystemStatus } from '../types/api';
import { HealthScoreBadge } from './HealthScoreBadge';
import { Database, Play, Sparkles, RefreshCw, Terminal, Calendar, Clock, BarChart2 } from 'lucide-react';

interface Props {
  status: SystemStatus | null;
  activeSandboxCount: number;
  onOpenRunModal: () => void;
  onOpenSandboxesModal: () => void;
  onOpenSchedulerModal: () => void;
  onOpenBisectModal: () => void;
  onGenerateFixtures: () => void;
  onRefresh: () => void;
  generatingFixtures: boolean;
}

export const Navbar: React.FC<Props> = ({
  status,
  activeSandboxCount,
  onOpenRunModal,
  onOpenSandboxesModal,
  onOpenSchedulerModal,
  onOpenBisectModal,
  onGenerateFixtures,
  onRefresh,
  generatingFixtures,
}) => {
  return (
    <header className="border-b border-border/80 bg-surface/90 backdrop-blur sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center text-white shadow-lg shadow-blue-500/20">
            <Database className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-base tracking-tight text-white font-mono">restorecraft</span>
              <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/30">
                DR Engine
              </span>
            </div>
            <p className="text-[11px] text-slate-400 hidden sm:block">
              Database Disaster Recovery Drill &amp; Parity Engine
            </p>
          </div>
        </div>

        {/* System Status Summary */}
        {status && (
          <div className="hidden xl:flex items-center gap-4 text-xs font-mono">
            <div className="flex items-center gap-2 bg-surface-elevated/40 px-3 py-1.5 rounded-lg border border-border/50">
              <span className="text-slate-400">Latest Grade:</span>
              <HealthScoreBadge grade={status.latest_grade} score={status.latest_score} size="sm" />
            </div>

            <div className="flex items-center gap-2 bg-surface-elevated/40 px-3 py-1.5 rounded-lg border border-border/50">
              <span className="text-slate-400">Avg RTO:</span>
              <span className="text-blue-400 font-bold">{status.average_rto_ms} ms</span>
            </div>

            <div className="flex items-center gap-2 bg-surface-elevated/40 px-3 py-1.5 rounded-lg border border-border/50">
              <span className="text-slate-400">Success Rate:</span>
              <span className="text-emerald-400 font-bold">
                {status.total_drills > 0
                  ? `${Math.round((status.successful_drills / status.total_drills) * 100)}%`
                  : 'N/A'}
              </span>
            </div>
          </div>
        )}

        {/* Feature Tools & Action CTAs */}
        <div className="flex items-center gap-2">
          {/* Active Sandboxes Pill */}
          <button
            type="button"
            onClick={onOpenSandboxesModal}
            className={`hidden md:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-mono border transition-colors ${
              activeSandboxCount > 0
                ? 'bg-emerald-500/10 border-emerald-500/40 text-emerald-300 hover:bg-emerald-500/20'
                : 'bg-surface-elevated/50 border-border/60 text-slate-400 hover:bg-slate-800'
            }`}
            title="Inspect running ephemeral database sandboxes"
          >
            <Terminal className="w-3.5 h-3.5 text-emerald-400" />
            <span>Sandboxes</span>
            {activeSandboxCount > 0 && (
              <span className="px-1.5 py-0.2 rounded-full bg-emerald-500 text-black text-[10px] font-bold">
                {activeSandboxCount}
              </span>
            )}
          </button>

          {/* Time-Travel Bisect */}
          <button
            type="button"
            onClick={onOpenBisectModal}
            className="hidden lg:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-surface-elevated/50 hover:bg-slate-800 text-slate-300 border border-border/60 text-xs font-mono transition-colors"
            title="Time-Travel Snapshot Bisect Engine"
          >
            <Clock className="w-3.5 h-3.5 text-indigo-400" />
            <span>Bisect</span>
          </button>

          {/* Scheduler */}
          <button
            type="button"
            onClick={onOpenSchedulerModal}
            className="hidden lg:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-surface-elevated/50 hover:bg-slate-800 text-slate-300 border border-border/60 text-xs font-mono transition-colors"
            title="Automated Cron Drill Scheduler"
          >
            <Calendar className="w-3.5 h-3.5 text-blue-400" />
            <span>Scheduler</span>
          </button>

          {/* Prometheus Metrics Link */}
          <a
            href="/metrics"
            target="_blank"
            rel="noopener noreferrer"
            className="p-2 rounded-lg bg-surface-elevated hover:bg-slate-800 text-slate-400 hover:text-slate-200 border border-border transition-colors hidden sm:block"
            title="Open Prometheus /metrics"
          >
            <BarChart2 className="w-4 h-4 text-purple-400" />
          </a>

          {/* Refresh */}
          <button
            type="button"
            onClick={onRefresh}
            className="p-2 rounded-lg bg-surface-elevated hover:bg-slate-800 text-slate-300 border border-border transition-colors"
            title="Refresh dashboard metrics"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          {/* Test Fixtures */}
          <button
            type="button"
            onClick={onGenerateFixtures}
            disabled={generatingFixtures}
            className="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-surface-elevated hover:bg-slate-800 text-slate-300 text-xs font-medium border border-border transition-colors disabled:opacity-50"
            title="Generate testbed database snapshots"
          >
            <Sparkles className="w-3.5 h-3.5 text-blue-400" />
            {generatingFixtures ? 'Generating...' : 'Test Fixtures'}
          </button>

          {/* Execute Drill CTA */}
          <button
            type="button"
            onClick={onOpenRunModal}
            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold transition-all shadow-md shadow-blue-500/20 active:scale-95"
          >
            <Play className="w-3.5 h-3.5 fill-current" />
            <span>Execute Drill</span>
          </button>
        </div>
      </div>
    </header>
  );
};
