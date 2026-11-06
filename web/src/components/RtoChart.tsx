import React, { useState } from 'react';
import { DrillRun } from '../types/api';

interface Props {
  drills: DrillRun[];
}

export const RtoChart: React.FC<Props> = ({ drills }) => {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  // Take the last 15 drills in chronological order (oldest to newest)
  const sorted = [...drills]
    .filter((d) => d.status === 'completed')
    .slice(0, 15)
    .reverse();

  if (sorted.length < 2) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center text-slate-400 bg-surface/50 border border-border/40 rounded-xl h-64">
        <p className="text-sm font-medium">Insufficient drill history to render RTO trendline.</p>
        <p className="text-xs text-slate-500 mt-1">Run at least 2 drills to visualize recovery performance.</p>
      </div>
    );
  }

  const maxRto = Math.max(...sorted.map((d) => d.timings.total_rto_ms), 10);
  const chartHeight = 180;
  const chartWidth = 600;
  const paddingX = 40;
  const paddingY = 25;

  const innerWidth = chartWidth - paddingX * 2;
  const innerHeight = chartHeight - paddingY * 2;

  const points = sorted.map((d, i) => {
    const x = paddingX + (i / (sorted.length - 1)) * innerWidth;
    const y = paddingY + innerHeight - (d.timings.total_rto_ms / maxRto) * innerHeight;
    return { x, y, drill: d };
  });

  const pathD = points.reduce((acc, p, i) => {
    return i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`;
  }, '');

  const areaD = `${pathD} L ${points[points.length - 1].x} ${chartHeight - paddingY} L ${points[0].x} ${chartHeight - paddingY} Z`;

  return (
    <div className="relative bg-surface border border-border/70 rounded-xl p-5 shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-sm font-semibold uppercase tracking-wider text-slate-300">
            Recovery Time Objective (RTO) Trendline
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Measured execution duration in milliseconds across recent disaster recovery drills
          </p>
        </div>
        <div className="flex items-center gap-4 text-xs font-mono">
          <div className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-blue-500"></span>
            <span className="text-slate-300">Measured RTO</span>
          </div>
          <div className="text-slate-400">
            Peak: <span className="text-slate-200 font-bold">{maxRto}ms</span>
          </div>
        </div>
      </div>

      <div className="relative overflow-hidden w-full">
        <svg
          viewBox={`0 0 ${chartWidth} ${chartHeight}`}
          className="w-full h-48 select-none"
          preserveAspectRatio="none"
        >
          <defs>
            <linearGradient id="rtoGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.35" />
              <stop offset="100%" stopColor="#3b82f6" stopOpacity="0.0" />
            </linearGradient>
          </defs>

          {/* Grid lines */}
          <line
            x1={paddingX}
            y1={paddingY}
            x2={chartWidth - paddingX}
            y2={paddingY}
            stroke="#334155"
            strokeDasharray="3 3"
            strokeOpacity="0.5"
          />
          <line
            x1={paddingX}
            y1={paddingY + innerHeight / 2}
            x2={chartWidth - paddingX}
            y2={paddingY + innerHeight / 2}
            stroke="#334155"
            strokeDasharray="3 3"
            strokeOpacity="0.3"
          />
          <line
            x1={paddingX}
            y1={chartHeight - paddingY}
            x2={chartWidth - paddingX}
            y2={chartHeight - paddingY}
            stroke="#334155"
            strokeOpacity="0.8"
          />

          {/* Area fill */}
          <path d={areaD} fill="url(#rtoGradient)" />

          {/* Trendline */}
          <path
            d={pathD}
            fill="none"
            stroke="#3b82f6"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />

          {/* Data points */}
          {points.map((p, i) => (
            <circle
              key={i}
              cx={p.x}
              cy={p.y}
              r={hoveredIdx === i ? 6 : 4}
              className="cursor-pointer transition-all duration-150"
              fill={hoveredIdx === i ? '#60a5fa' : '#3b82f6'}
              stroke="#090d16"
              strokeWidth="2"
              onMouseEnter={() => setHoveredIdx(i)}
              onMouseLeave={() => setHoveredIdx(null)}
            />
          ))}
        </svg>

        {/* Hover Tooltip */}
        {hoveredIdx !== null && points[hoveredIdx] && (
          <div
            className="absolute z-10 pointer-events-none bg-slate-900 border border-slate-700 px-3 py-1.5 rounded-lg shadow-xl text-xs font-mono text-slate-200"
            style={{
              left: `${(points[hoveredIdx].x / chartWidth) * 100}%`,
              top: `${(points[hoveredIdx].y / chartHeight) * 100 - 30}%`,
              transform: 'translate(-50%, -100%)',
            }}
          >
            <div className="font-bold text-blue-400">
              {points[hoveredIdx].drill.timings.total_rto_ms} ms
            </div>
            <div className="text-[10px] text-slate-400">
              Grade {points[hoveredIdx].drill.health_grade} ({points[hoveredIdx].drill.engine})
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
