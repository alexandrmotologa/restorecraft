import React, { useState } from 'react';
import { DisasterRecoveryCertificate } from '../types/api';
import { HealthScoreBadge } from './HealthScoreBadge';
import { X, ShieldCheck, Copy, Check, Lock, CheckCircle2, AlertCircle } from 'lucide-react';

interface Props {
  cert: DisasterRecoveryCertificate | null;
  onClose: () => void;
}

export const CertificateViewer: React.FC<Props> = ({ cert, onClose }) => {
  const [copied, setCopied] = useState(false);
  const [verifying, setVerifying] = useState(false);
  const [verifyResult, setVerifyResult] = useState<{ valid: boolean; reason?: string } | null>(null);

  if (!cert) return null;

  const handleCopyJSON = () => {
    navigator.clipboard.writeText(JSON.stringify(cert, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleVerify = async () => {
    setVerifying(true);
    try {
      const res = await fetch('/api/certificates/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ certificate: cert }),
      });
      const data = await res.json();
      setVerifyResult(data);
    } catch (e: any) {
      setVerifyResult({ valid: false, reason: e.message });
    } finally {
      setVerifying(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 bg-gradient-to-r from-blue-900/30 to-surface-elevated/40 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-sm font-bold uppercase tracking-wider text-slate-100 font-mono">
                Disaster Recovery Health Certificate
              </h3>
              <p className="text-xs text-slate-400">Cryptographically signed verification audit proof</p>
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

        {/* Certificate Body */}
        <div className="p-6 space-y-4 font-mono text-xs">
          {/* Certificate Badge Banner */}
          <div className="p-4 rounded-xl bg-surface-elevated/50 border border-border flex items-center justify-between">
            <div>
              <span className="text-[10px] text-slate-400 uppercase tracking-wider block">Target Database</span>
              <span className="text-sm font-bold text-slate-200 capitalize font-sans">{cert.drill_name}</span>
              <span className="text-[11px] text-slate-500 block mt-0.5">{cert.certificate_id}</span>
            </div>
            <HealthScoreBadge grade={cert.health_grade} score={cert.health_score} size="lg" />
          </div>

          {/* Key Facts */}
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-surface-elevated/30 border border-border/40 p-3 rounded-lg">
              <span className="text-slate-400 text-[10px] block">ENGINE</span>
              <span className="text-slate-200 font-bold capitalize">{cert.database_engine}</span>
            </div>
            <div className="bg-surface-elevated/30 border border-border/40 p-3 rounded-lg">
              <span className="text-slate-400 text-[10px] block">MEASURED RTO</span>
              <span className="text-blue-400 font-bold">{cert.rto_ms} ms</span>
            </div>
            <div className="bg-surface-elevated/30 border border-border/40 p-3 rounded-lg">
              <span className="text-slate-400 text-[10px] block">SLA COMPLIANCE</span>
              <span className={cert.sla_met ? 'text-emerald-400 font-bold' : 'text-rose-400 font-bold'}>
                {cert.sla_met ? 'Within Threshold' : 'Threshold Exceeded'}
              </span>
            </div>
            <div className="bg-surface-elevated/30 border border-border/40 p-3 rounded-lg">
              <span className="text-slate-400 text-[10px] block">RELATIONAL INTEGRITY</span>
              <span className={cert.integrity_ok ? 'text-emerald-400 font-bold' : 'text-rose-400 font-bold'}>
                {cert.integrity_ok ? '100% Passed' : 'Violations Detected'}
              </span>
            </div>
          </div>

          {/* Cryptographic Hashes */}
          <div className="p-3.5 bg-slate-950/70 border border-border/60 rounded-xl space-y-2 text-[11px]">
            <div>
              <span className="text-slate-500 block text-[10px]">SNAPSHOT SHA-256</span>
              <span className="text-slate-300 break-all">{cert.snapshot_hash}</span>
            </div>

            <div>
              <span className="text-slate-500 block text-[10px]">SIGNER AUTHORITY</span>
              <span className="text-blue-400">{cert.signer_identity}</span>
            </div>

            <div>
              <span className="text-slate-500 block text-[10px]">HMAC-SHA256 SIGNATURE</span>
              <span className="text-emerald-400/90 break-all">{cert.signature}</span>
            </div>
          </div>

          {/* Live Verification Status */}
          {verifyResult && (
            <div
              className={`p-3 rounded-xl border flex items-center gap-2 ${
                verifyResult.valid
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'
                  : 'bg-rose-500/10 border-rose-500/30 text-rose-300'
              }`}
            >
              {verifyResult.valid ? (
                <>
                  <CheckCircle2 className="w-4 h-4 text-emerald-400 flex-shrink-0" />
                  <span>Signature Verified: Certificate is authentic and uncompromised.</span>
                </>
              ) : (
                <>
                  <AlertCircle className="w-4 h-4 text-rose-400 flex-shrink-0" />
                  <span>Verification Failed: {verifyResult.reason || 'Invalid signature'}</span>
                </>
              )}
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="px-6 py-3.5 bg-surface-elevated/60 border-t border-border flex items-center justify-between">
          <button
            type="button"
            onClick={handleCopyJSON}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs transition-colors"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
            {copied ? 'Copied JSON' : 'Copy JSON'}
          </button>

          <button
            type="button"
            onClick={handleVerify}
            disabled={verifying}
            className="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
          >
            <Lock className="w-3.5 h-3.5" />
            {verifying ? 'Verifying...' : 'Verify Cryptographic Signature'}
          </button>
        </div>
      </div>
    </div>
  );
};
