'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  Cpu, 
  DollarSign, 
  Clock, 
  Activity,
  Zap,
  Server,
  Shield,
  TrendingUp,
  Settings,
  Bell,
  LogOut,
  Play,
  Pause,
  Square,
  Monitor,
  HardDrive,
  Thermometer,
  BarChart3,
  Eye,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  XCircle,
  MapPin,
  Wallet,
  CreditCard
} from 'lucide-react';
import { authService, providerService } from '@/lib/api';

interface DashboardStats {
  totalProviders: number;
  availableGPUs: number;
  activeJobs: number;
  totalEarnings: string;
  walletBalance: string;
}

interface GPUMetrics {
  utilization: number;
  temperature: number;
  powerDraw: number;
  memoryUsage: number;
  vramTotal: number;
  vramUsed: number;
}

interface ActiveJob {
  id: string;
  name: string;
  provider: string;
  status: 'running' | 'pending' | 'completed' | 'failed';
  startTime: string;
  duration: string;
  cost: string;
  gpuModel: string;
}

interface RecentTransaction {
  id: string;
  type: 'payment' | 'earning' | 'refund';
  amount: string;
  description: string;
  timestamp: string;
  status: 'completed' | 'pending' | 'failed';
}

// Simple Progress component implementation
interface ProgressProps {
  value: number;
  max?: number;
  className?: string;
}

const Progress: React.FC<ProgressProps> = ({ value, max = 100, className = "" }) => {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

  return (
    <div className={`relative h-2 w-full overflow-hidden rounded-full bg-secondary border-2 border-black ${className}`}>
      <div
        className="h-full bg-primary transition-all duration-300 ease-in-out"
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
};

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats>({
    totalProviders: 0,
    availableGPUs: 0,
    activeJobs: 0,
    totalEarnings: '0.00',
    walletBalance: '0.00'
  });
  
  const [providers, setProviders] = useState<any[]>([]);
  const [activeJobs, setActiveJobs] = useState<ActiveJob[]>([]);
  const [transactions, setTransactions] = useState<RecentTransaction[]>([]);
  const [gpuMetrics, setGpuMetrics] = useState<GPUMetrics>({
    utilization: 0,
    temperature: 0,
    powerDraw: 0,
    memoryUsage: 0,
    vramTotal: 0,
    vramUsed: 0
  });

  useEffect(() => {
    loadDashboardData();
    
    // Set up real-time updates
    const interval = setInterval(loadDashboardData, 30000); // Update every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      await Promise.all([
        loadProviders(),
        loadActiveJobs(),
        loadTransactions(),
        loadStats(),
        loadGPUMetrics()
      ]);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadProviders = async () => {
    try {
      const providersData = await providerService.listProviders();
      setProviders(providersData || []);
      
      // Update stats based on providers
      const totalGPUs = providersData?.reduce((total, provider) => total + (provider.gpus?.length || 0), 0) || 0;
      setStats(prev => ({
        ...prev,
        totalProviders: providersData?.length || 0,
        availableGPUs: totalGPUs
      }));
    } catch (error) {
      console.error('Failed to load providers:', error);
      // Mock data for demo
      setProviders([
        {
          id: '1',
          name: 'CloudGPU Pro',
          status: 'online',
          location: 'US-East',
          gpus: [
            {
              model_name: 'NVIDIA RTX 4090',
              vram_mb: 24576,
              driver_version: '536.23',
              is_healthy: true,
              utilization_gpu_percent: 75,
              temperature_c: 68,
              power_draw_w: 320
            }
          ]
        },
        {
          id: '2',
          name: 'AI Compute Hub',
          status: 'online',
          location: 'EU-West',
          gpus: [
            {
              model_name: 'NVIDIA A100',
              vram_mb: 81920,
              driver_version: '535.86',
              is_healthy: true,
              utilization_gpu_percent: 90,
              temperature_c: 72,
              power_draw_w: 400
            }
          ]
        }
      ]);
      setStats(prev => ({ ...prev, totalProviders: 2, availableGPUs: 2 }));
    }
  };

  const loadActiveJobs = async () => {
    // Mock data for active jobs
    setActiveJobs([
      {
        id: 'job-001',
        name: 'AI Training - ResNet50',
        provider: 'CloudGPU Pro',
        status: 'running',
        startTime: '2024-01-15T10:30:00Z',
        duration: '2h 45m',
        cost: '1.25',
        gpuModel: 'NVIDIA RTX 4090'
      },
      {
        id: 'job-002',
        name: 'Video Rendering',
        provider: 'AI Compute Hub',
        status: 'pending',
        startTime: '2024-01-15T12:00:00Z',
        duration: '0m',
        cost: '0.00',
        gpuModel: 'NVIDIA A100'
      }
    ]);
    setStats(prev => ({ ...prev, activeJobs: 2 }));
  };

  const loadTransactions = async () => {
    // Mock transaction data
    setTransactions([
      {
        id: 'tx-001',
        type: 'payment',
        amount: '2.50',
        description: 'GPU Rental Payment - RTX 4090',
        timestamp: '2024-01-15T14:30:00Z',
        status: 'completed'
      },
      {
        id: 'tx-002',
        type: 'earning',
        amount: '5.75',
        description: 'Provider Earnings - A100',
        timestamp: '2024-01-15T13:45:00Z',
        status: 'completed'
      }
    ]);
  };

  const loadStats = async () => {
    setStats(prev => ({
      ...prev,
      totalEarnings: '125.50',
      walletBalance: '47.25'
    }));
  };

  const loadGPUMetrics = async () => {
    // Simulate real GPU metrics
    setGpuMetrics({
      utilization: 82,
      temperature: 70,
      powerDraw: 285,
      memoryUsage: 78,
      vramTotal: 24576,
      vramUsed: 19200
    });
  };

  const handleLogout = () => {
    logout();
    router.push('/');
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-100 text-green-800';
      case 'running': return 'bg-blue-100 text-blue-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
      case 'running':
      case 'completed':
        return <CheckCircle className="w-4 h-4" />;
      case 'pending':
        return <Clock className="w-4 h-4" />;
      case 'failed':
        return <XCircle className="w-4 h-4" />;
      default:
        return <AlertCircle className="w-4 h-4" />;
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-professional flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-professional">
      {/* Navigation Header */}
      <nav className="bg-card border-b-2 border-black">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <div className="w-8 h-8 logo-container flex items-center justify-center">
                <Cpu className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">DanteGPU Dashboard</h1>
                <p className="text-xs text-gray-600">Welcome back, {user?.username}</p>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                <Bell className="w-4 h-4 mr-2" />
                Notifications
              </Button>
              <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                <Settings className="w-4 h-4 mr-2" />
                Settings
              </Button>
              <Button 
                onClick={handleLogout}
                variant="outline" 
                size="sm" 
                className="border-2 border-red-500 text-red-600 hover:bg-red-50 btn-hover-professional"
              >
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Providers</p>
                  <p className="text-3xl font-bold text-gray-900">{stats.totalProviders}</p>
                </div>
                <Server className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Available GPUs</p>
                  <p className="text-3xl font-bold text-gray-900">{stats.availableGPUs}</p>
                </div>
                <Monitor className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Active Jobs</p>
                  <p className="text-3xl font-bold text-gray-900">{stats.activeJobs}</p>
                </div>
                <Activity className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Wallet Balance</p>
                  <p className="text-3xl font-bold text-gray-900">{stats.walletBalance} dGPU</p>
                </div>
                <Wallet className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Providers & Jobs */}
          <div className="lg:col-span-2 space-y-6">
            {/* Available Providers */}
            <Card className="card-professional">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center space-x-2">
                    <Server className="w-5 h-5 text-primary" />
                    <span>Available Providers</span>
                  </CardTitle>
                  <Button 
                    size="sm" 
                    variant="outline" 
                    onClick={loadProviders}
                    className="border-2 border-black btn-hover-professional"
                  >
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Refresh
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {providers.map((provider) => (
                    <div key={provider.id} className="border-2 border-black rounded-lg p-4 bg-secondary">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center space-x-3">
                          <h3 className="font-semibold text-gray-900">{provider.name}</h3>
                          <Badge className={`badge-professional ${getStatusColor(provider.status)}`}>
                            {getStatusIcon(provider.status)}
                            <span className="ml-1">{provider.status}</span>
                          </Badge>
                        </div>
                        <div className="flex items-center text-sm text-gray-600">
                          <MapPin className="w-4 h-4 mr-1" />
                          {provider.location}
                        </div>
                      </div>
                      
                      {provider.gpus?.map((gpu: any, index: number) => (
                        <div key={index} className="bg-card border border-black rounded-md p-3 mt-2">
                          <div className="flex items-center justify-between mb-2">
                            <span className="font-medium text-gray-900">{gpu.model_name}</span>
                            <Badge className={`badge-professional ${gpu.is_healthy ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                              {gpu.is_healthy ? 'Healthy' : 'Issue'}
                            </Badge>
                          </div>
                          
                          <div className="grid grid-cols-3 gap-4 text-sm">
                            <div>
                              <p className="text-gray-600">VRAM</p>
                              <p className="font-medium">{Math.round(gpu.vram_mb / 1024)} GB</p>
                            </div>
                            <div>
                              <p className="text-gray-600">Utilization</p>
                              <p className="font-medium">{gpu.utilization_gpu_percent || 0}%</p>
                            </div>
                            <div>
                              <p className="text-gray-600">Temperature</p>
                              <p className="font-medium">{gpu.temperature_c || 0}°C</p>
                            </div>
                          </div>
                          
                          <div className="mt-2">
                            <div className="flex justify-between text-sm mb-1">
                              <span>GPU Utilization</span>
                              <span>{gpu.utilization_gpu_percent || 0}%</span>
                            </div>
                            <Progress value={gpu.utilization_gpu_percent || 0} className="h-2" />
                          </div>
                        </div>
                      ))}
                      
                      <div className="flex space-x-2 mt-4">
                        <Button size="sm" className="bg-primary hover:bg-primary/90 border border-black btn-hover-professional">
                          <Play className="w-4 h-4 mr-2" />
                          Rent GPU
                        </Button>
                        <Button size="sm" variant="outline" className="border-2 border-black btn-hover-professional">
                          <Eye className="w-4 h-4 mr-2" />
                          View Details
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Active Jobs */}
            <Card className="card-professional">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Activity className="w-5 h-5 text-primary" />
                  <span>Active Jobs</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {activeJobs.map((job) => (
                    <div key={job.id} className="border-2 border-black rounded-lg p-4 bg-secondary">
                      <div className="flex items-center justify-between mb-2">
                        <h3 className="font-semibold text-gray-900">{job.name}</h3>
                        <Badge className={`badge-professional ${getStatusColor(job.status)}`}>
                          {getStatusIcon(job.status)}
                          <span className="ml-1">{job.status}</span>
                        </Badge>
                      </div>
                      
                      <div className="grid grid-cols-2 gap-4 text-sm text-gray-600 mb-3">
                        <div>Provider: {job.provider}</div>
                        <div>GPU: {job.gpuModel}</div>
                        <div>Duration: {job.duration}</div>
                        <div>Cost: {job.cost} dGPU</div>
                      </div>
                      
                      <div className="flex space-x-2">
                        <Button size="sm" variant="outline" className="border-2 border-black btn-hover-professional">
                          <Pause className="w-4 h-4 mr-2" />
                          Pause
                        </Button>
                        <Button size="sm" variant="outline" className="border-2 border-red-500 text-red-600 hover:bg-red-50 btn-hover-professional">
                          <Square className="w-4 h-4 mr-2" />
                          Stop
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Metrics & Transactions */}
          <div className="space-y-6">
            {/* GPU Metrics */}
            <Card className="card-professional">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <BarChart3 className="w-5 h-5 text-primary" />
                  <span>GPU Metrics</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span>GPU Utilization</span>
                      <span>{gpuMetrics.utilization}%</span>
                    </div>
                    <Progress value={gpuMetrics.utilization} className="h-2" />
                  </div>
                  
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span>Memory Usage</span>
                      <span>{gpuMetrics.memoryUsage}%</span>
                    </div>
                    <Progress value={gpuMetrics.memoryUsage} className="h-2" />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4 pt-4 border-t-2 border-black">
                    <div className="text-center">
                      <div className="flex items-center justify-center space-x-2 mb-1">
                        <Thermometer className="w-4 h-4 text-orange-500" />
                        <span className="text-sm text-gray-600">Temperature</span>
                      </div>
                      <p className="text-lg font-bold text-gray-900">{gpuMetrics.temperature}°C</p>
                    </div>
                    
                    <div className="text-center">
                      <div className="flex items-center justify-center space-x-2 mb-1">
                        <Zap className="w-4 h-4 text-yellow-500" />
                        <span className="text-sm text-gray-600">Power</span>
                      </div>
                      <p className="text-lg font-bold text-gray-900">{gpuMetrics.powerDraw}W</p>
                    </div>
                  </div>
                  
                  <div className="pt-2">
                    <div className="flex items-center justify-between text-sm text-gray-600 mb-1">
                      <span>VRAM Usage</span>
                      <span>{Math.round(gpuMetrics.vramUsed / 1024)} / {Math.round(gpuMetrics.vramTotal / 1024)} GB</span>
                    </div>
                    <Progress value={(gpuMetrics.vramUsed / gpuMetrics.vramTotal) * 100} className="h-2" />
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Recent Transactions */}
            <Card className="card-professional">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <CreditCard className="w-5 h-5 text-primary" />
                  <span>Recent Transactions</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {transactions.map((tx) => (
                    <div key={tx.id} className="border-2 border-black rounded-lg p-3 bg-secondary">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium text-gray-900">
                          {tx.type === 'payment' ? '-' : '+'}{tx.amount} dGPU
                        </span>
                        <Badge className={`badge-professional ${getStatusColor(tx.status)}`}>
                          {tx.status}
                        </Badge>
                      </div>
                      <p className="text-sm text-gray-600">{tx.description}</p>
                      <p className="text-xs text-gray-500 mt-1">
                        {new Date(tx.timestamp).toLocaleString()}
                      </p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card className="card-professional">
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Button className="w-full bg-primary hover:bg-primary/90 border-2 border-black btn-hover-professional">
                    <DollarSign className="w-4 h-4 mr-2" />
                    Top Up Wallet
                  </Button>
                  <Button variant="outline" className="w-full border-2 border-black btn-hover-professional">
                    <Play className="w-4 h-4 mr-2" />
                    Start New Job
                  </Button>
                  <Button variant="outline" className="w-full border-2 border-black btn-hover-professional">
                    <TrendingUp className="w-4 h-4 mr-2" />
                    View Analytics
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
} 