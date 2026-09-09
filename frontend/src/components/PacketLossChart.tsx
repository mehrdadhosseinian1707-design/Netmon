import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import type { Measurement } from '../types';
import { format } from 'date-fns';

interface PacketLossChartProps {
  measurements: Measurement[];
}

export const PacketLossChart: React.FC<PacketLossChartProps> = ({ measurements }) => {
  const data = measurements
    .filter(m => m.packet_loss !== undefined)
    .reverse()
    .map(m => ({
      time: format(new Date(m.time), 'HH:mm:ss'),
      packet_loss: m.packet_loss || 0,
    }));

  if (data.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: '#666' }}>No data available</div>;
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
        <XAxis
          dataKey="time"
          stroke="#666"
          tick={{ fontSize: 12 }}
          angle={-45}
          textAnchor="end"
          height={70}
        />
        <YAxis stroke="#666" tick={{ fontSize: 12 }} label={{ value: '%', angle: -90, position: 'insideLeft' }} />
        <Tooltip
          contentStyle={{
            backgroundColor: '#fff',
            border: '1px solid #e0e0e0',
            borderRadius: '4px',
          }}
        />
        <Legend />
        <Bar dataKey="packet_loss" fill="#ef4444" name="Packet Loss (%)" />
      </BarChart>
    </ResponsiveContainer>
  );
};
