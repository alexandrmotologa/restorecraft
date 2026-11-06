import React from 'react';
import { DrillRun } from '../types/api';
import { X, Printer, ShieldCheck, CheckCircle2, AlertTriangle, FileText, Lock, Calendar } from 'lucide-react';
import { HealthScoreBadge } from './HealthScoreBadge';

interface Props {
  drill: DrillRun | null;
  onClose: () => void;
}

export const ComplianceReportModal: React.FC<Props> = ({ drill, onClose }) => {
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  if (!drill) return null;

  const cert = drill.certificate;
  const isCompliant = drill.health_score >= 80 && drill.sla_met && drill.audit.integrity_ok;
  const piiAudit = drill.audit?.pii_audit;
  const schemaDiff = drill.audit?.schema_diff;
  const parity = drill.audit?.parity;

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-4xl max-h-[92vh] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header Bar (Hidden during print) */}
        <div className="print:hidden px-6 py-3.5 border-b border-border flex items-center justify-between bg-surface-elevated/70">
          <div className="flex items-center gap-2">
            <FileText className="w-4 h-4 text-blue-400" />
            <h3 className="text-sm font-bold text-slate-100 font-mono uppercase tracking-wider">
              Disaster Recovery Compliance Pack
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handlePrint}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition-colors shadow-sm"
            >
              <Printer className="w-3.5 h-3.5" />
              Print / Export PDF
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

        {/* Printable Document Body */}
        <div className="flex-1 overflow-y-auto p-8 print:p-0 space-y-6 text-slate-200 bg-surface print:bg-white print:text-black">
          {/* Document Header */}
          <div className="border-b border-border/80 print:border-slate-300 pb-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2.5">
                <span className="text-xs uppercase font-mono tracking-widest font-bold text-blue-400 print:text-blue-700">
                  OFFICIAL AUDIT REPORT
                </span>
                <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 print:border-emerald-700 print:text-emerald-700">
                  VERIFIED EVIDENCE
                </span>
              </div>
              <h1 className="text-2xl font-bold tracking-tight text-white print:text-black font-sans mt-1">
                Database Disaster Recovery Verification Report
              </h1>
              <p className="text-xs text-slate-400 print:text-slate-600 mt-0.5 font-mono">
                Regulatory Standards: SOC 2 Type II (CC9.1 / CC9.2) · ISO/IEC 27001:2022 (A.12.3.1) · HIPAA (§ 164.308)
              </p>
            </div>

            <div className="text-right font-mono text-xs">
              <div className="text-slate-400 print:text-slate-600">Report Reference:</div>
              <div className="font-bold text-slate-200 print:text-black">{drill.id}</div>
              <div className="text-slate-500 print:text-slate-600 text-[11px] mt-0.5">
                Timestamp: {new Date(drill.started_at).toUTCString()}
              </div>
            </div>
          </div>

          {/* Compliance Verdict Banner */}
          <div
            className={`p-4 rounded-xl border flex items-center justify-between ${
              isCompliant
                ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-200 print:bg-emerald-50 print:border-emerald-300 print:text-emerald-950'
                : 'bg-rose-500/10 border-rose-500/30 text-rose-200 print:bg-rose-50 print:border-rose-300 print:text-rose-950'
            }`}
          >
            <div className="flex items-center gap-3">
              <div
                className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                  isCompliant
                    ? 'bg-emerald-500/20 text-emerald-400 print:bg-emerald-200 print:text-emerald-800'
                    : 'bg-rose-500/20 text-rose-400 print:bg-rose-200 print:text-rose-800'
                }`}
              >
                {isCompliant ? <ShieldCheck className="w-6 h-6" /> : <AlertTriangle className="w-6 h-6" />}
              </div>
              <div>
                <div className="font-bold text-sm tracking-wide uppercase font-mono">
                  {isCompliant ? 'Compliance Status: AUDIT VERIFIED PASS' : 'Compliance Status: ACTION REQUIRED / DEFICIENT'}
                </div>
                <div className="text-xs opacity-90">
                  {isCompliant
                    ? 'All database restoration milestones, relational integrity checks, and RTO SLA bounds were strictly satisfied.'
                    : 'Disaster recovery run encountered integrity defects, relational orphan keys, or SLA overrun.'}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <HealthScoreBadge grade={drill.health_grade} score={drill.health_score} size="md" />
            </div>
          </div>

          {/* Execution Environment & Evidence Parameters */}
          <div className="space-y-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 print:text-slate-700 font-mono">
              1. Execution &amp; Evidence Parameters
            </h3>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 font-mono text-xs">
              <div className="p-3 rounded-lg bg-surface-elevated/40 print:bg-slate-50 border border-border/50 print:border-slate-200">
                <span className="text-[10px] uppercase text-slate-400 print:text-slate-500 block">Database Engine</span>
                <span className="font-bold text-slate-200 print:text-black capitalize mt-0.5 block">{drill.engine}</span>
              </div>
              <div className="p-3 rounded-lg bg-surface-elevated/40 print:bg-slate-50 border border-border/50 print:border-slate-200">
                <span className="text-[10px] uppercase text-slate-400 print:text-slate-500 block">Measured RTO</span>
                <span className="font-bold text-blue-400 print:text-blue-700 mt-0.5 block">{drill.timings?.total_rto_ms} ms</span>
              </div>
              <div className="p-3 rounded-lg bg-surface-elevated/40 print:bg-slate-50 border border-border/50 print:border-slate-200">
                <span className="text-[10px] uppercase text-slate-400 print:text-slate-500 block">Target SLA</span>
                <span className="font-bold text-slate-200 print:text-black mt-0.5 block">{drill.rto_sla_seconds}s</span>
              </div>
              <div className="p-3 rounded-lg bg-surface-elevated/40 print:bg-slate-50 border border-border/50 print:border-slate-200">
                <span className="text-[10px] uppercase text-slate-400 print:text-slate-500 block">Snapshot Payload</span>
                <span className="font-bold text-slate-200 print:text-black mt-0.5 block">
                  {(drill.snapshot_bytes / 1024).toFixed(1)} KB
                </span>
              </div>
            </div>

            <div className="p-3 rounded-lg bg-surface-elevated/30 print:bg-slate-50 border border-border/40 print:border-slate-200 font-mono text-xs space-y-1">
              <div>
                <span className="text-slate-500">Source Artifact:</span> <span className="text-slate-300 print:text-black">{drill.snapshot_path}</span>
              </div>
              <div className="truncate">
                <span className="text-slate-500">SHA-256 Digest:</span> <span className="text-slate-300 print:text-black">{drill.snapshot_hash || 'Computed during restoration stream'}</span>
              </div>
            </div>
          </div>

          {/* SLA Benchmark Analysis */}
          <div className="space-y-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 print:text-slate-700 font-mono">
              2. Recovery Time Objective (RTO) SLA Benchmark
            </h3>
            <div className="p-4 rounded-xl bg-surface-elevated/30 print:bg-slate-50 border border-border/50 print:border-slate-200 text-xs font-mono space-y-3">
              <div className="flex items-center justify-between">
                <span>Total Recovery Execution: <strong>{drill.timings?.total_rto_ms} ms</strong></span>
                <span>Contractual SLA Target: <strong>{drill.rto_sla_seconds * 1000} ms</strong></span>
                <span className={`font-bold ${drill.sla_met ? 'text-emerald-400 print:text-emerald-700' : 'text-rose-400 print:text-rose-700'}`}>
                  {drill.sla_met ? '✓ SLA TARGET SATISFIED' : '✗ SLA BREACHED'}
                </span>
              </div>

              {/* Breakdown Table */}
              <div className="grid grid-cols-3 sm:grid-cols-6 gap-2 text-[11px] pt-2 border-t border-border/40 print:border-slate-200">
                <div>
                  <div className="text-slate-500">Fetch</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.fetch_ms}ms</div>
                </div>
                <div>
                  <div className="text-slate-500">Decompress</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.decompress_ms}ms</div>
                </div>
                <div>
                  <div className="text-slate-500">Sandbox Boot</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.sandbox_ms}ms</div>
                </div>
                <div>
                  <div className="text-slate-500">Restore Injection</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.restore_ms}ms</div>
                </div>
                <div>
                  <div className="text-slate-500">Integrity Audit</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.audit_ms}ms</div>
                </div>
                <div>
                  <div className="text-slate-500">Teardown</div>
                  <div className="font-bold text-slate-200 print:text-black">{drill.timings?.teardown_ms}ms</div>
                </div>
              </div>
            </div>
          </div>

          {/* Relational Parity & Table Integrity */}
          <div className="space-y-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 print:text-slate-700 font-mono">
              3. Relational Parity &amp; Referential Integrity
            </h3>
            {parity && parity.records && parity.records.length > 0 ? (
              <div className="border border-border/60 print:border-slate-300 rounded-xl overflow-hidden font-mono text-xs">
                <table className="w-full text-left">
                  <thead className="bg-surface-elevated/70 print:bg-slate-100 text-slate-400 print:text-slate-700 text-[11px]">
                    <tr>
                      <th className="py-2 px-3">Table Identifier</th>
                      <th className="py-2 px-3">Expected Min Rows</th>
                      <th className="py-2 px-3">Restored Row Count</th>
                      <th className="py-2 px-3">Parity Status</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/40 print:divide-slate-200">
                    {parity.records.map((r, i) => (
                      <tr key={i}>
                        <td className="py-2 px-3 font-bold text-slate-200 print:text-black">{r.table}</td>
                        <td className="py-2 px-3 text-slate-400 print:text-slate-600">{r.expected_rows}</td>
                        <td className="py-2 px-3 text-slate-200 print:text-black">{r.actual_rows}</td>
                        <td className="py-2 px-3">
                          {r.matched ? (
                            <span className="text-emerald-400 print:text-emerald-700 font-semibold">Matched</span>
                          ) : (
                            <span className="text-rose-400 print:text-rose-700 font-semibold">Discrepant ({r.diff})</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="p-3 rounded-lg bg-surface-elevated/20 print:bg-slate-50 border border-border/40 print:border-slate-200 text-xs text-slate-400 print:text-slate-600 font-mono">
                No explicit baseline row-count assertions declared. Database structural pages passed integrity check.
              </div>
            )}
          </div>

          {/* PII Sanitization Leak Audit Section */}
          <div className="space-y-2">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 print:text-slate-700 font-mono">
              4. Privacy &amp; Data Sanitization Audit (GDPR Art. 32 / HIPAA)
            </h3>
            <div className="p-4 rounded-xl bg-surface-elevated/30 print:bg-slate-50 border border-border/50 print:border-slate-200 text-xs font-mono">
              {piiAudit ? (
                piiAudit.sanitized_ok ? (
                  <div className="flex items-center gap-2 text-emerald-400 print:text-emerald-700">
                    <CheckCircle2 className="w-4 h-4" />
                    <span>No unmasked PII leaks detected. Synthetic or masked snapshot conforms to privacy safeguards.</span>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-rose-400 print:text-rose-700 font-bold">
                      <AlertTriangle className="w-4 h-4" />
                      <span>Privacy Warning: Raw unmasked sensitive data discovered in restored snapshot!</span>
                    </div>
                    <div className="space-y-1 text-slate-300 print:text-black">
                      {piiAudit.findings.map((f, i) => (
                        <div key={i} className="pl-6 text-[11px]">
                          • <strong>{f.table}.{f.column}</strong>: {f.description} ({f.sample_count} samples flagged)
                        </div>
                      ))}
                    </div>
                  </div>
                )
              ) : (
                <div className="text-slate-400 print:text-slate-600">
                  PII scan not requested for this drill run.
                </div>
              )}
            </div>
          </div>

          {/* Schema Drift Audit Section */}
          {schemaDiff && (
            <div className="space-y-2">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-300 print:text-slate-700 font-mono">
                5. Schema Evolution &amp; Drift Detection
              </h3>
              <div className="p-4 rounded-xl bg-surface-elevated/30 print:bg-slate-50 border border-border/50 print:border-slate-200 text-xs font-mono">
                {schemaDiff.matched ? (
                  <div className="flex items-center gap-2 text-emerald-400 print:text-emerald-700">
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Schema definition conforms 100% with gold reference baseline. No missing or modified columns.</span>
                  </div>
                ) : (
                  <div className="text-rose-400 print:text-rose-700 space-y-1">
                    <div>Schema discrepancy detected:</div>
                    {schemaDiff.missing_tables.length > 0 && (
                      <div>Missing Tables: {schemaDiff.missing_tables.join(', ')}</div>
                    )}
                    {schemaDiff.unexpected_tables.length > 0 && (
                      <div>Unexpected Tables: {schemaDiff.unexpected_tables.join(', ')}</div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Cryptographic Seal & Executive Sign-off Block */}
          <div className="pt-6 border-t border-border/80 print:border-slate-300 space-y-6">
            {/* Cryptographic Seal */}
            {cert && (
              <div className="p-4 rounded-xl bg-blue-500/10 print:bg-blue-50 border border-blue-500/30 print:border-blue-200 font-mono text-xs space-y-1">
                <div className="flex items-center gap-2 text-blue-400 print:text-blue-700 font-bold uppercase">
                  <Lock className="w-4 h-4" /> Cryptographic Integrity Signature (HMAC-SHA256)
                </div>
                <div className="text-[11px] text-slate-300 print:text-black break-all">
                  Signer Identity: {cert.signer_identity}
                </div>
                <div className="text-[11px] text-slate-400 print:text-slate-700 break-all">
                  Signature: {cert.signature}
                </div>
              </div>
            )}

            {/* Executive Sign-off Blocks */}
            <div className="grid grid-cols-2 gap-8 pt-4">
              <div className="border-t border-slate-700 print:border-slate-400 pt-3 font-mono text-xs">
                <div className="font-bold text-slate-200 print:text-black">Lead Infrastructure &amp; Database Architect</div>
                <div className="text-slate-400 print:text-slate-600 text-[11px] mt-0.5">Automated Disaster Recovery Validation</div>
                <div className="mt-8 text-[11px] text-slate-500 print:text-slate-600 flex items-center gap-1">
                  <Calendar className="w-3.5 h-3.5" /> Date: ________________________
                </div>
              </div>

              <div className="border-t border-slate-700 print:border-slate-400 pt-3 font-mono text-xs">
                <div className="font-bold text-slate-200 print:text-black">Chief Information Security Officer (CISO)</div>
                <div className="text-slate-400 print:text-slate-600 text-[11px] mt-0.5">Compliance &amp; Business Continuity Sign-off</div>
                <div className="mt-8 text-[11px] text-slate-500 print:text-slate-600 flex items-center gap-1">
                  <Calendar className="w-3.5 h-3.5" /> Date: ________________________
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
