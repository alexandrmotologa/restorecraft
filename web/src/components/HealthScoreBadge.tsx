import React from 'react';
import { HealthGrade } from '../types/api';

interface Props {
  grade: HealthGrade;
  score?: number;
  size?: 'sm' | 'md' | 'lg';
}

export const HealthScoreBadge: React.FC<Props> = ({ grade, score, size = 'md' }) => {
  const getStyles = () => {
    switch (grade) {
      case 'A+':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
      case 'A':
        return 'bg-teal-500/10 text-teal-400 border-teal-500/30';
      case 'B':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
      case 'C':
        return 'bg-orange-500/10 text-orange-400 border-orange-500/30';
      case 'F':
      default:
        return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
    }
  };

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'text-xs px-2 py-0.5';
      case 'lg':
        return 'text-base px-3 py-1 font-bold';
      case 'md':
      default:
        return 'text-sm px-2.5 py-1 font-semibold';
    }
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-md border font-mono tracking-wide ${getStyles()} ${getSizeClasses()}`}
      title={score !== undefined ? `Score: ${score.toFixed(1)} / 100` : undefined}
    >
      <span className="font-extrabold">{grade}</span>
      {score !== undefined && (
        <span className="opacity-80 text-xs font-normal">({score.toFixed(0)}%)</span>
      )}
    </span>
  );
};
