# TODO - Volcanion Stress Test Tool

## Overview

This document tracks remaining improvements and future enhancements for the Volcanion Stress Test Tool. All critical issues have been resolved - these are low-priority (P3) items for future development.

**Last Updated:** January 2025

---

## Completed ✅

All 15 phases of improvements have been successfully implemented:

- [x] Phase 1-8: Initial development and core features
- [x] Phase 9: Security Hardening (10 issues fixed)
- [x] Phase 10: Reliability & Performance (4 issues fixed)
- [x] Phase 11: Testing (unit tests + benchmarks)
- [x] Phase 12: Frontend Improvements (5 issues fixed)
- [x] Phase 13: API & Documentation (4 issues fixed)
- [x] Phase 14: Observability (4 issues fixed)
- [x] Phase 15: DevOps & Deployment (6 issues fixed)
- [x] Additional: Documentation (LICENSE, README, etc.)
- [x] Additional: Frontend Docker support
- [x] Additional: Bug fixes (Go version, Prometheus metrics)

---

## Remaining Tasks (P3 - Low Priority)

### 🔒 Security Enhancements

- [ ] **S1: HttpOnly Cookies for Tokens**
  - Location: `web/src/contexts/AuthContext.tsx`
  - Current: JWT stored in localStorage
  - Improvement: Use httpOnly cookies to prevent XSS token theft
  - Priority: Low (current implementation is common practice)

### 🧪 Testing

- [ ] **T1: Frontend Unit Tests**
  - Location: `web/src/`
  - Add: Vitest or Jest tests for critical components
  - Components to test:
    - `AuthContext` - authentication flow
    - `Dashboard` - statistics calculations
    - `TestPlanForm` - form validation
    - `TestRunLive` - WebSocket connection
  - Priority: Medium

- [ ] **T2: CLI Command Tests**
  - Location: `cmd/volcanion/cmd/`
  - Add: Cobra command tests
  - Commands to test:
    - `run` - test execution
    - `config` - configuration management
  - Priority: Low

- [ ] **T3: End-to-End Tests**
  - Add: Playwright or Cypress tests
  - Flows to test:
    - Login → Create Plan → Run Test → View Results
    - API key authentication
    - WebSocket real-time updates
  - Priority: Low

- [ ] **T4: Load Testing on Self**
  - Add: Performance benchmarks with realistic loads
  - Metrics to capture:
    - Max concurrent connections
    - Memory usage under load
    - Response time percentiles
  - Priority: Low

### 📊 Database

- [ ] **D1: Add Database Indexes**
  - Location: `migrations/*.sql`
  - Add indexes for:
    - `test_runs.completed_at` (for filtering by date)
    - `test_runs.status` (for status queries)
    - `test_plans.created_by` (for user queries)
  - Priority: Low (only needed at scale)

- [ ] **D2: Database Migrations Versioning**
  - Add: golang-migrate or similar tool
  - Current: Manual SQL migrations
  - Priority: Low

### 🚀 Features

- [ ] **F1: Token Refresh Endpoint**
  - Add: `POST /api/v1/auth/refresh`
  - Implement: JWT refresh token flow
  - Priority: Medium

- [ ] **F2: User Profile Management**
  - Add: Profile update endpoint
  - Features:
    - Change password
    - Update display name
    - Email preferences
  - Priority: Low

- [ ] **F3: Test Results Export**
  - Add: Export to CSV/Excel
  - Formats:
    - Detailed request logs
    - Summary statistics
    - Performance metrics
  - Priority: Low

- [ ] **F4: Custom Load Patterns**
  - Add: Plugin system for load patterns
  - Patterns:
    - Spike testing
    - Soak testing
    - Step-up/step-down
  - Priority: Low

- [ ] **F5: Scheduled Tests**
  - Add: Cron-based test scheduling
  - Features:
    - Recurring tests
    - Time-based triggers
    - Slack/Email notifications
  - Priority: Low

### 🏗️ Infrastructure

- [ ] **I1: Kubernetes Helm Chart**
  - Add: Helm chart for K8s deployment
  - Include:
    - Deployment, Service, Ingress
    - ConfigMap, Secret templates
    - HPA for auto-scaling
  - Priority: Low

- [ ] **I2: Multi-Region Workers**
  - Add: Distributed worker architecture
  - Features:
    - Workers in multiple regions
    - Aggregated results
    - Geographic load distribution
  - Priority: Low

- [ ] **I3: GraphQL API**
  - Add: Alternative GraphQL endpoint
  - Benefits:
    - Flexible queries
    - Reduced over-fetching
    - Real-time subscriptions
  - Priority: Low

### 📖 Documentation

- [ ] **DOC1: Video Tutorials**
  - Create: Getting started video
  - Topics:
    - Installation walkthrough
    - Creating first test
    - Reading results
  - Priority: Low

- [ ] **DOC2: Advanced Configuration Guide**
  - Add: docs/ADVANCED_CONFIG.md
  - Topics:
    - Custom headers/auth
    - Load patterns
    - Distributed testing
  - Priority: Low

---

## Priority Legend

| Priority | Description |
|----------|-------------|
| High | Critical for production use |
| Medium | Important for better UX/DX |
| Low | Nice to have, future enhancement |

---

## How to Contribute

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:
- Setting up development environment
- Code style and conventions
- Submitting pull requests
- Testing requirements

---

## Notes

- All P1 (Critical) and P2 (High) issues have been resolved
- Current codebase is production-ready
- These P3 items are for future roadmap consideration
- Prioritize based on user feedback and usage patterns
