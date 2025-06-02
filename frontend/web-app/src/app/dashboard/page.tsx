'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useRouter } from 'next/navigation';
import { providerService, type Provider, type GPUDetail } from '@/lib/api';
import {
  Zap,
  User,
  Wallet,
  History,
  Settings,
  LogOut,
  CreditCard,
  Activity,
  TrendingUp,
  Clock,
  AlertTriangle,
  CheckCircle,
  Play,
  Pause,
  Plus
} from 'lucide-react';
import Link from 'next/link';

// Updated interfaces to match API
interface WalletData {
  balance: number;
  lockedBalance: number;
  pendingBalance: number;
  availableBalance: number;
  totalBalance: number;
}

interface JobStatus {
  id: string;
  name: string;
  status: 'running' | 'completed' | 'failed' | 'pending';
  gpu: string;
  startTime: string;
  cost: number;
  provider?: string;
  progress?: number;
}

interface DashboardStats {
  totalSpent: number;
  activeJobs: number;
  completedJobs: number;
  totalHours: number;
}

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const [walletData, setWalletData] = useState<WalletData | null>(null);
  const [jobStatuses, setJobStatuses] = useState<JobStatus[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [stats, setStats] = useState<DashboardStats>({
    totalSpent: 0,
    activeJobs: 0,
    completedJobs: 0,
    totalHours: 0
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      // Load data in parallel
      await Promise.allSettled([
        loadWalletData(),
        loadJobStatuses(),
        loadProviders(),
        loadStats()
      ]);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      setError('Failed to load dashboard data');
    } finally {
      setIsLoading(false);
    }
  };

  const loadWalletData = async () => {
    try {
      // For now, use mock data - we'll implement real billing API calls later
      // TODO: Replace with actual billing API call
      await new Promise(resolve => setTimeout(resolve, 500));
      setWalletData({
        balance: 150.75,
        lockedBalance: 25.50,
        pendingBalance: 5.25,
        availableBalance: 125.25,
        totalBalance: 181.50
      });
    } catch (error) {
      console.error('Failed to load wallet data:', error);
    }
  };

  const loadJobStatuses = async () => {
    try {
      // For now, use mock data - we'll implement real job API calls later
      // TODO: Replace with actual gateway API call
      await new Promise(resolve => setTimeout(resolve, 300));
      setJobStatuses([
        {
          id: 'job_1',
          name: 'AI Model Training',
          status: 'running',
          gpu: 'NVIDIA RTX 4090',
          startTime: '2 hours ago',
          cost: 12.50,
          provider: 'CloudGPU Pro',
          progress: 65
        },
        {
          id: 'job_2',
          name: 'Video Rendering',
          status: 'completed',
          gpu: 'NVIDIA A100',
          startTime: '5 hours ago',
          cost: 8.75,
          provider: 'AI Compute Hub'
        },
        {
          id: 'job_3',
          name: 'Data Processing',
          status: 'pending',
          gpu: 'NVIDIA RTX 3080',
          startTime: 'Just now',
          cost: 0,
          provider: 'RenderFarm Elite'
        }
      ]);
    } catch (error) {
      console.error('Failed to load job statuses:', error);
    }
  };

  const loadProviders = async () => {
    try {
      const providersData = await providerService.listProviders();
      setProviders(providersData);
    } catch (error) {
      console.error('Failed to load providers:', error);
      // Use fallback data if API fails
      setProviders([
        {
          id: 'provider_1',
          owner_id: 'owner_1',
          name: 'CloudGPU Pro',
          status: 'online',
          location: 'US-East',
          gpus: [
            {
              model_name: 'NVIDIA RTX 4090',
              vram_mb: 24576,
              driver_version: '536.23',
              is_healthy: true
            }
          ],
          registered_at: new Date().toISOString(),
          last_seen_at: new Date().toISOString()
        }
      ]);
    }
  };

  const loadStats = async () => {
    try {
      // For now, use mock data - we'll implement real stats API calls later
      await new Promise(resolve => setTimeout(resolve, 200));
      setStats({
        totalSpent: 245.80,
        activeJobs: 1,
        completedJobs: 12,
        totalHours: 48.5
      });
    } catch (error) {
      console.error('Failed to load stats:', error);
    }
  };

  const handleLogout = () => {
    logout();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-blue-100 text-blue-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <Play className="w-3 h-3" />;
      case 'completed': return <CheckCircle className="w-3 h-3" />;
      case 'failed': return <AlertTriangle className="w-3 h-3" />;
      case 'pending': return <Clock className="w-3 h-3" />;
      default: return <Clock className="w-3 h-3" />;
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 backdrop-blur-md bg-white/80 border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-red-800 to-red-900 rounded-lg shadow-lg flex items-center justify-center">
                <Image 
                  src="/dantegpu-logo.png" 
                  alt="DanteGPU Logo" 
                  width={20} 
                  height={20}
                  className="w-5 h-5"
                />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">DanteGPU</h1>
                <p className="text-xs text-gray-600">GPU Rental Platform</p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <User className="w-4 h-4 text-gray-600" />
                <span className="text-sm text-gray-700">{user?.username}</span>
                <Badge variant="secondary" className="text-xs">
                  {user?.role}
                </Badge>
              </div>
              <Button 
                variant="outline" 
                onClick={handleLogout}
                className="text-red-600 hover:text-red-700 hover:bg-red-50"
              >
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </nav>

      {/* Dashboard Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700">{error}</p>
            <Button 
              variant="outline" 
              onClick={loadDashboardData}
              className="mt-2"
            >
              Retry
            </Button>
          </div>
        )}

        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            Welcome back, {user?.username}!
          </h2>
          <p className="text-gray-600">
            Manage your GPU rentals and monitor your computing resources.
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Spent</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-800">
                {stats.totalSpent.toFixed(2)} dGPU
              </div>
              <p className="text-xs text-muted-foreground">
                All-time spending
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Jobs</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-blue-600">
                {stats.activeJobs}
              </div>
              <p className="text-xs text-muted-foreground">
                Currently running
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Completed Jobs</CardTitle>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">
                {stats.completedJobs}
              </div>
              <p className="text-xs text-muted-foreground">
                Successfully finished
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Hours</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-purple-600">
                {stats.totalHours}h
              </div>
              <p className="text-xs text-muted-foreground">
                GPU compute time
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Wallet & Quick Actions */}
          <div className="space-y-6">
            {/* Wallet Balance */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Wallet className="w-5 h-5" />
                  <span>Wallet Balance</span>
                </CardTitle>
                <CardDescription>
                  Your dGPU token balance and pending transactions
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {walletData ? (
                  <>
                    <div className="space-y-3">
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">Available:</span>
                        <span className="text-lg font-semibold text-green-600">
                          {walletData.balance.toFixed(2)} dGPU
                        </span>
                      </div>
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">Locked:</span>
                        <span className="text-sm text-orange-600">
                          {walletData.lockedBalance.toFixed(2)} dGPU
                        </span>
                      </div>
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">Pending:</span>
                        <span className="text-sm text-blue-600">
                          {walletData.pendingBalance.toFixed(2)} dGPU
                        </span>
                      </div>
                    </div>
                    
                    <div className="border-t pt-3">
                      <div className="flex justify-between items-center">
                        <span className="font-medium">Total Balance:</span>
                        <span className="text-xl font-bold text-indigo-600">
                          {walletData.totalBalance.toFixed(2)} dGPU
                        </span>
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="animate-pulse space-y-3">
                    <div className="h-4 bg-gray-200 rounded"></div>
                    <div className="h-4 bg-gray-200 rounded"></div>
                    <div className="h-4 bg-gray-200 rounded"></div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>
                  Commonly used actions and shortcuts
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <Button className="w-full justify-start" variant="outline" asChild>
                  <Link href="/providers">
                    <Plus className="w-4 h-4 mr-2" />
                    Browse GPU Market
                  </Link>
                </Button>
                <Button className="w-full justify-start" variant="outline">
                  <CreditCard className="w-4 h-4 mr-2" />
                  Top Up Wallet
                </Button>
                <Button className="w-full justify-start" variant="outline">
                  <History className="w-4 h-4 mr-2" />
                  View Job History
                </Button>
                <Button className="w-full justify-start" variant="outline">
                  <Settings className="w-4 h-4 mr-2" />
                  Account Settings
                </Button>
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Active Jobs */}
          <div className="lg:col-span-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <Activity className="w-5 h-5" />
                    <span>Active Jobs</span>
                  </div>
                  <Button size="sm">
                    <Plus className="w-4 h-4 mr-2" />
                    New Job
                  </Button>
                </CardTitle>
                <CardDescription>
                  Monitor your running and recent GPU jobs
                </CardDescription>
              </CardHeader>
              <CardContent>
                {jobStatuses.length > 0 ? (
                  <div className="space-y-4">
                    {jobStatuses.map((job) => (
                      <div key={job.id} className="border rounded-lg p-4 hover:bg-gray-50 transition-colors">
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center space-x-3">
                            <Badge className={`${getStatusColor(job.status)} flex items-center space-x-1`}>
                              {getStatusIcon(job.status)}
                              <span className="capitalize">{job.status}</span>
                            </Badge>
                            <h3 className="font-medium text-gray-900">{job.name}</h3>
                          </div>
                          <span className="text-sm font-medium text-indigo-600">
                            {job.cost > 0 ? `${job.cost.toFixed(2)} dGPU` : 'Free'}
                          </span>
                        </div>
                        
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm text-gray-600">
                          <div>
                            <span className="font-medium">GPU:</span> {job.gpu}
                          </div>
                          <div>
                            <span className="font-medium">Started:</span> {job.startTime}
                          </div>
                          {job.provider && (
                            <div>
                              <span className="font-medium">Provider:</span> {job.provider}
                            </div>
                          )}
                        </div>

                        {job.progress && job.status === 'running' && (
                          <div className="mt-3">
                            <div className="flex justify-between text-sm text-gray-600 mb-1">
                              <span>Progress</span>
                              <span>{job.progress}%</span>
                            </div>
                            <div className="w-full bg-gray-200 rounded-full h-2">
                              <div 
                                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                                style={{ width: `${job.progress}%` }}
                              ></div>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <Activity className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-gray-900 mb-2">No active jobs</h3>
                    <p className="text-gray-600 mb-4">
                      Start by renting a GPU from our marketplace
                    </p>
                    <Button>
                      <Plus className="w-4 h-4 mr-2" />
                      Browse GPUs
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Available Providers Section */}
        <div className="mt-8">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <TrendingUp className="w-5 h-5" />
                <span>Available Providers</span>
              </CardTitle>
              <CardDescription>
                Browse and connect to GPU providers
              </CardDescription>
            </CardHeader>
            <CardContent>
              {providers.length > 0 ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {providers.slice(0, 6).map((provider) => (
                    <div key={provider.id} className="border rounded-lg p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-center justify-between mb-2">
                        <h3 className="font-medium text-gray-900">{provider.name}</h3>
                        <Badge variant={provider.status === 'online' ? 'default' : 'secondary'}>
                          {provider.status}
                        </Badge>
                      </div>
                      
                      <div className="space-y-1 text-sm text-gray-600">
                        {provider.location && (
                          <div>📍 {provider.location}</div>
                        )}
                        <div>🔧 {provider.gpus.length} GPU(s) available</div>
                        {provider.gpus[0] && (
                          <div>💾 {provider.gpus[0].model_name}</div>
                        )}
                      </div>
                      
                      <Button size="sm" className="w-full mt-3">
                        View Details
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <TrendingUp className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">No providers available</h3>
                  <p className="text-gray-600">Check back later for available GPU providers</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
} 