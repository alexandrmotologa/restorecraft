import React, { useState } from 'react';
import { X, Play, Database, ShieldAlert, Sparkles } from 'lucide-react';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onDrillStarted: () => void;
}

export const RunDrillDialog: React.FC<Props> = ({ isOpen, onClose, onDrillStarted }) => {
  const [engine, setEngine] = useState('sqlite');
  const [preset, setPreset] = useState('valid_ecommerce.sqlite');
  const [customPath, setCustomPath] = useState('');
  const [sla, setSla] = useState(60);
  const [loading, setLoading] = useState(false);

  if (!isOpen) return null;

  const presets = [
    {
      id: 'valid_ecommerce.sqlite',
      name: 'Valid E-commerce SQLite Database',
      desc: 'Clean production snapshot with orders, users, and intact foreign keys.',
      engine: 'sqlite',
      path: './test-backups/valid_ecommerce.sqlite',
      type: 'clean',
    },
    {
      id: 'valid_ecommerce.sql.gz',
      name: 'Valid Gzipped SQL Dump',
      desc: 'Compressed streaming SQL dump verifying decompression & restoration.',
      engine: 'sqlite',
      path: './test-backups/valid_ecommerce.sql.gz',
      type: 'clean',
    },
    {
      id: 'fk_violation.sqlite',
      name: 'Broken Relational Foreign Keys (Chaos)',
      desc: 'Corrupt snapshot with orphan invoice rows testing constraint alerts.',
      engine: 'sqlite',
      path: './test-backups/fk_violation.sqlite',
      type: 'chaos',
    },
    {
      id: 'corrupt_header.sqlite',
      name: 'Storage Bit-Rot Header Corruption (Chaos)',
      desc: 'Damaged page header testing page integrity diagnostics & Grade F alerts.',
      engine: 'sqlite',
      path: './test-backups/corrupt_header.sqlite',
      type: 'chaos',
    },
  ];

  const handleLaunch = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    let snapshotPath = customPath;
    if (!snapshotPath) {
      const selectedPreset = presets.find((p) => p.id === preset);
      snapshotPath = selectedPreset ? selectedPreset.path : './test-backups/valid_ecommerce.sqlite';
    }

    try {
      await fetch('/api/drills/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          engine,
          snapshot_path: snapshotPath,
          storage_type: 'local',
          rto_sla_seconds: Number(sla),
        }),
      });
      onDrillStarted();
      onClose();
    } catch (err) {
      console.error('Launch drill failed:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-surface border border-border w-full max-w-xl rounded-2xl shadow-2xl overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 bg-surface-elevated/60 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400">
              <Play className="w-4 h-4 fill-current" />
            </div>
            <div>
              <h3 className="text-sm font-bold uppercase tracking-wider text-slate-100 font-mono">
                Launch Disaster Recovery Drill
              </h3>
              <p className="text-xs text-slate-400">Spawn an isolated ephemeral sandbox to verify backup</p>
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

        {/* Form Body */}
        <form onSubmit={handleLaunch} className="p-6 space-y-5 text-sm font-sans">
          {/* Engine Selector */}
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 block mb-2 font-mono">
              Database Engine
            </label>
            <div className="grid grid-cols-3 gap-2 font-mono text-xs">
              {['sqlite', 'postgres', 'mysql'].map((eng) => (
                <button
                  key={eng}
                  type="button"
                  onClick={() => setEngine(eng)}
                  className={`p-2.5 rounded-lg border text-center capitalize transition-colors flex items-center justify-center gap-1.5 ${
                    engine === eng
                      ? 'bg-blue-600/20 border-blue-500 text-blue-300 font-bold'
                      : 'bg-surface-elevated/40 border-border/60 text-slate-400 hover:bg-slate-800'
                  }`}
                >
                  <Database className="w-3.5 h-3.5" />
                  {eng}
                </button>
              ))}
            </div>
          </div>

          {/* Preset or Custom Snapshot */}
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 block mb-2 font-mono">
              Snapshot Testbed Preset
            </label>
            <div className="space-y-2">
              {presets.map((p) => (
                <label
                  key={p.id}
                  className={`flex items-start gap-3 p-3 rounded-xl border cursor-pointer transition-colors ${
                    preset === p.id && !customPath
                      ? 'bg-blue-500/10 border-blue-500/40 text-slate-100'
                      : 'bg-surface-elevated/30 border-border/50 text-slate-300 hover:bg-slate-800/40'
                  }`}
                >
                  <input
                    type="radio"
                    name="preset"
                    checked={preset === p.id && !customPath}
                    onChange={() => {
                      setPreset(p.id);
                      setEngine(p.engine);
                      setCustomPath('');
                    }}
                    className="mt-0.5"
                  />
                  <div className="flex-1 text-xs">
                    <div className="font-semibold flex items-center gap-2">
                      {p.name}
                      {p.type === 'chaos' && (
                        <span className="text-[10px] px-1.5 py-0.2 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30 flex items-center gap-1 font-mono">
                          <ShieldAlert className="w-3 h-3" /> Chaos
                        </span>
                      )}
                    </div>
                    <p className="text-slate-400 mt-0.5">{p.desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Custom Path Override */}
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-slate-400 block mb-1 font-mono">
              Or Custom Snapshot Path / Key (Optional)
            </label>
            <input
              type="text"
              placeholder="./backups/my_snapshot.sqlite or S3 Key"
              value={customPath}
              onChange={(e) => setCustomPath(e.target.value)}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          {/* RTO SLA Threshold */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 font-mono">
                RTO SLA Threshold: <span className="text-blue-400">{sla} seconds</span>
              </label>
            </div>
            <input
              type="range"
              min="5"
              max="600"
              step="5"
              value={sla}
              onChange={(e) => setSla(Number(e.target.value))}
              className="w-full accent-blue-500 cursor-pointer"
            />
          </div>

          {/* Footer Actions */}
          <div className="pt-2 flex items-center justify-end gap-3 border-t border-border">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold transition-colors flex items-center gap-2 disabled:opacity-50 shadow-lg shadow-blue-500/20"
            >
              <Sparkles className="w-4 h-4" />
              {loading ? 'Launching...' : 'Execute Recovery Drill'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
