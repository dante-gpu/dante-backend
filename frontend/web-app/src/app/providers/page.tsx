'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { providerService, type Provider, type GPUDetail } from '@/lib/api';
import {
  Search,
  Filter,
  MapPin,
  Cpu,
  Zap,
  CheckCircle,
  Clock,
  AlertCircle,
  TrendingUp,
  Activity,
  Wifi,
  Settings,
  BarChart3
} from 'lucide-react';

interface ProviderFilters {
  search: string;
  location: string;
  status: string;
  gpuModel: string;
  sortBy: string;
}

export default function ProvidersPage() {
  const { user } = useAuth();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [filteredProviders, setFilteredProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<ProviderFilters>({
    search: '',
    location: '',
    status: '',
    gpuModel: '',
    sortBy: 'name'
  });

  useEffect(() => {
    loadProviders();
  }, []);

  useEffect(() => {
    filterProviders();
  }, [providers, filters]);

  const loadProviders = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const providersData = await providerService.listProviders();
      setProviders(providersData);
    } catch (error) {
      console.error('Failed to load providers:', error);
      setError('Failed to load providers');
      
      // Use fallback demo data if API fails
      const fallbackProviders: Provider[] = [
        {
          id: 'provider_1',
          owner_id: 'owner_1',
          name: 'CloudGPU Pro',
          hostname: 'cloudgpu-pro.com',
          ip_address: '192.168.1.100',
          status: 'online',
          location: 'US-East-1',
          gpus: [
            {
              model_name: 'NVIDIA RTX 4090',
              vram_mb: 24576,
              driver_version: '536.23',
              architecture: 'Ada Lovelace',
              compute_capability: '8.9',
              cuda_cores: 16384,
              tensor_cores: 512,
              memory_bandwidth_gb_s: 1008,
              power_consumption_w: 450,
              utilization_gpu_percent: 0,
              utilization_memory_percent: 0,
              temperature_c: 35,
              power_draw_w: 50,
              is_healthy: true
            }
          ],
          registered_at: '2024-01-15T10:30:00Z',
          last_seen_at: new Date().toISOString(),
          metadata: {
            rating: 4.8,
            hourly_rate: 2.5,
            total_jobs: 156,
            uptime: 99.5
          }
        },
        {
          id: 'provider_2',
          owner_id: 'owner_2',
          name: 'AI Compute Hub',
          hostname: 'ai-compute.net',
          ip_address: '192.168.1.101',
          status: 'online',
          location: 'EU-West-1',
          gpus: [
            {
              model_name: 'NVIDIA A100',
              vram_mb: 81920,
              driver_version: '535.183',
              architecture: 'Ampere',
              compute_capability: '8.0',
              cuda_cores: 6912,
              tensor_cores: 432,
              memory_bandwidth_gb_s: 1935,
              power_consumption_w: 400,
              utilization_gpu_percent: 65,
              utilization_memory_percent: 45,
              temperature_c: 68,
              power_draw_w: 320,
              is_healthy: true
            },
            {
              model_name: 'NVIDIA A100',
              vram_mb: 81920,
              driver_version: '535.183',
              architecture: 'Ampere',
              compute_capability: '8.0',
              cuda_cores: 6912,
              tensor_cores: 432,
              memory_bandwidth_gb_s: 1935,
              power_consumption_w: 400,
              utilization_gpu_percent: 0,
              utilization_memory_percent: 0,
              temperature_c: 42,
              power_draw_w: 75,
              is_healthy: true
            }
          ],
          registered_at: '2024-02-01T14:20:00Z',
          last_seen_at: new Date().toISOString(),
          metadata: {
            rating: 4.6,
            hourly_rate: 4.2,
            total_jobs: 89,
            uptime: 97.8
          }
        },
        {
          id: 'provider_3',
          owner_id: 'owner_3',
          name: 'RenderFarm Elite',
          hostname: 'renderfarm-elite.io',
          ip_address: '192.168.1.102',
          status: 'busy',
          location: 'US-West-2',
          gpus: [
            {
              model_name: 'NVIDIA RTX 3080',
              vram_mb: 10240,
              driver_version: '531.79',
              architecture: 'Ampere',
              compute_capability: '8.6',
              cuda_cores: 8704,
              tensor_cores: 272,
              memory_bandwidth_gb_s: 760,
              power_consumption_w: 320,
              utilization_gpu_percent: 95,
              utilization_memory_percent: 87,
              temperature_c: 78,
              power_draw_w: 310,
              is_healthy: true
            }
          ],
          registered_at: '2024-01-28T08:45:00Z',
          last_seen_at: new Date().toISOString(),
          metadata: {
            rating: 4.2,
            hourly_rate: 1.8,
            total_jobs: 234,
            uptime: 95.2
          }
        }
      ];
      
      setProviders(fallbackProviders);
    } finally {
      setIsLoading(false);
    }
  };

  const filterProviders = () => {
    let filtered = [...providers];
    
    // Search filter
    if (filters.search) {
      filtered = filtered.filter(provider =>
        provider.name.toLowerCase().includes(filters.search.toLowerCase()) ||
        provider.location?.toLowerCase().includes(filters.search.toLowerCase()) ||
        provider.gpus.some(gpu => gpu.model_name.toLowerCase().includes(filters.search.toLowerCase()))
      );
    }
    
    // Location filter
    if (filters.location) {
      filtered = filtered.filter(provider => provider.location === filters.location);
    }
    
    // Status filter
    if (filters.status) {
      filtered = filtered.filter(provider => provider.status === filters.status);
    }
    
    // GPU Model filter
    if (filters.gpuModel) {
      filtered = filtered.filter(provider =>
        provider.gpus.some(gpu => gpu.model_name.includes(filters.gpuModel))
      );
    }
    
    // Sort
    filtered.sort((a, b) => {
      switch (filters.sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'rating':
          return (b.metadata?.rating || 0) - (a.metadata?.rating || 0);
        case 'price':
          return (a.metadata?.hourly_rate || 0) - (b.metadata?.hourly_rate || 0);
        case 'location':
          return (a.location || '').localeCompare(b.location || '');
        default:
          return 0;
      }
    });
    
    setFilteredProviders(filtered);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-100 text-green-800';
      case 'busy': return 'bg-yellow-100 text-yellow-800';
      case 'offline': return 'bg-red-100 text-red-800';
      case 'maintenance': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online': return <CheckCircle className="w-3 h-3" />;
      case 'busy': return <Activity className="w-3 h-3" />;
      case 'offline': return <AlertCircle className="w-3 h-3" />;
      case 'maintenance': return <Settings className="w-3 h-3" />;
      default: return <Clock className="w-3 h-3" />;
    }
  };

  const formatLastSeen = (lastSeenAt: string) => {
    const lastSeen = new Date(lastSeenAt);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - lastSeen.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 1) return 'Just now';
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`;
    return `${Math.floor(diffInMinutes / 1440)}d ago`;
  };

  const getAllLocations = () => {
    const locations = providers.map(p => p.location).filter(Boolean);
    return [...new Set(locations)];
  };

  const getAllGPUModels = () => {
    const models = providers.flatMap(p => p.gpus.map(gpu => gpu.model_name));
    return [...new Set(models)];
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-gray-600">Loading providers...</p>
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
            <Link href="/dashboard" className="flex items-center space-x-3">
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
                <p className="text-xs text-gray-600">Provider Marketplace</p>
              </div>
            </Link>

            <div className="flex items-center space-x-4">
              <Link href="/dashboard">
                <Button variant="outline">
                  ← Back to Dashboard
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Page Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">GPU Provider Marketplace</h1>
          <p className="text-gray-600">
            Browse and connect to available GPU providers for your computing needs.
          </p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700">{error}</p>
            <Button variant="outline" onClick={loadProviders} className="mt-2">
              Retry
            </Button>
          </div>
        )}

        {/* Filters */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Filter className="w-5 h-5" />
              <span>Filters</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                <Input
                  placeholder="Search providers..."
                  value={filters.search}
                  onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                  className="pl-10"
                />
              </div>
              
              <select 
                value={filters.location} 
                onChange={(e) => setFilters({ ...filters, location: e.target.value })}
                className="border border-gray-300 rounded-md px-3 py-2 bg-white text-sm"
              >
                <option value="">All locations</option>
                {getAllLocations().map(location => (
                  <option key={location} value={location}>{location}</option>
                ))}
              </select>

              <select 
                value={filters.status} 
                onChange={(e) => setFilters({ ...filters, status: e.target.value })}
                className="border border-gray-300 rounded-md px-3 py-2 bg-white text-sm"
              >
                <option value="">All statuses</option>
                <option value="online">Online</option>
                <option value="busy">Busy</option>
                <option value="offline">Offline</option>
                <option value="maintenance">Maintenance</option>
              </select>

              <select 
                value={filters.gpuModel} 
                onChange={(e) => setFilters({ ...filters, gpuModel: e.target.value })}
                className="border border-gray-300 rounded-md px-3 py-2 bg-white text-sm"
              >
                <option value="">All GPU models</option>
                {getAllGPUModels().map(model => (
                  <option key={model} value={model}>{model}</option>
                ))}
              </select>

              <select 
                value={filters.sortBy} 
                onChange={(e) => setFilters({ ...filters, sortBy: e.target.value })}
                className="border border-gray-300 rounded-md px-3 py-2 bg-white text-sm"
              >
                <option value="name">Sort by Name</option>
                <option value="rating">Sort by Rating</option>
                <option value="price">Sort by Price</option>
                <option value="location">Sort by Location</option>
              </select>
            </div>
          </CardContent>
        </Card>

        {/* Provider Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {filteredProviders.map((provider) => (
            <Card key={provider.id} className="hover:shadow-lg transition-shadow duration-200">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">{provider.name}</CardTitle>
                  <Badge className={`${getStatusColor(provider.status)} flex items-center space-x-1`}>
                    {getStatusIcon(provider.status)}
                    <span className="capitalize">{provider.status}</span>
                  </Badge>
                </div>
                <CardDescription className="flex items-center space-x-2">
                  <MapPin className="w-4 h-4" />
                  <span>{provider.location || 'Unknown location'}</span>
                </CardDescription>
              </CardHeader>
              
              <CardContent className="space-y-4">
                {/* Provider Stats */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="flex items-center space-x-2">
                    <TrendingUp className="w-4 h-4 text-blue-500" />
                    <span>Rating: {provider.metadata?.rating || 'N/A'}</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Zap className="w-4 h-4 text-green-500" />
                    <span>{provider.metadata?.hourly_rate || 'N/A'} dGPU/hr</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <BarChart3 className="w-4 h-4 text-red-800" />
                    <span>{provider.metadata?.total_jobs || 0} jobs</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Wifi className="w-4 h-4 text-orange-500" />
                    <span>{provider.metadata?.uptime || 0}% uptime</span>
                  </div>
                </div>

                {/* GPUs */}
                <div>
                  <h4 className="font-medium text-sm text-gray-700 mb-2">Available GPUs:</h4>
                  <div className="space-y-2">
                    {provider.gpus.map((gpu, index) => (
                      <div key={index} className="bg-gray-50 rounded-lg p-3">
                        <div className="flex items-center justify-between mb-1">
                          <span className="font-medium text-sm">{gpu.model_name}</span>
                          <div className="flex items-center space-x-1">
                            {gpu.is_healthy ? (
                              <CheckCircle className="w-4 h-4 text-green-500" />
                            ) : (
                              <AlertCircle className="w-4 h-4 text-red-500" />
                            )}
                          </div>
                        </div>
                        
                        <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
                          <div>VRAM: {Math.round(gpu.vram_mb / 1024)} GB</div>
                          <div>Driver: {gpu.driver_version}</div>
                          {gpu.utilization_gpu_percent !== undefined && (
                            <>
                              <div>GPU: {gpu.utilization_gpu_percent}%</div>
                              <div>VRAM: {gpu.utilization_memory_percent}%</div>
                            </>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Last Seen */}
                <div className="flex items-center justify-between text-sm text-gray-500">
                  <span>Last seen: {formatLastSeen(provider.last_seen_at)}</span>
                  <span>Jobs: {provider.gpus.length} GPU{provider.gpus.length !== 1 ? 's' : ''}</span>
                </div>

                {/* Action Buttons */}
                <div className="grid grid-cols-2 gap-2">
                  <Button 
                    variant="outline" 
                    size="sm"
                    disabled={provider.status !== 'online'}
                  >
                    View Details
                  </Button>
                  <Button 
                    size="sm"
                    disabled={provider.status !== 'online'}
                  >
                    Rent GPU
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* No Results */}
        {filteredProviders.length === 0 && (
          <div className="text-center py-12">
            <Cpu className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-medium text-gray-900 mb-2">No providers found</h3>
            <p className="text-gray-600 mb-4">
              Try adjusting your filters or check back later for new providers.
            </p>
            <Button onClick={() => setFilters({ search: '', location: '', status: '', gpuModel: '', sortBy: 'name' })}>
              Clear Filters
            </Button>
          </div>
        )}
      </div>
    </div>
  );
} 