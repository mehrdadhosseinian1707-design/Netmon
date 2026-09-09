import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import type { Measurement } from '../types';
import { format } from 'date-fns';

interface LatencyChartProps {
  measurements: Measurement[];
}

export const LatencyChart: React.FC<LatencyChartProps> = ({ measurements }) => {
  const data = measurements
    .filter(m => m.success && m.latency_ms !== undefined)
    .reverse()
    .map(m => ({
      time: format(new Date(m.time), 'HH:mm:ss'),
      latency: m.latency_ms,
      jitter: m.jitter_ms || 0,
    }));

  if (data.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: '#666' }}>No data available</div>;
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
        <XAxis
          dataKey="time"
          stroke="#666"
          tick={{ fontSize: 12 }}
          angle={-45}
          textAnchor="end"
          height={70}
        />
        <YAxis stroke="#666" tick={{ fontSize: 12 }} label={{ value: 'ms', angle: -90, position: 'insideLeft' }} />
        <Tooltip
          contentStyle={{
            backgroundColor: '#fff',
            border: '1px solid #e0e0e0',
            borderRadius: '4px',
          }}
        />
        <Legend />
        <Line
          type="monotone"
          dataKey="latency"
          stroke="#3b82f6"
          strokeWidth={2}
          dot={{ r: 3 }}
          name="Latency (ms)"
        />
        <Line
          type="monotone"
          dataKey="jitter"
          stroke="#f59e0b"
          strokeWidth={2}
          dot={{ r: 3 }}
          name="Jitter (ms)"
        />
      </LineChart>
    </ResponsiveContainer>
  );
};
