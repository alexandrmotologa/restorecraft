import React, { useEffect, useState, useCallback } from 'react';
import { DrillRun, LogEntry, SystemStatus } from './types/api';
import { Navbar } from './components/Navbar';
import { RtoChart } from './components/RtoChart';
import { DrillHistoryTable } from './components/DrillHistoryTable';
import { LiveLogStream } from './components/LiveLogStream';
import { DrillDetailModal } from './components/DrillDetailModal';
import { CertificateViewer } from './components/CertificateViewer';
import { RunDrillDialog } from './components/RunDrillDialog';
import { HealthScoreBadge } from './components/HealthScoreBadge';
import { ShieldCheck, Activity, AlertOctagon, CheckCircle2 } from 'lucide-react';

export const App: React.FC = () => {
  const [drills, setDrills] = useState<DrillRun[]>([]);
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isSseConnected, setIsSseConnected] = useState(false);
  const [selectedDrill, setSelectedDrill] = useState<DrillRun | null>(null);
  const [certDrill, setCertDrill] = useState<DrillRun | null>(null);
  const [isRunModalOpen, setIsRunModalOpen] = useState(false);
  const [generatingFixtures, setGeneratingFixtures] = useState(false);

  // Fetch drill history and system status
  const fetchData = useCallback(async () => {
    try {
      const [drillsRes, statusRes] = await Promise.all([
        fetch('/api/drills'),
        fetch('/api/status'),
      ]);
      if (drillsRes.ok) {
        const data = await drillsRes.json();
        setDrills(data);
      }
      if (statusRes.ok) {
        const data = await statusRes.json();
        setStatus(data);
      }
    } catch (e) {
      console.error('Failed to fetch data:', e);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 8000);
    return () => clearInterval(interval);
  }, [fetchData]);

  // Server-Sent Events (SSE) telemetry stream
  useEffect(() => {
    let eventSource: EventSource | null = null;

    const connectSSE = () => {
      eventSource = new EventSource('/api/drills/stream');

      eventSource.onopen = () => {
        setIsSseConnected(true);
      };

      eventSource.addEventListener('log', (event: MessageEvent) => {
        try {
          const entry: LogEntry = JSON.parse(event.data);
          setLogs((prev) => [...prev.slice(-300), entry]);
          // Refresh drills when drill step finishes
          if (entry.phase === 'done') {
            fetchData();
          }
        } catch (e) {
          console.error('Failed to parse SSE log:', e);
        }
      });

      eventSource.onerror = () => {
        setIsSseConnected(false);
        eventSource?.close();
        setTimeout(connectSSE, 4000);
      };
    };

    connectSSE();
    return () => {
      eventSource?.close();
    };
  }, [fetchData]);

  const handleGenerateFixtures = async () => {
    setGeneratingFixtures(true);
    try {
      await fetch('/api/fixtures/generate', { method: 'POST' });
      fetchData();
    } catch (e) {
      console.error('Generate fixtures error:', e);
    } finally {
      setGeneratingFixtures(false);
    }
  };

  return (
    <div className="min-h-screen bg-background text-slate-100 flex flex-col font-sans">
      <Navbar
        status={status}
        onOpenRunModal={() => setIsRunModalOpen(true)}
        onGenerateFixtures={handleGenerateFixtures}
        onRefresh={fetchData}
        generatingFixtures={generatingFixtures}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Metric Cards Row */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-surface border border-border/70 p-4 rounded-xl shadow-md">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs uppercase font-mono tracking-wider">Health Grade</span>
              <ShieldCheck className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="flex items-center gap-3">
              <HealthScoreBadge
                grade={status?.latest_grade || 'A+'}
                score={status?.latest_score}
                size="lg"
              />
              <span className="text-xs text-slate-400 font-mono">Formula Score</span>
            </div>
          </div>

          <div className="bg-surface border border-border/70 p-4 rounded-xl shadow-md">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs uppercase font-mono tracking-wider">Average RTO</span>
              <Activity className="w-4 h-4 text-blue-400" />
            </div>
            <div className="text-2xl font-bold font-mono text-blue-400">
              {status?.average_rto_ms ?? 0} <span className="text-xs text-slate-400 font-normal">ms</span>
            </div>
            <p className="text-[11px] text-slate-500 mt-1">Mean Recovery Time Objective</p>
          </div>

          <div className="bg-surface border border-border/70 p-4 rounded-xl shadow-md">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs uppercase font-mono tracking-wider">Verified Runs</span>
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold font-mono text-emerald-400">
              {status?.successful_drills ?? 0}
            </div>
            <p className="text-[11px] text-slate-500 mt-1">Clean restores with 100% parity</p>
          </div>

          <div className="bg-surface border border-border/70 p-4 rounded-xl shadow-md">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs uppercase font-mono tracking-wider">Failures Flagged</span>
              <AlertOctagon className="w-4 h-4 text-rose-400" />
            </div>
            <div className="text-2xl font-bold font-mono text-rose-400">
              {status?.failed_drills ?? 0}
            </div>
            <p className="text-[11px] text-slate-500 mt-1">Bit-rot, truncation, or SLA overruns</p>
          </div>
        </div>

        {/* Charts & Telemetry Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <RtoChart drills={drills} />
          <LiveLogStream
            logs={logs}
            isConnected={isSseConnected}
            onClearLogs={() => setLogs([])}
          />
        </div>

        {/* Drill History Table */}
        <DrillHistoryTable
          drills={drills}
          onSelectDrill={(d) => setSelectedDrill(d)}
          onViewCert={(d) => setCertDrill(d)}
        />
      </main>

      {/* Footer */}
      <footer className="border-t border-border/60 py-6 text-center text-xs text-slate-500 font-mono">
        <p>restorecraft — Automated Disaster Recovery Drill, Backup Verification &amp; Parity Engine</p>
        <p className="text-[11px] text-slate-600 mt-1">Engineered in Go 1.23+ and React 19</p>
      </footer>

      {/* Modals */}
      <DrillDetailModal
        drill={selectedDrill}
        onClose={() => setSelectedDrill(null)}
        onViewCert={(d) => {
          setSelectedDrill(null);
          setCertDrill(d);
        }}
      />

      <CertificateViewer
        cert={certDrill?.certificate || null}
        onClose={() => setCertDrill(null)}
      />

      <RunDrillDialog
        isOpen={isRunModalOpen}
        onClose={() => setIsRunModalOpen(false)}
        onDrillStarted={() => {
          fetchData();
        }}
      />
    </div>
  );
};
