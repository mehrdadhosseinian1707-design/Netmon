import { LineChart, Line, AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

interface ChartData {
  timestamp: string;
  value: number;
}

interface LatencyChartProps {
  data: ChartData[];
  height?: number;
}

export const LatencyChart = ({ data, height = 300 }: LatencyChartProps) => {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data}>
        <defs>
          <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#00d9ff" stopOpacity={0.8}/>
            <stop offset="95%" stopColor="#00d9ff" stopOpacity={0}/>
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
        <XAxis
          dataKey="timestamp"
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
        />
        <YAxis
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
          label={{ value: 'ms', angle: -90, position: 'insideLeft', fill: 'var(--text-muted)' }}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--glass-border)',
            borderRadius: '8px',
            color: 'var(--text-primary)'
          }}
        />
        <Area
          type="monotone"
          dataKey="value"
          stroke="#00d9ff"
          strokeWidth={2}
          fill="url(#latencyGradient)"
          animationDuration={1000}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
};

interface PacketLossChartProps {
  data: ChartData[];
  height?: number;
}

export const PacketLossChart = ({ data, height = 300 }: PacketLossChartProps) => {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <LineChart data={data}>
        <defs>
          <linearGradient id="packetLossGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#ef4444" stopOpacity={0.8}/>
            <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
        <XAxis
          dataKey="timestamp"
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
        />
        <YAxis
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
          label={{ value: '%', angle: -90, position: 'insideLeft', fill: 'var(--text-muted)' }}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--glass-border)',
            borderRadius: '8px',
            color: 'var(--text-primary)'
          }}
        />
        <Line
          type="monotone"
          dataKey="value"
          stroke="#ef4444"
          strokeWidth={2}
          dot={{ fill: '#ef4444', r: 3 }}
          animationDuration={1000}
        />
      </LineChart>
    </ResponsiveContainer>
  );
};

interface BandwidthChartProps {
  data: Array<{ timestamp: string; download: number; upload: number }>;
  height?: number;
}

export const BandwidthChart = ({ data, height = 300 }: BandwidthChartProps) => {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data}>
        <defs>
          <linearGradient id="downloadGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#10b981" stopOpacity={0.8}/>
            <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
          </linearGradient>
          <linearGradient id="uploadGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#a855f7" stopOpacity={0.8}/>
            <stop offset="95%" stopColor="#a855f7" stopOpacity={0}/>
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
        <XAxis
          dataKey="timestamp"
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
        />
        <YAxis
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
          label={{ value: 'Mbps', angle: -90, position: 'insideLeft', fill: 'var(--text-muted)' }}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--glass-border)',
            borderRadius: '8px',
            color: 'var(--text-primary)'
          }}
        />
        <Legend />
        <Area
          type="monotone"
          dataKey="download"
          stroke="#10b981"
          strokeWidth={2}
          fill="url(#downloadGradient)"
          name="Download"
          animationDuration={1000}
        />
        <Area
          type="monotone"
          dataKey="upload"
          stroke="#a855f7"
          strokeWidth={2}
          fill="url(#uploadGradient)"
          name="Upload"
          animationDuration={1000}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
};

interface UptimeChartProps {
  data: Array<{ timestamp: string; uptime: number }>;
  height?: number;
}

export const UptimeChart = ({ data, height = 300 }: UptimeChartProps) => {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
        <XAxis
          dataKey="timestamp"
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
        />
        <YAxis
          stroke="var(--text-muted)"
          style={{ fontSize: '12px' }}
          label={{ value: '%', angle: -90, position: 'insideLeft', fill: 'var(--text-muted)' }}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--glass-border)',
            borderRadius: '8px',
            color: 'var(--text-primary)'
          }}
        />
        <Bar
          dataKey="uptime"
          fill="url(#uptimeGradient)"
          radius={[8, 8, 0, 0]}
          animationDuration={1000}
        />
        <defs>
          <linearGradient id="uptimeGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#10b981" stopOpacity={0.9}/>
            <stop offset="95%" stopColor="#10b981" stopOpacity={0.6}/>
          </linearGradient>
        </defs>
      </BarChart>
    </ResponsiveContainer>
  );
};
