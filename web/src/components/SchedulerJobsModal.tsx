import React, { useEffect, useState } from 'react';
import { ScheduledJob } from '../types/api';
import { X, Calendar, Clock, Database, CheckCircle2, RefreshCw, BarChart2, Activity } from 'lucide-react';

interface Props {
  isOpen: boolean;
  onClose: () => void;
}

export const SchedulerJobsModal: React.FC<Props> = ({ isOpen, onClose }) => {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchJobs = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/scheduler/jobs');
      if (res.ok) {
        const data: ScheduledJob[] = await res.json();
        setJobs(data);
      }
    } catch (e) {
      console.error('Failed to fetch scheduler jobs:', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchJobs();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-3xl max-h-[90vh] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-surface-elevated/60">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400">
              <Calendar className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-sm font-bold uppercase tracking-wider text-slate-100 font-mono">
                Automated Drill Scheduler &amp; Cron Engine
              </h2>
              <p className="text-xs text-slate-400">
                Continuous unassisted disaster recovery drills running in background
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={fetchJobs}
              className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
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

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 text-xs font-mono">
          {jobs.length === 0 ? (
            <div className="p-8 text-center rounded-xl border border-dashed border-border bg-surface-elevated/20 text-slate-400">
              No scheduled cron drills defined in <code className="text-blue-300">restorecraft.yaml</code>.
              Add a <code className="text-blue-300">schedule: "@daily"</code> or <code className="text-blue-300">schedule: "0 3 * * *"</code> field to any drill config to activate automated verification.
            </div>
          ) : (
            <div className="space-y-3">
              {jobs.map((job, i) => (
                <div
                  key={i}
                  className="p-4 rounded-xl bg-surface-elevated/40 border border-border/60 space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4 text-blue-400" />
                      <span className="font-bold text-slate-100 text-sm">{job.drill_config.name}</span>
                      <span className="text-[10px] uppercase px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-300 border border-blue-500/30">
                        {job.drill_config.engine}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      {job.running ? (
                        <span className="px-2 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30 text-[10px] flex items-center gap-1 font-bold">
                          <Activity className="w-3 h-3 animate-spin" /> RUNNING NOW
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 text-[10px] flex items-center gap-1">
                          <CheckCircle2 className="w-3 h-3" /> ACTIVE CRON
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 pt-2 border-t border-border/40 text-slate-300">
                    <div>
                      <span className="text-slate-500 block text-[10px] uppercase">Cron Expression</span>
                      <span className="font-bold text-indigo-400 block mt-0.5">{job.schedule}</span>
                    </div>

                    <div>
                      <span className="text-slate-500 block text-[10px] uppercase">Next Scheduled Run</span>
                      <span className="text-slate-200 block mt-0.5 flex items-center gap-1">
                        <Clock className="w-3 h-3 text-slate-400" />
                        {new Date(job.next_run_at).toLocaleString()}
                      </span>
                    </div>

                    <div>
                      <span className="text-slate-500 block text-[10px] uppercase">Last Executed</span>
                      <span className="text-slate-400 block mt-0.5">
                        {job.last_run_at ? new Date(job.last_run_at).toLocaleString() : 'Never (Pending First Trigger)'}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Prometheus Metrics Callout */}
          <div className="p-4 rounded-xl bg-surface-elevated/30 border border-border/60 flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400">
                <BarChart2 className="w-4 h-4" />
              </div>
              <div>
                <div className="font-bold text-slate-200">Prometheus Telemetry Endpoint</div>
                <div className="text-slate-400 text-[11px]">
                  Real-time metrics exposed on <code className="text-purple-300">/metrics</code> (RTO gauges, pass/fail counters, health scores)
                </div>
              </div>
            </div>

            <a
              href="/metrics"
              target="_blank"
              rel="noopener noreferrer"
              className="px-3 py-1.5 rounded-lg bg-surface-elevated hover:bg-slate-800 text-slate-200 border border-border transition-colors text-xs font-semibold"
            >
              Open /metrics
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};
