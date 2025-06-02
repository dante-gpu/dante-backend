import { Decimal } from 'decimal.js';

// Authentication & User Management
export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  lastLoginAt?: string;
}

export interface AuthResponse {
  token: string;
  refresh_token: string;
  expires_at: string;
  user_id: string;
  username: string;
  email: string;
  role: string;
  permissions: string[];
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

// Wallet & Billing
export interface Wallet {
  id: string;
  userId: string;
  walletType: WalletType;
  solanaAddress: string;
  balance: string;
  lockedBalance: string;
  pendingBalance: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  lastActivityAt?: string;
}

export type WalletType = 'user' | 'provider';

export interface BalanceResponse {
  walletId: string;
  balance: string;
  lockedBalance: string;
  pendingBalance: string;
  availableBalance: string;
  totalBalance: string;
  lastUpdated: string;
}

export interface Transaction {
  id: string;
  fromWalletId?: string;
  toWalletId?: string;
  type: TransactionType;
  status: TransactionStatus;
  amount: string;
  fee: string;
  description: string;
  solanaSignature?: string;
  sessionId?: string;
  jobId?: string;
  createdAt: string;
  updatedAt: string;
  confirmedAt?: string;
}

export type TransactionType = 'deposit' | 'withdrawal' | 'payment' | 'payout' | 'refund' | 'fee';
export type TransactionStatus = 'pending' | 'processing' | 'confirmed' | 'failed' | 'cancelled';

// GPU & Provider Management
export interface GPUDetail {
  modelName: string;
  vramMb: number;
  driverVersion: string;
  architecture?: string;
  computeCapability?: string;
  cudaCores?: number;
  tensorCores?: number;
  memoryBandwidthGbS?: number;
  powerConsumptionW?: number;
  utilizationGpu?: number;
  utilizationMem?: number;
  temperature?: number;
  powerDraw?: number;
  isHealthy: boolean;
}

export interface Provider {
  id: string;
  ownerId: string;
  name: string;
  hostname?: string;
  ipAddress?: string;
  status: ProviderStatus;
  gpus: GPUDetail[];
  location?: string;
  registeredAt: string;
  lastSeenAt: string;
  metadata?: Record<string, any>;
  reputation?: ProviderReputation;
  pricing?: ProviderPricing;
}

export type ProviderStatus = 'online' | 'offline' | 'busy' | 'maintenance' | 'suspended';

export interface ProviderReputation {
  rating: number;
  totalJobs: number;
  successfulJobs: number;
  avgCompletionTime: string;
  reliability: number;
  communicationRating: number;
}

export interface ProviderPricing {
  baseHourlyRate: string;
  vramRatePerGb: string;
  powerMultiplier: string;
  minimumDuration: number;
  availabilityDiscount?: string;
  bulkDiscount?: string;
}

// Job Management
export interface Job {
  id: string;
  userId: string;
  type: string;
  name: string;
  description?: string;
  priority: number;
  submittedAt: string;
  gpuType?: string;
  gpuCount?: number;
  params: Record<string, any>;
  tags?: string[];
  status: JobStatus;
  providerId?: string;
  sessionId?: string;
  startedAt?: string;
  completedAt?: string;
  error?: string;
  result?: JobResult;
  resourceUsage?: ResourceUsage;
  cost?: JobCost;
}

export type JobStatus = 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';

export interface JobSubmissionRequest {
  type: string;
  name: string;
  description?: string;
  executionType: ExecutionType;
  priority?: number;
  tags?: string[];
  dockerImage?: string;
  dockerCommand?: string[];
  script?: string;
  scriptLanguage?: string;
  environment?: Record<string, string>;
  requirements: ResourceRequirements;
  constraints: JobConstraints;
  inputFiles?: FileSpec[];
  outputFiles?: FileSpec[];
  maxCostDGPU: string;
  maxDurationMinutes: number;
  preferredProviders?: string[];
  excludedProviders?: string[];
  preferredLocation?: string;
  requireGpuAccess: boolean;
  retryCount?: number;
  notificationWebhook?: string;
  customParams?: Record<string, any>;
}

export type ExecutionType = 'docker' | 'script' | 'python' | 'bash';

export interface ResourceRequirements {
  gpuModel?: string;
  gpuMemoryMb: number;
  gpuComputeUnits?: number;
  minGpuMemoryMb?: number;
  cpuCores: number;
  memoryMb: number;
  diskSpaceMb: number;
  networkBandwidthMb?: number;
  architecture?: string;
}

export interface JobConstraints {
  maxCpuUsagePercent: number;
  maxMemoryUsagePercent: number;
  maxGpuUsagePercent: number;
  maxNetworkUsageMb?: number;
  allowNetworkAccess: boolean;
  allowFileSystemAccess: boolean;
  isolationLevel: string;
}

export interface FileSpec {
  url: string;
  path: string;
  size?: number;
  checksum?: string;
  compression?: string;
  headers?: Record<string, string>;
  required: boolean;
}

export interface JobResult {
  success: boolean;
  exitCode: number;
  output?: string;
  error?: string;
  outputFiles?: string[];
  metrics?: Record<string, any>;
  artifacts?: Artifact[];
}

export interface Artifact {
  name: string;
  type: string;
  size: number;
  url: string;
  checksum: string;
  createdAt: string;
}

export interface ResourceUsage {
  cpuPercent: number;
  memoryMb: number;
  memoryPercent: number;
  diskReadMb: number;
  diskWriteMb: number;
  networkTxMb: number;
  networkRxMb: number;
  gpuUtilization?: number;
  gpuMemoryUsage?: number;
  powerConsumption?: number;
  timestamp: string;
}

export interface JobCost {
  estimated: string;
  actual: string;
  breakdown: CostBreakdown;
  currency: 'DGPU';
}

export interface CostBreakdown {
  baseRate: string;
  vramCost: string;
  powerCost: string;
  platformFee: string;
  total: string;
}

// Pricing & Estimation
export interface PricingEstimateRequest {
  gpuModel: string;
  requestedVramGb: number;
  estimatedPowerW: number;
  durationHours: string;
  location?: string;
  providerId?: string;
  cpuCores?: number;
  memoryGb?: number;
  storageGb?: number;
  networkBandwidth?: number;
  priority?: string;
}

export interface PricingEstimateResponse {
  baseHourlyRate: string;
  vramHourlyRate: string;
  powerHourlyRate: string;
  cpuHourlyRate: string;
  memoryHourlyRate: string;
  storageHourlyRate: string;
  networkHourlyRate: string;
  totalHourlyRate: string;
  totalCost: string;
  platformFee: string;
  platformFeePercent: string;
  providerEarnings: string;
  vramPercentage: string;
  calculatedAt: string;
  validUntil: string;
  discountApplied?: string;
  recommendedGpus: string[];
}

// Session & Billing
export interface RentalSession {
  id: string;
  userId: string;
  providerId: string;
  jobId?: string;
  status: SessionStatus;
  gpuModel: string;
  allocatedVram: number;
  totalVram: number;
  vramPercentage: string;
  hourlyRate: string;
  vramRate: string;
  powerRate: string;
  platformFeeRate: string;
  estimatedPowerW: number;
  actualPowerW?: number;
  startedAt: string;
  endedAt?: string;
  lastBilledAt: string;
  totalCost: string;
  platformFee: string;
  providerEarnings: string;
  createdAt: string;
  updatedAt: string;
}

export type SessionStatus = 'active' | 'completed' | 'cancelled' | 'suspended' | 'error';

export interface UsageRecord {
  id: string;
  sessionId: string;
  recordedAt: string;
  gpuUtilization: number;
  vramUtilization: number;
  powerDraw: number;
  temperature: number;
  periodMinutes: number;
  periodCost: string;
  createdAt: string;
}

// Marketplace & Search
export interface MarketplaceFilter {
  location?: string;
  gpuModel?: string;
  minVram?: number;
  maxPricePerHour?: string;
  minRating?: number;
  isOnline?: boolean;
  hasCapacity?: boolean;
  sortBy?: 'price' | 'rating' | 'capacity' | 'location';
  sortOrder?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
}

export interface MarketplaceProvider extends Provider {
  availableGpus: number;
  totalGpus: number;
  currentLoad: number;
  averagePrice: string;
  estimatedWaitTime: string;
  features: string[];
  certifications: string[];
  supportedExecutionTypes: ExecutionType[];
}

// Notifications & Alerts
export interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  title: string;
  message: string;
  data?: Record<string, any>;
  read: boolean;
  createdAt: string;
  expiresAt?: string;
}

export type NotificationType = 'job_completed' | 'job_failed' | 'payment_success' | 'payment_failed' | 'provider_offline' | 'low_balance' | 'security_alert';

// Analytics & Reporting
export interface DashboardMetrics {
  totalJobs: number;
  activeJobs: number;
  completedJobs: number;
  failedJobs: number;
  totalSpent: string;
  averageJobDuration: string;
  averageCostPerJob: string;
  favoriteGpuModel: string;
  reliability: number;
}

export interface ProviderDashboardMetrics {
  totalEarnings: string;
  pendingEarnings: string;
  jobsCompleted: number;
  activeJobs: number;
  uptime: number;
  averageUtilization: number;
  reputation: number;
  totalHours: string;
}

// Error Handling
export interface APIError {
  code: string;
  message: string;
  details?: Record<string, any>;
  timestamp: string;
}

export interface ValidationError {
  field: string;
  message: string;
  value?: any;
}

// WebSocket Events
export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

export interface JobStatusUpdate extends WebSocketMessage {
  type: 'job_status_update';
  data: {
    jobId: string;
    status: JobStatus;
    progress?: number;
    message?: string;
    resourceUsage?: ResourceUsage;
  };
}

export interface UsageUpdate extends WebSocketMessage {
  type: 'usage_update';
  data: {
    sessionId: string;
    usage: ResourceUsage;
    cost: string;
    remainingBalance: string;
  };
}

export interface ProviderStatusUpdate extends WebSocketMessage {
  type: 'provider_status_update';
  data: {
    providerId: string;
    status: ProviderStatus;
    availableGpus: number;
    load: number;
  };
}

// Settings & Preferences
export interface UserSettings {
  notifications: NotificationSettings;
  display: DisplaySettings;
  billing: BillingSettings;
  security: SecuritySettings;
}

export interface NotificationSettings {
  email: boolean;
  browser: boolean;
  jobCompletions: boolean;
  paymentAlerts: boolean;
  providerUpdates: boolean;
  weeklyReports: boolean;
}

export interface DisplaySettings {
  theme: 'light' | 'dark' | 'auto';
  timezone: string;
  currency: string;
  dateFormat: string;
  language: string;
}

export interface BillingSettings {
  autoRefill: boolean;
  autoRefillThreshold: string;
  autoRefillAmount: string;
  defaultMaxCost: string;
  spendingLimits: SpendingLimits;
}

export interface SpendingLimits {
  daily: string;
  weekly: string;
  monthly: string;
}

export interface SecuritySettings {
  twoFactorEnabled: boolean;
  sessionTimeout: number;
  ipWhitelist: string[];
  webhookSecurity: boolean;
}

// Form States
export interface FormState<T> {
  data: T;
  errors: Record<string, string>;
  loading: boolean;
  touched: Record<string, boolean>;
}

// API Response Types
export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: APIError;
  pagination?: PaginationInfo;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  hasNext: boolean;
  hasPrev: boolean;
}

// Component Props
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

// Theme & Styling
export interface Theme {
  colors: {
    primary: string;
    secondary: string;
    accent: string;
    background: string;
    surface: string;
    text: string;
    textSecondary: string;
    border: string;
    success: string;
    warning: string;
    error: string;
    info: string;
  };
  spacing: Record<string, string>;
  typography: Record<string, any>;
  shadows: Record<string, string>;
  borderRadius: Record<string, string>;
} 