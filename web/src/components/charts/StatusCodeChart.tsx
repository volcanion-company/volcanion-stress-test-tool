import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export interface StatusCodeChartProps {
  data: Record<string, number>;
}

const COLORS: Record<string, string> = {
  '2xx': '#10b981', // green
  '3xx': '#3b82f6', // blue
  '4xx': '#f59e0b', // orange
  '5xx': '#ef4444', // red
};

export function StatusCodeChart({ data }: StatusCodeChartProps) {
  // Transform data to chart format
  const chartData = Object.entries(data).map(([code, count]) => {
    const category = code.startsWith('2') ? '2xx' :
                     code.startsWith('3') ? '3xx' :
                     code.startsWith('4') ? '4xx' : '5xx';
    return {
      name: `${code}`,
      value: count,
      category,
    };
  });

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={chartData}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }) => `${name} (${(percent * 100).toFixed(0)}%)`}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {chartData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[entry.category] || '#8884d8'} />
          ))}
        </Pie>
        <Tooltip formatter={(value: number) => value.toLocaleString()} />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
}
