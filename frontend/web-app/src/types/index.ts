// Remove Decimal.js dependency and use simple number types
// import Decimal from 'decimal.js';

// Authentication & User Management
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'user' | 'provider' | 'admin';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  profile?: UserProfile;
  wallet?: WalletInfo;
}

export interface UserProfile {
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  bio?: string;
  location?: string;
  website?: string;
  social_links?: SocialLinks;
  preferences?: UserPreferences;
}

export interface SocialLinks {
  twitter?: string;
  github?: string;
  linkedin?: string;
  discord?: string;
}

export interface UserPreferences {
  notifications: NotificationPreferences;
  privacy: PrivacySettings;
  display: DisplaySettings;
}

export interface NotificationPreferences {
  email_notifications: boolean;
  push_notifications: boolean;
  job_updates: boolean;
  billing_alerts: boolean;
  security_alerts: boolean;
  marketing_emails: boolean;
}

export interface PrivacySettings {
  profile_visibility: 'public' | 'private' | 'friends';
  show_activity: boolean;
  show_earnings: boolean;
  allow_direct_messages: boolean;
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
export interface WalletInfo {
  address: string;
  balance: number; // dGPU tokens
  locked_balance: number; // dGPU tokens
  pending_balance: number; // dGPU tokens
  available_balance: number; // dGPU tokens
  
  // Solana-specific
  sol_balance?: number;
  token_accounts?: TokenAccount[];
  
  // Transaction history
  recent_transactions?: Transaction[];
  
  // Settings
  auto_refill_enabled?: boolean;
  auto_refill_threshold?: number;
  auto_refill_amount?: number;
  
  last_updated: string;
}

export interface TokenAccount {
  mint: string;
  balance: number;
  decimals: number;
  symbol: string;
  name: string;
}

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
  tx_hash?: string;
  type: TransactionType;
  amount: number;
  currency: 'dGPU' | 'SOL' | 'USDC';
  
  from_address?: string;
  to_address?: string;
  
  description: string;
  status: TransactionStatus;
  
  // Associated entities
  job_id?: string;
  provider_id?: string;
  user_id?: string;
  
  // Metadata
  gas_fee?: number;
  exchange_rate?: number;
  reference?: string;
  
  created_at: string;
  confirmed_at?: string;
  
  blockchain_data?: Record<string, any>;
}

export type TransactionType = 'deposit' | 'withdrawal' | 'payment' | 'earning' | 'refund' | 'fee' | 'bonus' | 'transfer';
export type TransactionStatus = 'pending' | 'confirming' | 'confirmed' | 'completed' | 'failed' | 'cancelled';

// GPU & Provider Management
export interface GPUDetail {
  id?: string;
  model_name: string;
  vram_mb: number;
  driver_version: string;
  architecture?: string;
  compute_capability?: string;
  cuda_cores?: number;
  tensor_cores?: number;
  memory_bandwidth_gb_s?: number;
  power_consumption_w?: number;
  performance_score?: number;
  
  // Real-time metrics
  utilization_gpu_percent?: number;
  utilization_memory_percent?: number;
  temperature_c?: number;
  power_draw_w?: number;
  fan_speed_percent?: number;
  clock_graphics_mhz?: number;
  clock_memory_mhz?: number;
  
  // Health and status
  is_healthy: boolean;
  last_health_check?: string;
  health_issues?: string[];
}

export interface Provider {
  id: string;
  owner_id: string;
  name: string;
  description?: string;
  hostname?: string;
  ip_address?: string;
  port?: number;
  status: ProviderStatus;
  
  // GPU configuration
  gpus: GPUDetail[];
  total_gpus: number;
  available_gpus: number;
  
  // Location and networking
  location?: string;
  region?: string;
  country?: string;
  continent?: string;
  
  // Pricing and terms
  pricing: ProviderPricing;
  
  // Performance and reliability
  performance_metrics: ProviderMetrics;
  
  // Timestamps
  registered_at: string;
  last_seen_at: string;
  
  // Additional metadata
  metadata?: Record<string, any>;
  tags?: string[];
  certifications?: string[];
}

export type ProviderStatus = 'online' | 'offline' | 'busy' | 'maintenance' | 'suspended';

export interface ProviderPricing {
  base_rate: number; // dGPU tokens per hour
  currency: 'dGPU';
  billing_increment_minutes: number;
  minimum_rental_minutes: number;
  maximum_rental_hours: number;
  setup_fee?: number;
  data_transfer_fee_gb?: number;
  storage_fee_gb_hour?: number;
}

export interface ProviderMetrics {
  uptime_percentage: number;
  average_response_time_ms: number;
  total_jobs_completed: number;
  total_compute_hours: number;
  customer_rating: number;
  total_ratings: number;
  success_rate: number;
  last_updated: string;
}

// Job Management
export interface Job {
  id: string;
  user_id: string;
  provider_id: string;
  
  // Basic job information
  name: string;
  description?: string;
  type: ExecutionType;
  status: JobStatus;
  priority: JobPriority;
  
  // Execution details
  execution: JobExecution;
  
  // Resource requirements
  requirements: ResourceRequirements;
  constraints: JobConstraints;
  
  // File management
  input_files?: FileSpec[];
  output_files?: FileSpec[];
  
  // Billing and cost
  cost_estimate: number; // dGPU tokens
  cost_actual?: number; // dGPU tokens
  billing_details?: BillingDetails;
  
  // Timing
  created_at: string;
  queued_at?: string;
  started_at?: string;
  completed_at?: string;
  expires_at?: string;
  
  // Progress and monitoring
  progress?: JobProgress;
  metrics?: JobMetrics;
  logs?: JobLog[];
  
  // Notifications and webhooks
  notifications?: NotificationSettings;
  
  // Additional metadata
  metadata?: Record<string, any>;
  tags?: string[];
}

export type ExecutionType = 'docker' | 'script' | 'notebook' | 'custom';
export type JobStatus = 'pending' | 'queued' | 'starting' | 'running' | 'pausing' | 'paused' | 'stopping' | 'completed' | 'failed' | 'cancelled' | 'expired';
export type JobPriority = 'low' | 'normal' | 'high' | 'urgent';

export interface JobExecution {
  type: ExecutionType;
  docker_image?: string;
  docker_command?: string[];
  docker_entrypoint?: string[];
  environment_variables?: Record<string, string>;
  working_directory?: string;
  
  script_content?: string;
  script_language?: 'python' | 'bash' | 'javascript' | 'julia' | 'r';
  
  notebook_content?: string;
  notebook_kernel?: string;
  
  custom_runtime?: string;
  custom_config?: Record<string, any>;
}

export interface ResourceRequirements {
  gpu_count: number;
  gpu_memory_gb: number;
  gpu_models?: string[];
  
  cpu_cores?: number;
  ram_gb?: number;
  storage_gb?: number;
  network_bandwidth_mbps?: number;
  
  requires_internet: boolean;
  requires_custom_drivers: boolean;
  custom_requirements?: string[];
}

export interface JobConstraints {
  max_cost_dgpu: number;
  max_duration_hours: number;
  
  preferred_providers?: string[];
  excluded_providers?: string[];
  
  preferred_regions?: string[];
  excluded_regions?: string[];
  
  preferred_gpu_models?: string[];
  min_gpu_memory_gb?: number;
  min_provider_rating?: number;
  
  require_ssd_storage?: boolean;
  require_high_bandwidth?: boolean;
  
  custom_constraints?: Record<string, any>;
}

export interface FileSpec {
  name: string;
  path: string;
  size_bytes?: number;
  checksum?: string;
  type?: 'input' | 'output' | 'checkpoint' | 'log';
  required?: boolean;
}

export interface BillingDetails {
  hourly_rate: number;
  setup_fee?: number;
  data_transfer_fee?: number;
  storage_fee?: number;
  total_compute_hours?: number;
  billing_breakdown?: BillingLineItem[];
}

export interface BillingLineItem {
  description: string;
  quantity: number;
  unit_price: number;
  total_amount: number;
  type: 'compute' | 'storage' | 'transfer' | 'setup' | 'other';
}

export interface JobProgress {
  percentage: number;
  stage: string;
  stage_details?: string;
  estimated_completion?: string;
  stages_completed?: string[];
  stages_remaining?: string[];
}

export interface JobMetrics {
  gpu_utilization: TimeSeries;
  gpu_memory_usage: TimeSeries;
  cpu_utilization: TimeSeries;
  ram_usage: TimeSeries;
  network_io: TimeSeries;
  disk_io: TimeSeries;
  power_consumption: TimeSeries;
  temperature: TimeSeries;
}

export interface TimeSeries {
  timestamps: string[];
  values: number[];
  unit: string;
  aggregation?: 'average' | 'sum' | 'max' | 'min';
}

export interface JobLog {
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  source: 'system' | 'user' | 'gpu' | 'network';
  message: string;
  metadata?: Record<string, any>;
}

export interface NotificationSettings {
  webhook_url?: string;
  webhook_events?: NotificationEvent[];
  email_notifications?: boolean;
  email_events?: NotificationEvent[];
  slack_webhook?: string;
  discord_webhook?: string;
}

export type NotificationEvent = 'job_started' | 'job_completed' | 'job_failed' | 'job_cancelled' | 'progress_update' | 'cost_threshold' | 'duration_threshold';

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

// Marketplace & Search
export interface MarketplaceFilter {
  gpu_models?: string[];
  min_vram_gb?: number;
  max_price_hour?: number;
  min_rating?: number;
  regions?: string[];
  availability?: 'available' | 'busy' | 'all';
  features?: string[];
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

export interface MarketplaceStats {
  total_providers: number;
  total_gpus: number;
  available_gpus: number;
  average_price: number;
  top_gpu_models: Array<{
    model: string;
    count: number;
    avg_price: number;
  }>;
  regions: Array<{
    region: string;
    provider_count: number;
    gpu_count: number;
  }>;
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
}

export interface ValidationError {
  field: string;
  message: string;
  value?: any;
}

// WebSocket Events
export interface WSMessage {
  type: string;
  payload: any;
  timestamp: string;
  id?: string;
}

export interface WSJobUpdate extends WSMessage {
  type: 'job_update';
  payload: {
    job_id: string;
    status: JobStatus;
    progress?: JobProgress;
    metrics?: Partial<JobMetrics>;
  };
}

export interface WSProviderUpdate extends WSMessage {
  type: 'provider_update';
  payload: {
    provider_id: string;
    status: ProviderStatus;
    gpu_metrics?: GPUDetail[];
  };
}

export interface WSWalletUpdate extends WSMessage {
  type: 'wallet_update';
  payload: {
    user_id: string;
    balance: number;
    transaction?: Transaction;
  };
}

// Settings & Preferences
export interface UserSettings {
  notifications: NotificationSettings;
  display: DisplaySettings;
  billing: BillingSettings;
  security: SecuritySettings;
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
  daily_limit?: number;
  weekly_limit?: number;
  monthly_limit?: number;
  per_job_limit?: number;
  enabled: boolean;
}

export interface SecuritySettings {
  two_factor_enabled: boolean;
  login_notifications: boolean;
  api_keys: APIKey[];
  trusted_devices: TrustedDevice[];
  session_timeout_minutes: number;
}

export interface APIKey {
  id: string;
  name: string;
  key_preview: string;
  permissions: string[];
  created_at: string;
  last_used?: string;
  expires_at?: string;
  is_active: boolean;
}

export interface TrustedDevice {
  id: string;
  name: string;
  device_type: 'desktop' | 'mobile' | 'tablet';
  browser?: string;
  os?: string;
  ip_address: string;
  location?: string;
  first_seen: string;
  last_seen: string;
  is_current: boolean;
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
  metadata?: APIMetadata;
}

export interface APIMetadata {
  page?: number;
  limit?: number;
  total?: number;
  has_next?: boolean;
  has_prev?: boolean;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  pages: number;
  has_next: boolean;
  has_prev: boolean;
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