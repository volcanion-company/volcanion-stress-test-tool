# Volcanion Web UI

Modern React dashboard for Volcanion Stress Test Tool.

## Features

- 📊 **Real-time Monitoring** - WebSocket-based live metrics
- 📈 **Interactive Charts** - Response time, throughput, status codes
- 🎨 **Modern UI** - Clean, responsive design with Tailwind CSS
- 🔐 **Authentication** - Secure login with JWT
- 🎯 **Test Management** - Create, run, and monitor tests
- 📱 **Responsive** - Works on desktop and mobile

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **TanStack Query** - Server state management
- **Recharts** - Charts library
- **Tailwind CSS** - Styling
- **Axios** - HTTP client
- **React Router** - Routing

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Backend server running on port 8080

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Open `http://localhost:5173`

### Build for Production

```bash
# Create optimized build
npm run build

# Preview production build
npm run preview
```

Build output in `dist/` directory.

## Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── ui/              # Reusable UI components
│   │   │   ├── Button.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Select.tsx
│   │   │   ├── Table.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── LoadingSpinner.tsx
│   │   │   ├── ErrorMessage.tsx
│   │   │   └── Alert.tsx
│   │   └── charts/          # Chart components
│   │       ├── ResponseTimeChart.tsx
│   │       ├── ThroughputChart.tsx
│   │       ├── StatusCodeChart.tsx
│   │       └── VirtualUsersChart.tsx
│   ├── pages/               # Page components
│   │   ├── Dashboard.tsx
│   │   ├── TestPlans.tsx
│   │   ├── NewTestPlan.tsx
│   │   ├── TestPlanDetail.tsx
│   │   ├── TestRuns.tsx
│   │   ├── TestRunDetail.tsx
│   │   ├── TestRunLive.tsx
│   │   └── Login.tsx
│   ├── contexts/            # React contexts
│   │   └── AuthContext.tsx
│   ├── hooks/               # Custom hooks
│   │   ├── useTestPlans.ts
│   │   ├── useTestRuns.ts
│   │   └── useWebSocket.ts
│   ├── services/            # API services
│   │   └── api.ts
│   ├── types/               # TypeScript types
│   │   └── index.ts
│   ├── utils/               # Utilities
│   │   └── ProtectedRoute.tsx
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── public/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
└── postcss.config.js
```

## Configuration

### API Endpoint

Update in `vite.config.ts`:

```typescript
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### Environment Variables

Create `.env.local`:

```env
VITE_API_URL=http://localhost:8080
```

## Components

### UI Components

**Button**
```tsx
<Button variant="primary" onClick={handleClick}>
  Click Me
</Button>
```

**Card**
```tsx
<Card>
  <CardHeader title="Test Results" />
  <CardContent>Content here</CardContent>
</Card>
```

**Input**
```tsx
<Input
  label="Test Name"
  value={name}
  onChange={(e) => setName(e.target.value)}
/>
```

**Table**
```tsx
<Table
  columns={columns}
  data={data}
  onRowClick={handleRowClick}
/>
```

### Chart Components

**ResponseTimeChart**
```tsx
<ResponseTimeChart data={metrics} />
```

Shows line chart with P50, P95, P99 percentiles.

**ThroughputChart**
```tsx
<ThroughputChart data={metrics} />
```

Area chart for requests per second.

**StatusCodeChart**
```tsx
<StatusCodeChart data={statusCodes} />
```

Pie chart for HTTP status distribution.

## Hooks

### useTestPlans

```tsx
const { data: plans, isLoading } = useTestPlans();
```

Fetches all test plans from API.

### useTestRuns

```tsx
const { data: runs } = useTestRuns();
```

Fetches test run history.

### useWebSocket

```tsx
const { data, isConnected } = useWebSocket(url, {
  onMessage: (data) => console.log(data),
});
```

Real-time WebSocket connection with auto-reconnect.

## Pages

### Dashboard
- Overview statistics
- Recent test runs
- Quick actions

### Test Plans
- List all test plans
- Create new plan
- Run existing plan

### New Test Plan
- Form to create test configuration
- Load pattern selection
- SLA configuration

### Test Run Live
- Real-time metrics
- Live charts
- WebSocket connection status
- Stop test button

### Test Run Detail
- Complete test results
- All metrics and charts
- Export options

## API Integration

All API calls in `src/services/api.ts`:

```typescript
import api from '@/services/api';

// Fetch test plans
const plans = await api.get('/api/test-plans');

// Create test plan
const plan = await api.post('/api/test-plans', data);

// Start test
const run = await api.post('/api/test-runs/start', { test_plan_id });

// Get live metrics
const metrics = await api.get(`/api/test-runs/${id}/live`);
```

## Authentication

Login flow:

1. User enters credentials
2. POST to `/api/auth/login`
3. Store JWT token
4. Add token to requests
5. Protected routes check auth

```tsx
// Login
const { login } = useAuth();
await login(email, password);

// Logout
const { logout } = useAuth();
logout();

// Protected route
<ProtectedRoute>
  <Dashboard />
</ProtectedRoute>
```

## WebSocket Integration

Real-time metrics streaming:

```tsx
const wsUrl = `ws://localhost:8080/api/ws/test-runs/${id}/metrics`;

const { data, isConnected } = useWebSocket(wsUrl, {
  onMessage: (metrics) => {
    // Update UI with live metrics
    setCurrentMetrics(metrics);
  },
});
```

Features:
- Auto-reconnect (up to 5 attempts)
- Fallback to HTTP polling
- Connection status indicator
- Error handling

## Styling

Uses Tailwind CSS utility classes:

```tsx
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
  <h1 className="text-2xl font-bold text-gray-900">Title</h1>
</div>
```

Custom colors in `tailwind.config.js`:

```js
theme: {
  extend: {
    colors: {
      primary: '#3b82f6',
      secondary: '#64748b',
    },
  },
}
```

## Development

### Run Dev Server

```bash
npm run dev
```

Hot reload enabled at `http://localhost:5173`

### Type Checking

```bash
npm run type-check
```

### Linting

```bash
npm run lint
```

### Format Code

```bash
npm run format
```

## Build Optimization

Current build stats:
- Bundle size: ~721 KB (minified)
- Gzip: ~206 KB

Optimization tips:
1. Use dynamic imports for code splitting
2. Lazy load routes
3. Optimize images
4. Enable compression on server

```tsx
// Code splitting
const Dashboard = lazy(() => import('./pages/Dashboard'));
```

## Testing

```bash
# Run tests
npm test

# Coverage
npm run test:coverage
```

## Deployment

### Static Hosting

```bash
npm run build
# Deploy dist/ folder to:
# - Vercel
# - Netlify
# - AWS S3 + CloudFront
# - GitHub Pages
```

### Docker

```dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Troubleshooting

### Port Already in Use

```bash
# Change port in vite.config.ts
server: {
  port: 3000
}
```

### API Connection Failed

- Ensure backend is running on port 8080
- Check CORS settings
- Verify proxy configuration

### WebSocket Not Connecting

- Check browser console
- Verify WebSocket endpoint
- Check firewall settings
- Try HTTP polling fallback

### Build Errors

```bash
# Clear cache
rm -rf node_modules
npm install

# Clear Vite cache
rm -rf .vite
```

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Performance

- First Contentful Paint: < 1s
- Time to Interactive: < 2s
- Lighthouse Score: 90+

## Accessibility

- Semantic HTML
- ARIA labels
- Keyboard navigation
- Screen reader support

## Contributing

1. Follow code style (Prettier + ESLint)
2. Write TypeScript types
3. Add tests for new features
4. Update documentation

## License

MIT
