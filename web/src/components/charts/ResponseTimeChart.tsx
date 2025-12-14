import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export interface ResponseTimeChartProps {
  data: Array<{
    timestamp: string;
    avg: number;
    p50: number;
    p95: number;
    p99: number;
  }>;
}

export function ResponseTimeChart({ data }: ResponseTimeChartProps) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis 
          dataKey="timestamp" 
          tickFormatter={(value) => new Date(value).toLocaleTimeString()}
        />
        <YAxis label={{ value: 'Response Time (ms)', angle: -90, position: 'insideLeft' }} />
        <Tooltip 
          labelFormatter={(value) => new Date(value).toLocaleString()}
          formatter={(value: number) => `${value.toFixed(2)} ms`}
        />
        <Legend />
        <Line type="monotone" dataKey="avg" stroke="#8884d8" name="Average" strokeWidth={2} />
        <Line type="monotone" dataKey="p50" stroke="#82ca9d" name="P50" />
        <Line type="monotone" dataKey="p95" stroke="#ffc658" name="P95" />
        <Line type="monotone" dataKey="p99" stroke="#ff7c7c" name="P99" />
      </LineChart>
    </ResponsiveContainer>
  );
}
