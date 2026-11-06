import React from 'react';
import { DrillRun } from '../types/api';
import { HealthScoreBadge } from './HealthScoreBadge';
import { Award, CheckCircle2, XCircle, Clock, Database, ChevronRight } from 'lucide-react';

interface Props {
  drills: DrillRun[];
  onSelectDrill: (drill: DrillRun) => void;
  onViewCert: (drill: DrillRun) => void;
}

export const DrillHistoryTable: React.FC<Props> = ({ drills, onSelectDrill, onViewCert }) => {
  if (drills.length === 0) {
    return (
      <div className="bg-surface border border-border/70 rounded-xl p-12 text-center text-slate-400">
        <Database className="w-10 h-10 mx-auto text-slate-500 mb-3" />
        <h3 className="text-base font-semibold text-slate-200">No Disaster Recovery Drills Executed</h3>
        <p className="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
          Trigger your first recovery drill using the button above to benchmark RTO and test database integrity.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-surface border border-border/70 rounded-xl overflow-hidden shadow-xl">
      <div className="px-6 py-4 border-b border-border/70 flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold uppercase tracking-wider text-slate-200">
            Disaster Recovery Drill History
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Audit logs and recovery certificates issued from isolated sandbox drills
          </p>
        </div>
        <span className="text-xs font-mono text-slate-400 bg-surface-elevated px-2.5 py-1 rounded-md border border-border/40">
          Total Executions: {drills.length}
        </span>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs">
          <thead className="bg-surface-elevated/80 text-slate-400 uppercase font-mono tracking-wider border-b border-border/50 text-[11px] whitespace-nowrap">
            <tr>
              <th className="py-3 px-5">Drill Name &amp; ID</th>
              <th className="py-3 px-4">Engine</th>
              <th className="py-3 px-4">Health Grade</th>
              <th className="py-3 px-4">Measured RTO</th>
              <th className="py-3 px-4">Integrity</th>
              <th className="py-3 px-4">Executed At</th>
              <th className="py-3 px-5 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 font-mono">
            {drills.map((drill) => {
              const formattedDate = new Date(drill.started_at).toLocaleString('en-US', {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
              });

              return (
                <tr
                  key={drill.id}
                  className="hover:bg-slate-800/40 transition-colors group cursor-pointer"
                  onClick={() => onSelectDrill(drill)}
                >
                  <td className="py-3.5 px-5 whitespace-nowrap">
                    <div className="font-sans font-semibold text-slate-200 group-hover:text-blue-400 transition-colors text-sm">
                      {drill.name}
                    </div>
                    <div className="text-[11px] text-slate-500 font-mono mt-0.5">{drill.id}</div>
                  </td>

                  <td className="py-3.5 px-4 whitespace-nowrap">
                    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-slate-800 border border-slate-700 text-slate-300 capitalize text-xs shrink-0">
                      <Database className="w-3 h-3 text-slate-400" />
                      {drill.engine}
                    </span>
                  </td>

                  <td className="py-3.5 px-4 whitespace-nowrap">
                    <HealthScoreBadge grade={drill.health_grade} score={drill.health_score} size="sm" />
                  </td>

                  <td className="py-3.5 px-4 whitespace-nowrap">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span className="font-bold text-slate-200">
                        {drill.timings?.total_rto_ms ?? 0} ms
                      </span>
                    </div>
                    <div className="text-[10px] text-slate-400 mt-0.5">
                      {drill.sla_met ? (
                        <span className="text-emerald-400 font-sans">Under {drill.rto_sla_seconds}s SLA</span>
                      ) : (
                        <span className="text-rose-400 font-sans font-medium">SLA Exceeded</span>
                      )}
                    </div>
                  </td>

                  <td className="py-3.5 px-4 whitespace-nowrap">
                    {drill.audit?.integrity_ok ? (
                      <span className="inline-flex items-center gap-1 text-emerald-400 font-sans text-xs">
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        Verified
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-rose-400 font-sans text-xs font-semibold">
                        <XCircle className="w-3.5 h-3.5" />
                        Violations
                      </span>
                    )}
                  </td>

                  <td className="py-3.5 px-4 text-slate-400 text-xs font-sans whitespace-nowrap">
                    {formattedDate}
                  </td>

                  <td className="py-3.5 px-5 text-right whitespace-nowrap" onClick={(e) => e.stopPropagation()}>
                    <div className="flex items-center justify-end gap-2">
                      {drill.certificate && (
                        <button
                          type="button"
                          onClick={() => onViewCert(drill)}
                          className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 border border-blue-500/30 transition-colors font-sans font-medium whitespace-nowrap shrink-0"
                          title="View signed recovery certificate"
                        >
                          <Award className="w-3.5 h-3.5 shrink-0" />
                          Certificate
                        </button>
                      )}
                      <button
                        type="button"
                        onClick={() => onSelectDrill(drill)}
                        className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded bg-slate-800 text-slate-300 hover:bg-slate-700 border border-slate-700 transition-colors font-sans whitespace-nowrap shrink-0"
                      >
                        Details
                        <ChevronRight className="w-3 h-3 text-slate-400 shrink-0" />
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
