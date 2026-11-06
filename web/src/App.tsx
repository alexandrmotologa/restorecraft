import React, { useEffect, useState, useCallback } from 'react';
import { DrillRun, LogEntry, SystemStatus, ActiveSandboxInfo } from './types/api';
import { Navbar } from './components/Navbar';
import { RtoChart } from './components/RtoChart';
import { DrillHistoryTable } from './components/DrillHistoryTable';
import { LiveLogStream } from './components/LiveLogStream';
import { DrillDetailModal } from './components/DrillDetailModal';
import { CertificateViewer } from './components/CertificateViewer';
import { ComplianceReportModal } from './components/ComplianceReportModal';
import { ActiveSandboxesModal } from './components/ActiveSandboxesModal';
import { SchedulerJobsModal } from './components/SchedulerJobsModal';
import { TimeTravelBisectModal } from './components/TimeTravelBisectModal';
import { RunDrillDialog } from './components/RunDrillDialog';
import { HealthScoreBadge } from './components/HealthScoreBadge';
import { ShieldCheck, Activity, AlertOctagon, CheckCircle2, Terminal, Clock, Calendar } from 'lucide-react';

export const App: React.FC = () => {
  const [drills, setDrills] = useState<DrillRun[]>([]);
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isSseConnected, setIsSseConnected] = useState(false);
  const [selectedDrill, setSelectedDrill] = useState<DrillRun | null>(null);
  const [certDrill, setCertDrill] = useState<DrillRun | null>(null);
  const [complianceDrill, setComplianceDrill] = useState<DrillRun | null>(null);
  const [isRunModalOpen, setIsRunModalOpen] = useState(false);
  const [isSandboxesModalOpen, setIsSandboxesModalOpen] = useState(false);
  const [isSchedulerModalOpen, setIsSchedulerModalOpen] = useState(false);
  const [isBisectModalOpen, setIsBisectModalOpen] = useState(false);
  const [activeSandboxCount, setActiveSandboxCount] = useState(0);
  const [generatingFixtures, setGeneratingFixtures] = useState(false);

  // Fetch drill history and system status
  const fetchData = useCallback(async () => {
    try {
      const [drillsRes, statusRes, sandboxesRes] = await Promise.all([
        fetch('/api/drills'),
        fetch('/api/status'),
        fetch('/api/sandboxes'),
      ]);
      if (drillsRes.ok) {
        const data = await drillsRes.json();
        setDrills(data);
      }
      if (statusRes.ok) {
        const data = await statusRes.json();
        setStatus(data);
      }
      if (sandboxesRes.ok) {
        const data: ActiveSandboxInfo[] = await sandboxesRes.json();
        setActiveSandboxCount(data.length);
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
        activeSandboxCount={activeSandboxCount}
        onOpenRunModal={() => setIsRunModalOpen(true)}
        onOpenSandboxesModal={() => setIsSandboxesModalOpen(true)}
        onOpenSchedulerModal={() => setIsSchedulerModalOpen(true)}
        onOpenBisectModal={() => setIsBisectModalOpen(true)}
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

        {/* Quick Tools Bar for Mobile / Smaller Screens */}
        <div className="flex md:hidden items-center justify-between gap-2 p-3 bg-surface border border-border/70 rounded-xl text-xs font-mono">
          <button
            type="button"
            onClick={() => setIsSandboxesModalOpen(true)}
            className="flex-1 py-2 px-3 rounded-lg bg-surface-elevated text-emerald-300 flex items-center justify-center gap-1.5"
          >
            <Terminal className="w-3.5 h-3.5" />
            <span>Sandboxes ({activeSandboxCount})</span>
          </button>
          <button
            type="button"
            onClick={() => setIsBisectModalOpen(true)}
            className="flex-1 py-2 px-3 rounded-lg bg-surface-elevated text-indigo-300 flex items-center justify-center gap-1.5"
          >
            <Clock className="w-3.5 h-3.5" />
            <span>Bisect</span>
          </button>
          <button
            type="button"
            onClick={() => setIsSchedulerModalOpen(true)}
            className="flex-1 py-2 px-3 rounded-lg bg-surface-elevated text-blue-300 flex items-center justify-center gap-1.5"
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Scheduler</span>
          </button>
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
        onOpenCompliance={(d) => {
          setSelectedDrill(null);
          setComplianceDrill(d);
        }}
      />

      <CertificateViewer
        cert={certDrill?.certificate || null}
        onClose={() => setCertDrill(null)}
      />

      <ComplianceReportModal
        drill={complianceDrill}
        onClose={() => setComplianceDrill(null)}
      />

      <ActiveSandboxesModal
        isOpen={isSandboxesModalOpen}
        onClose={() => setIsSandboxesModalOpen(false)}
        onSandboxCountChange={(count) => setActiveSandboxCount(count)}
      />

      <SchedulerJobsModal
        isOpen={isSchedulerModalOpen}
        onClose={() => setIsSchedulerModalOpen(false)}
      />

      <TimeTravelBisectModal
        isOpen={isBisectModalOpen}
        onClose={() => setIsBisectModalOpen(false)}
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
