import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export interface VirtualUsersChartProps {
  data: Array<{
    timestamp: string;
    active: number;
    total: number;
  }>;
}

export function VirtualUsersChart({ data }: VirtualUsersChartProps) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis 
          dataKey="timestamp" 
          tickFormatter={(value) => new Date(value).toLocaleTimeString()}
        />
        <YAxis label={{ value: 'Virtual Users', angle: -90, position: 'insideLeft' }} />
        <Tooltip 
          labelFormatter={(value) => new Date(value).toLocaleString()}
          formatter={(value: number) => value.toLocaleString()}
        />
        <Legend />
        <Bar dataKey="active" fill="#8884d8" name="Active VUs" />
        <Bar dataKey="total" fill="#82ca9d" name="Total VUs" />
      </BarChart>
    </ResponsiveContainer>
  );
}
