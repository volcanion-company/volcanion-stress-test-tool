import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export interface ThroughputChartProps {
  data: Array<{
    timestamp: string;
    rps: number;
    successRate: number;
  }>;
}

export function ThroughputChart({ data }: ThroughputChartProps) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <AreaChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis 
          dataKey="timestamp" 
          tickFormatter={(value) => new Date(value).toLocaleTimeString()}
        />
        <YAxis label={{ value: 'Requests/sec', angle: -90, position: 'insideLeft' }} />
        <Tooltip 
          labelFormatter={(value) => new Date(value).toLocaleString()}
          formatter={(value: number, name: string) => [
            name === 'rps' ? `${value.toFixed(0)} req/s` : `${value.toFixed(2)}%`,
            name === 'rps' ? 'Throughput' : 'Success Rate'
          ]}
        />
        <Area 
          type="monotone" 
          dataKey="rps" 
          stroke="#8884d8" 
          fill="#8884d8" 
          fillOpacity={0.6}
          name="RPS"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
