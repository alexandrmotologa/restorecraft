import React, { useEffect, useRef, useState } from 'react';
import { LogEntry } from '../types/api';
import { Terminal, Pause, Play, Trash2, Radio } from 'lucide-react';

interface Props {
  logs: LogEntry[];
  isConnected: boolean;
  onClearLogs: () => void;
}

export const LiveLogStream: React.FC<Props> = ({ logs, isConnected, onClearLogs }) => {
  const [autoScroll, setAutoScroll] = useState(true);
  const [filterLevel, setFilterLevel] = useState<string>('ALL');
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const filtered = logs.filter((l) => {
    if (filterLevel === 'ALL') return true;
    return l.level === filterLevel;
  });

  const getLevelBadge = (level: string) => {
    switch (level) {
      case 'STEP':
        return 'text-cyan-400 bg-cyan-400/10 border-cyan-400/30';
      case 'ERROR':
        return 'text-rose-400 bg-rose-400/10 border-rose-400/30';
      case 'WARN':
        return 'text-amber-400 bg-amber-400/10 border-amber-400/30';
      case 'INFO':
      default:
        return 'text-slate-400 bg-slate-400/10 border-slate-400/20';
    }
  };

  return (
    <div className="bg-[#0b101b] border border-border/80 rounded-xl overflow-hidden shadow-2xl flex flex-col h-96">
      {/* Terminal Title Bar */}
      <div className="px-4 py-2.5 bg-surface-elevated/70 border-b border-border/60 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <Terminal className="w-4 h-4 text-blue-400" />
          <span className="text-xs font-mono font-semibold uppercase tracking-wider text-slate-200">
            Live Drill Telemetry Console
          </span>
          <div className="flex items-center gap-1.5 ml-2 font-mono text-[11px]">
            <Radio className={`w-3 h-3 ${isConnected ? 'text-emerald-400 animate-pulse' : 'text-slate-500'}`} />
            <span className={isConnected ? 'text-emerald-400' : 'text-slate-500'}>
              {isConnected ? 'SSE Active' : 'Disconnected'}
            </span>
          </div>
        </div>

        {/* Toolbar Controls */}
        <div className="flex items-center gap-2">
          {/* Level Filter */}
          <select
            value={filterLevel}
            onChange={(e) => setFilterLevel(e.target.value)}
            className="bg-slate-900 border border-slate-700 text-slate-300 text-[11px] rounded px-2 py-0.5 font-mono focus:outline-none focus:border-blue-500"
          >
            <option value="ALL">All Levels</option>
            <option value="STEP">Steps</option>
            <option value="INFO">Info</option>
            <option value="WARN">Warnings</option>
            <option value="ERROR">Errors</option>
          </select>

          {/* Pause / Resume Auto-scroll */}
          <button
            type="button"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`p-1 rounded text-xs transition-colors ${
              autoScroll ? 'text-blue-400 hover:bg-slate-800' : 'text-slate-400 hover:bg-slate-800'
            }`}
            title={autoScroll ? 'Pause auto-scroll' : 'Resume auto-scroll'}
          >
            {autoScroll ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
          </button>

          {/* Clear Logs */}
          <button
            type="button"
            onClick={onClearLogs}
            className="p-1 rounded text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors"
            title="Clear terminal logs"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Terminal Output */}
      <div ref={containerRef} className="flex-1 p-4 font-mono text-xs overflow-y-auto space-y-1.5 select-text">
        {filtered.length === 0 ? (
          <div className="text-slate-600 italic py-8 text-center font-sans">
            Awaiting drill execution events... Trigger a disaster recovery drill to stream telemetry.
          </div>
        ) : (
          filtered.map((log, i) => {
            const time = new Date(log.timestamp).toLocaleTimeString('en-US', {
              hour12: false,
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
            });

            return (
              <div key={i} className="flex items-start gap-2 leading-relaxed hover:bg-white/[0.02] px-1 py-0.5 rounded">
                <span className="text-slate-600 select-none text-[10px] pt-0.5">[{time}]</span>
                <span className={`px-1.5 py-0.2 rounded border text-[10px] font-semibold select-none ${getLevelBadge(log.level)}`}>
                  {log.level}
                </span>
                {log.phase && (
                  <span className="text-slate-500 text-[11px] select-none">
                    [{log.phase}]
                  </span>
                )}
                <span className={`flex-1 break-words ${log.level === 'ERROR' ? 'text-rose-300 font-semibold' : log.level === 'STEP' ? 'text-cyan-200' : 'text-slate-300'}`}>
                  {log.message}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
