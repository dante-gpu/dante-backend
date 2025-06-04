'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts';
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
  CreditCard,
  MemoryStick,
  Gauge,
  Flame,
  Fan,
  Download,
  Upload,
  Wifi,
  Calendar,
  Plus,
  Minus,
  Home,
  Database,
  Network,
  Computer
} from 'lucide-react';
import { dashboardService, gpuService } from '@/lib/api';

// Real Data Interfaces
interface DashboardStats {
  totalProviders: number;
  availableGPUs: number;
  activeJobs: number;
  totalEarnings: string;
  walletBalance: string;
  totalSpent: string;
  jobsCompleted: number;
  computeHours: number;
}

interface RealGPUDevice {
  device_id: string;
  name: string;
  vendor: string;
  driver_version: string;
  memory_total_mb: number;
  memory_used_mb: number;
  memory_free_mb: number;
  utilization_gpu: number;
  utilization_memory: number;
  temperature_c: number;
  power_draw_w: number;
  clock_graphics_mhz: number;
  clock_memory_mhz: number;
  architecture: string;
  performance_score: number;
  is_available_for_rent: boolean;
}

interface SystemInfo {
  hostname: string;
  os_type: string;
  os_version: string;
  cpu_model: string;
  cpu_cores: number;
  ram_total_gb: number;
  ram_available_gb: number;
  disk_total_gb: number;
  disk_free_gb: number;
  uptime_seconds: number;
  network_interfaces: any[];
}

interface GPUMetric {
  timestamp: string;
  utilization_gpu: number;
  utilization_memory: number;
  temperature_c: number;
  power_draw_w: number;
  memory_used_mb: number;
  clock_graphics_mhz: number;
}

interface ActiveJob {
  id: string;
  name: string;
  provider: string;
  status: string;
  startTime: string;
  duration: string;
  cost: string;
  gpuModel: string;
  progress: number;
  description?: string;
}

interface RecentTransaction {
  id: string;
  type: string;
  amount: string;
  description: string;
  timestamp: string;
  status: string;
  tx_hash?: string;
}

interface Provider {
  id: string;
  name: string;
  location: string;
  status: string;
  hourlyRate: number;
  rating: number;
  totalJobs: number;
  successRate: number;
  gpus: any[];
  lastSeen: string;
}

// Progress component
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

export default function RealDashboardPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Real data states
  const [stats, setStats] = useState<DashboardStats>({
    totalProviders: 0,
    availableGPUs: 0,
    activeJobs: 0,
    totalEarnings: '0.00',
    walletBalance: '0.00',
    totalSpent: '0.00',
    jobsCompleted: 0,
    computeHours: 0.0
  });
  
  const [realGPUs, setRealGPUs] = useState<RealGPUDevice[]>([]);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [gpuMetrics, setGpuMetrics] = useState<GPUMetric[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [activeJobs, setActiveJobs] = useState<ActiveJob[]>([]);
  const [transactions, setTransactions] = useState<RecentTransaction[]>([]);
  const [selectedGPU, setSelectedGPU] = useState<string | null>(null);
  const [rentalRate, setRentalRate] = useState<number>(1.0);

  useEffect(() => {
    loadAllRealData();
    
    // Set up real-time updates every 15 seconds
    const interval = setInterval(loadAllRealData, 15000);
    return () => clearInterval(interval);
  }, []);

  const loadAllRealData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Load all real data in parallel
      const [
        statsData, 
        providersData, 
        jobsData, 
        transactionsData,
        realGPUData,
        systemData
      ] = await Promise.all([
        dashboardService.getStats(),
        dashboardService.getProviders(),
        dashboardService.getJobs(undefined, 10),
        dashboardService.getTransactions(undefined, 10),
        gpuService.detectGPUs().catch(() => []),
        gpuService.getSystemInfo().catch(() => null)
      ]);

      setStats(statsData);
      setProviders(providersData);
      setActiveJobs(jobsData as ActiveJob[]);
      setTransactions(transactionsData as RecentTransaction[]);
      setRealGPUs(realGPUData);
      setSystemInfo(systemData);

      // Load GPU metrics for the first detected GPU
      if (realGPUData.length > 0 && !selectedGPU) {
        const firstGPU = realGPUData[0];
        setSelectedGPU(firstGPU.device_id);
        
        try {
          const metrics = await gpuService.getGPUMetrics(firstGPU.device_id, 24);
          setGpuMetrics(metrics);
        } catch (error) {
          console.error('Failed to load GPU metrics:', error);
        }
      }
      
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      setError('Failed to load real dashboard data. Please check your connection.');
    } finally {
      setLoading(false);
    }
  };

  const handleGPUSelection = async (deviceId: string) => {
    setSelectedGPU(deviceId);
    try {
      const metrics = await gpuService.getGPUMetrics(deviceId, 24);
      setGpuMetrics(metrics);
    } catch (error) {
      console.error('Failed to load GPU metrics:', error);
    }
  };

  const handleRegisterGPUForRent = async (deviceId: string) => {
    try {
      await gpuService.registerGPUForRent(deviceId, rentalRate);
      await loadAllRealData(); // Refresh data
      alert('GPU successfully registered for rent!');
    } catch (error) {
      console.error('Failed to register GPU:', error);
      alert('Failed to register GPU for rent. Please try again.');
    }
  };

  const handleJobAction = async (jobId: string, action: string) => {
    try {
      await dashboardService.performJobAction(jobId, action);
      await loadAllRealData(); // Refresh data
    } catch (error) {
      console.error(`Failed to ${action} job:`, error);
      setError(`Failed to ${action} job. Please try again.`);
    }
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

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${minutes}m`;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  // Chart data preparation
  const prepareGPUChartData = () => {
    return gpuMetrics.slice(-24).map(metric => ({
      time: new Date(metric.timestamp).toLocaleTimeString(),
      gpu: metric.utilization_gpu,
      memory: metric.utilization_memory,
      temp: metric.temperature_c,
      power: metric.power_draw_w,
      clock: metric.clock_graphics_mhz / 100 // Scale down for chart
    }));
  };

  const prepareSystemResourceData = () => {
    if (!systemInfo) return [];
    
    return [
      { name: 'RAM', used: systemInfo.ram_total_gb - systemInfo.ram_available_gb, total: systemInfo.ram_total_gb },
      { name: 'Disk', used: systemInfo.disk_total_gb - systemInfo.disk_free_gb, total: systemInfo.disk_total_gb }
    ];
  };

  const CHART_COLORS = ['#ad0000', '#e2e0d0', '#ff6b6b', '#4ecdc4', '#45b7d1', '#f39c12'];

  if (loading) {
    return (
      <div className="min-h-screen bg-professional flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-gray-600">Loading real dashboard data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-professional flex items-center justify-center">
        <div className="text-center">
          <AlertCircle className="w-8 h-8 mx-auto mb-4 text-red-500" />
          <p className="text-red-600 mb-4">{error}</p>
          <Button onClick={loadAllRealData} className="bg-primary hover:bg-primary/90">
            Retry Loading Real Data
          </Button>
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
                <h1 className="text-xl font-bold text-gray-900">DanteGPU Real Dashboard</h1>
                <p className="text-xs text-gray-600">
                  {systemInfo ? `${systemInfo.hostname} - ${systemInfo.os_type} ${systemInfo.os_version}` : 'Loading system info...'}
                </p>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              <Button 
                variant="outline" 
                size="sm" 
                className="border-2 border-black btn-hover-professional"
                onClick={loadAllRealData}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh Real Data
              </Button>
              <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                <Bell className="w-4 h-4 mr-2" />
                Alerts
              </Button>
              <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                <Settings className="w-4 h-4 mr-2" />
                System Settings
              </Button>
              <Button 
                onClick={handleLogout}
                variant="outline" 
                size="sm" 
                className="border-2 border-red-500 text-red-600 hover:bg-red-50"
              >
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Real System Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Real GPUs Detected</p>
                  <p className="text-3xl font-bold text-gray-900">{realGPUs.length}</p>
                  <p className="text-xs text-gray-500">
                    {realGPUs.filter(gpu => gpu.is_available_for_rent).length} available for rent
                  </p>
                </div>
                <Monitor className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">System Uptime</p>
                  <p className="text-lg font-bold text-gray-900">
                    {systemInfo ? formatUptime(systemInfo.uptime_seconds) : 'Loading...'}
                  </p>
                  <p className="text-xs text-gray-500">
                    {systemInfo?.cpu_cores} CPU cores
                  </p>
                </div>
                <Clock className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Memory</p>
                  <p className="text-2xl font-bold text-gray-900">
                    {systemInfo ? `${systemInfo.ram_total_gb.toFixed(1)} GB` : 'Loading...'}
                  </p>
                  <p className="text-xs text-gray-500">
                    {systemInfo ? `${systemInfo.ram_available_gb.toFixed(1)} GB available` : ''}
                  </p>
                </div>
                <MemoryStick className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>

          <Card className="card-professional">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Storage</p>
                  <p className="text-2xl font-bold text-gray-900">
                    {systemInfo ? `${systemInfo.disk_total_gb.toFixed(0)} GB` : 'Loading...'}
                  </p>
                  <p className="text-xs text-gray-500">
                    {systemInfo ? `${systemInfo.disk_free_gb.toFixed(0)} GB free` : ''}
                  </p>
                </div>
                <HardDrive className="w-8 h-8 text-primary" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Real GPU Monitoring Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* GPU Performance Charts */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <BarChart3 className="w-5 h-5 text-primary" />
                <span>Real GPU Performance Metrics</span>
              </CardTitle>
              <CardDescription>
                Live monitoring of your actual GPU hardware
                {selectedGPU && (
                  <span className="block text-primary font-semibold mt-1">
                    {realGPUs.find(g => g.device_id === selectedGPU)?.name}
                  </span>
                )}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {gpuMetrics.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={prepareGPUChartData()}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e0d0" />
                    <XAxis dataKey="time" stroke="#666" />
                    <YAxis stroke="#666" />
                    <Tooltip 
                      contentStyle={{ 
                        backgroundColor: '#e2e0d0', 
                        border: '2px solid black',
                        borderRadius: '4px'
                      }} 
                    />
                    <Legend />
                    <Line type="monotone" dataKey="gpu" stroke="#ad0000" strokeWidth={2} name="GPU Usage %" />
                    <Line type="monotone" dataKey="memory" stroke="#4ecdc4" strokeWidth={2} name="Memory %" />
                    <Line type="monotone" dataKey="temp" stroke="#ff6b6b" strokeWidth={2} name="Temperature °C" />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex items-center justify-center h-300 text-gray-500">
                  <div className="text-center">
                    <Gauge className="w-12 h-12 mx-auto mb-4" />
                    <p>No GPU metrics available yet</p>
                    <p className="text-sm">Waiting for data collection...</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* System Resource Usage */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Monitor className="w-5 h-5 text-primary" />
                <span>System Resource Usage</span>
              </CardTitle>
              <CardDescription>Real-time system resource monitoring</CardDescription>
            </CardHeader>
            <CardContent>
              {systemInfo ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={prepareSystemResourceData()}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e2e0d0" />
                    <XAxis dataKey="name" stroke="#666" />
                    <YAxis stroke="#666" />
                    <Tooltip 
                      contentStyle={{ 
                        backgroundColor: '#e2e0d0', 
                        border: '2px solid black',
                        borderRadius: '4px'
                      }} 
                    />
                    <Bar dataKey="used" fill="#ad0000" name="Used GB" />
                    <Bar dataKey="total" fill="#e2e0d0" name="Total GB" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex items-center justify-center h-300 text-gray-500">
                  <div className="text-center">
                    <Database className="w-12 h-12 mx-auto mb-4" />
                    <p>Loading system information...</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Real GPU Device Management */}
        <Card className="card-professional mb-8">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Monitor className="w-5 h-5 text-primary" />
              <span>Your Real GPU Devices</span>
            </CardTitle>
            <CardDescription>
              Manage and rent out your actual detected GPU hardware
            </CardDescription>
          </CardHeader>
          <CardContent>
            {realGPUs.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {realGPUs.map((gpu) => (
                  <div 
                    key={gpu.device_id} 
                    className={`p-6 bg-secondary border-2 rounded-lg cursor-pointer transition-all ${
                      selectedGPU === gpu.device_id 
                        ? 'border-primary bg-primary/5' 
                        : 'border-black hover:border-primary'
                    }`}
                    onClick={() => handleGPUSelection(gpu.device_id)}
                  >
                    <div className="flex justify-between items-start mb-4">
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900">{gpu.name}</h3>
                        <p className="text-sm text-gray-600">{gpu.vendor} • {gpu.architecture}</p>
                        <p className="text-xs text-gray-500">Driver: {gpu.driver_version}</p>
                      </div>
                      <Badge className={`badge-professional ${gpu.is_available_for_rent ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                        {gpu.is_available_for_rent ? 'Available for Rent' : 'Private'}
                      </Badge>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4 mb-4">
                      <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">GPU Usage:</span>
                          <span className="font-medium">{gpu.utilization_gpu.toFixed(1)}%</span>
                        </div>
                        <Progress value={gpu.utilization_gpu} className="w-full" />
                        
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">Memory:</span>
                          <span className="font-medium">{gpu.utilization_memory.toFixed(1)}%</span>
                        </div>
                        <Progress value={gpu.utilization_memory} className="w-full" />
                      </div>
                      
                      <div className="space-y-3">
                        <div className="flex items-center space-x-2">
                          <Thermometer className="w-4 h-4 text-orange-500" />
                          <span className="text-sm">{gpu.temperature_c.toFixed(1)}°C</span>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Zap className="w-4 h-4 text-yellow-500" />
                          <span className="text-sm">{gpu.power_draw_w.toFixed(1)}W</span>
                        </div>
                        <div className="flex items-center space-x-2">
                          <MemoryStick className="w-4 h-4 text-blue-500" />
                          <span className="text-sm">{(gpu.memory_total_mb / 1024).toFixed(1)} GB</span>
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex justify-between items-center pt-4 border-t-2 border-black">
                      <div>
                        <p className="text-sm text-gray-600">Performance Score</p>
                        <p className="text-lg font-bold text-primary">{gpu.performance_score.toFixed(1)}/100</p>
                      </div>
                      
                      {!gpu.is_available_for_rent ? (
                        <div className="flex items-center space-x-2">
                          <input
                            type="number"
                            value={rentalRate}
                            onChange={(e) => setRentalRate(parseFloat(e.target.value))}
                            className="w-20 px-2 py-1 border border-black rounded text-sm"
                            placeholder="Rate"
                            step="0.1"
                            min="0.1"
                          />
                          <Button 
                            size="sm" 
                            className="bg-primary hover:bg-primary/90"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleRegisterGPUForRent(gpu.device_id);
                            }}
                          >
                            <Plus className="w-4 h-4 mr-1" />
                            Rent Out
                          </Button>
                        </div>
                      ) : (
                        <Badge className="badge-professional bg-green-100 text-green-800">
                          <DollarSign className="w-3 h-3 mr-1" />
                          ${rentalRate}/hour
                        </Badge>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <Monitor className="w-16 h-16 mx-auto mb-4 text-gray-400" />
                <h3 className="text-lg font-semibold text-gray-600 mb-2">No GPUs Detected</h3>
                <p className="text-gray-500 mb-4">
                  We couldn't detect any GPU hardware on your system.
                </p>
                <Button onClick={loadAllRealData} variant="outline" className="border-2 border-black">
                  <RefreshCw className="w-4 h-4 mr-2" />
                  Retry Detection
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Network Activity and Jobs */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Active Jobs */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Activity className="w-5 h-5 text-primary" />
                  <span>Active Compute Jobs</span>
                </div>
                <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                  <Eye className="w-4 h-4 mr-2" />
                  View All
                </Button>
              </CardTitle>
              <CardDescription>Real jobs running on the network</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {activeJobs.length === 0 ? (
                  <p className="text-gray-500 text-center py-4">No active jobs</p>
                ) : (
                  activeJobs.map((job) => (
                    <div key={job.id} className="p-4 bg-secondary border-2 border-black rounded-lg">
                      <div className="flex justify-between items-start mb-2">
                        <div>
                          <h4 className="font-semibold text-gray-900">{job.name}</h4>
                          <p className="text-sm text-gray-600">{job.provider} • {job.gpuModel}</p>
                        </div>
                        <Badge className={`badge-professional ${getStatusColor(job.status)}`}>
                          {getStatusIcon(job.status)}
                          <span className="ml-1">{job.status}</span>
                        </Badge>
                      </div>
                      
                      {job.status === 'running' && (
                        <div className="mb-3">
                          <div className="flex justify-between text-xs text-gray-600 mb-1">
                            <span>Progress</span>
                            <span>{job.progress.toFixed(1)}%</span>
                          </div>
                          <Progress value={job.progress} className="w-full" />
                        </div>
                      )}
                      
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-gray-600">Duration: {job.duration}</span>
                        <span className="font-semibold">{job.cost} dGPU</span>
                      </div>
                      
                      <div className="flex space-x-2 mt-3">
                        {job.status === 'running' && (
                          <Button 
                            size="sm" 
                            variant="outline" 
                            className="border border-black text-xs"
                            onClick={() => handleJobAction(job.id, 'pause')}
                          >
                            <Pause className="w-3 h-3 mr-1" />
                            Pause
                          </Button>
                        )}
                        {job.status === 'pending' && (
                          <Button 
                            size="sm" 
                            variant="outline" 
                            className="border border-black text-xs"
                            onClick={() => handleJobAction(job.id, 'resume')}
                          >
                            <Play className="w-3 h-3 mr-1" />
                            Resume
                          </Button>
                        )}
                        <Button 
                          size="sm" 
                          variant="outline" 
                          className="border border-red-500 text-red-600 text-xs"
                          onClick={() => handleJobAction(job.id, 'cancel')}
                        >
                          <Square className="w-3 h-3 mr-1" />
                          Cancel
                        </Button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>

          {/* Network Information */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Network className="w-5 h-5 text-primary" />
                <span>Network Interfaces</span>
              </CardTitle>
              <CardDescription>Your system's network connectivity</CardDescription>
            </CardHeader>
            <CardContent>
              {systemInfo?.network_interfaces ? (
                <div className="space-y-4">
                  {systemInfo.network_interfaces.slice(0, 4).map((iface, index) => (
                    <div key={index} className="p-3 bg-secondary border border-black rounded">
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="font-medium text-gray-900">{iface.name}</h4>
                        <Badge className="badge-professional bg-blue-100 text-blue-800">
                          <Wifi className="w-3 h-3 mr-1" />
                          Active
                        </Badge>
                      </div>
                      <div className="space-y-1">
                        {iface.addresses.slice(0, 2).map((addr: string, addrIndex: number) => (
                          <p key={addrIndex} className="text-xs text-gray-600 font-mono">{addr}</p>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <Network className="w-12 h-12 mx-auto mb-4" />
                  <p>Loading network information...</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Financial Overview */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Wallet and Earnings */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Wallet className="w-5 h-5 text-primary" />
                <span>Wallet & Earnings</span>
              </CardTitle>
              <CardDescription>Your real financial overview</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-4 bg-secondary border border-black rounded">
                    <p className="text-sm text-gray-600">Current Balance</p>
                    <p className="text-2xl font-bold text-primary">{stats.walletBalance} dGPU</p>
                  </div>
                  <div className="p-4 bg-secondary border border-black rounded">
                    <p className="text-sm text-gray-600">Total Earned</p>
                    <p className="text-2xl font-bold text-green-600">{stats.totalEarnings} dGPU</p>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-4 bg-secondary border border-black rounded">
                    <p className="text-sm text-gray-600">Total Spent</p>
                    <p className="text-xl font-bold text-red-600">{stats.totalSpent} dGPU</p>
                  </div>
                  <div className="p-4 bg-secondary border border-black rounded">
                    <p className="text-sm text-gray-600">Jobs Completed</p>
                    <p className="text-xl font-bold text-gray-900">{stats.jobsCompleted}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Recent Transactions */}
          <Card className="card-professional">
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <CreditCard className="w-5 h-5 text-primary" />
                  <span>Recent Transactions</span>
                </div>
                <Button variant="outline" size="sm" className="border-2 border-black btn-hover-professional">
                  <Eye className="w-4 h-4 mr-2" />
                  View All
                </Button>
              </CardTitle>
              <CardDescription>Real transaction history</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {transactions.length === 0 ? (
                  <p className="text-gray-500 text-center py-4">No recent transactions</p>
                ) : (
                  transactions.map((tx) => (
                    <div key={tx.id} className="flex justify-between items-center p-3 bg-secondary border border-black rounded">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <Badge className={`badge-professional ${getStatusColor(tx.type)}`}>
                            {tx.type}
                          </Badge>
                          <Badge className={`badge-professional ${getStatusColor(tx.status)}`}>
                            {tx.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-gray-900 font-medium">{tx.description}</p>
                        <p className="text-xs text-gray-600">{formatTimestamp(tx.timestamp)}</p>
                        {tx.tx_hash && (
                          <p className="text-xs text-gray-500 font-mono">{tx.tx_hash.substring(0, 20)}...</p>
                        )}
                      </div>
                      <div className="text-right">
                        <p className={`font-bold ${tx.type === 'payment' ? 'text-red-600' : 'text-green-600'}`}>
                          {tx.type === 'payment' ? '-' : '+'}{tx.amount} dGPU
                        </p>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
} 