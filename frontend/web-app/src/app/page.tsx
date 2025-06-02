'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useRouter } from 'next/navigation';
import { 
  Zap, 
  Shield, 
  Globe, 
  DollarSign, 
  TrendingUp, 
  Users,
  Star,
  MapPin,
  Monitor,
  Cpu,
  Activity,
  ArrowRight,
  CheckCircle,
  Sparkles
} from 'lucide-react';

interface Provider {
  id: string;
  name: string;
  location: string;
  gpu_model: string;
  vram_gb: number;
  hourly_rate: number;
  rating: number;
  available: boolean;
  gpus: GPUDetail[];
}

interface GPUDetail {
  modelName: string;
  vramMb: number;
  driverVersion: string;
  isHealthy: boolean;
}

export default function HomePage() {
  const { user, isAuthenticated } = useAuth();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    loadMarketplaceData();
  }, [isAuthenticated]);

  const loadMarketplaceData = async () => {
    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setProviders([
        {
          id: '1',
          name: 'CloudGPU Pro',
          location: 'US-East',
          gpu_model: 'NVIDIA RTX 4090',
          vram_gb: 24,
          hourly_rate: 0.75,
          rating: 4.9,
          available: true,
          gpus: [
            {
              modelName: 'NVIDIA RTX 4090',
              vramMb: 24576,
              driverVersion: '536.23',
              isHealthy: true
            }
          ]
        },
        {
          id: '2', 
          name: 'AI Compute Hub',
          location: 'EU-West',
          gpu_model: 'NVIDIA A100',
          vram_gb: 80,
          hourly_rate: 1.50,
          rating: 4.8,
          available: true,
          gpus: [
            {
              modelName: 'NVIDIA A100',
              vramMb: 81920,
              driverVersion: '535.86',
              isHealthy: true
            }
          ]
        },
        {
          id: '3',
          name: 'RenderFarm Elite',
          location: 'Asia-Pacific',
          gpu_model: 'NVIDIA RTX 3080',
          vram_gb: 10,
          hourly_rate: 0.45,
          rating: 4.7,
          available: true,
          gpus: [
            {
              modelName: 'NVIDIA RTX 3080',
              vramMb: 10240,
              driverVersion: '535.98',
              isHealthy: true
            }
          ]
        }
      ]);
    } catch (error) {
      console.error('Failed to load marketplace data:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-professional">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 nav-professional">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 logo-container flex items-center justify-center">
                <Image
                  src="/dantegpu-logo.png"
                  alt="DanteGPU Logo"
                  width={24}
                  height={24}
                  className="w-6 h-6"
                />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">DanteGPU</h1>
                <p className="text-xs text-gray-600">GPU Rental Platform</p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {isAuthenticated ? (
                <div className="flex items-center space-x-4">
                  <span className="text-sm text-gray-700">Welcome, {user?.username}!</span>
                  <Button 
                    onClick={() => router.push('/dashboard')}
                    className="bg-primary hover:bg-primary/90 text-white border-2 border-black btn-hover-professional"
                  >
                    Dashboard
                  </Button>
                </div>
              ) : (
                <>
                  <Link 
                    href="/login"
                    className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md transition-colors border border-black hover:bg-secondary"
                  >
                    Sign In
                  </Link>
                  <Link 
                    href="/register"
                    className="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md transition-colors border border-black hover:bg-secondary"
                  >
                    Sign Up
                  </Link>
                  <Button 
                    onClick={() => router.push('/register')}
                    className="bg-primary hover:bg-primary/90 text-white border-2 border-black btn-hover-professional"
                  >
                    Get Started
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative py-16 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto text-center">
          <div className="inline-flex items-center px-4 py-2 rounded-full bg-secondary border-2 border-black text-gray-800 text-sm font-medium mb-6 shadow-professional">
            <Sparkles className="w-4 h-4 mr-2 text-primary" />
            Powered by Solana Blockchain & dGPU Tokens
          </div>
          
          <h1 className="text-4xl md:text-6xl font-bold text-gray-900 mb-6">
            Rent <span className="text-primary">High-Performance</span> GPUs
          </h1>
          
          <p className="text-lg md:text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
            Access powerful GPU computing resources on-demand. Perfect for AI training, machine learning, 
            cryptocurrency mining, and high-performance computing workloads.
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
            <Button 
              size="lg"
              onClick={() => router.push('/register')}
              className="bg-primary hover:bg-primary/90 text-white px-8 py-3 border-2 border-black btn-hover-professional"
            >
              Start Renting GPUs
              <ArrowRight className="ml-2 h-5 w-5" />
            </Button>
            <Button 
              variant="outline" 
              size="lg"
              className="border-2 border-black text-gray-900 hover:bg-secondary px-8 py-3 btn-hover-professional"
            >
              View Marketplace
            </Button>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-4xl mx-auto">
            <div className="text-center p-4 bg-card border-2 border-black rounded-lg shadow-professional">
              <div className="text-3xl font-bold text-primary">1000+</div>
              <div className="text-gray-600">Available GPUs</div>
            </div>
            <div className="text-center p-4 bg-card border-2 border-black rounded-lg shadow-professional">
              <div className="text-3xl font-bold text-primary">99.9%</div>
              <div className="text-gray-600">Uptime Guarantee</div>
            </div>
            <div className="text-center p-4 bg-card border-2 border-black rounded-lg shadow-professional">
              <div className="text-3xl font-bold text-primary">24/7</div>
              <div className="text-gray-600">Expert Support</div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 px-4 sm:px-6 lg:px-8 bg-card">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Why Choose DanteGPU?
            </h2>
            <p className="text-lg text-gray-600 max-w-3xl mx-auto">
              Experience the future of GPU computing with our decentralized platform
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[
              {
                icon: <Zap className="w-6 h-6" />,
                title: "Lightning Fast Performance",
                description: "Access the latest NVIDIA RTX 4090, A100, and H100 GPUs with maximum performance for your workloads."
              },
              {
                icon: <Shield className="w-6 h-6" />,
                title: "Blockchain Security",
                description: "Secure transactions powered by Solana blockchain with dGPU token payments for transparent billing."
              },
              {
                icon: <Globe className="w-6 h-6" />,
                title: "Global Network",
                description: "Choose from providers worldwide with low-latency access to computing resources near you."
              },
              {
                icon: <DollarSign className="w-6 h-6" />,
                title: "Cost Effective",
                description: "Pay only for what you use with competitive pricing and no long-term commitments required."
              },
              {
                icon: <TrendingUp className="w-6 h-6" />,
                title: "Auto Scaling",
                description: "Automatically scale your compute resources up or down based on demand and budget constraints."
              },
              {
                icon: <Users className="w-6 h-6" />,
                title: "24/7 Support",
                description: "Expert technical support available around the clock to help with your computing needs."
              }
            ].map((feature, index) => (
              <Card key={index} className="card-professional">
                <CardHeader>
                  <div className="flex items-center space-x-3">
                    <div className="p-2 bg-secondary border border-black rounded-lg text-primary">
                      {feature.icon}
                    </div>
                    <CardTitle className="text-lg">{feature.title}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-gray-600">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Use Cases Section */}
      <section className="py-16 px-4 sm:px-6 lg:px-8 bg-secondary">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Perfect for Every Use Case
            </h2>
            <p className="text-lg text-gray-600 max-w-3xl mx-auto">
              From AI research to cryptocurrency mining, our platform supports all your computing needs
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              {
                title: "AI & Machine Learning",
                description: "Train neural networks, run inference, and develop AI models with powerful GPU acceleration.",
                icon: <Monitor className="w-6 h-6" />,
                color: "bg-primary"
              },
              {
                title: "3D Rendering",
                description: "Render complex 3D scenes, animations, and visual effects with professional-grade GPUs.",
                icon: <Cpu className="w-6 h-6" />,
                color: "bg-primary"
              },
              {
                title: "Cryptocurrency Mining",
                description: "Mine cryptocurrencies efficiently with optimized GPU configurations and competitive rates.",
                icon: <DollarSign className="w-6 h-6" />,
                color: "bg-primary"
              },
              {
                title: "Scientific Computing",
                description: "Accelerate research computations, simulations, and data analysis with parallel processing.",
                icon: <Activity className="w-6 h-6" />,
                color: "bg-primary"
              }
            ].map((useCase, index) => (
              <Card key={index} className="card-professional">
                <CardHeader>
                  <div className={`w-10 h-10 ${useCase.color} border-2 border-black rounded-lg flex items-center justify-center text-white mb-3`}>
                    {useCase.icon}
                  </div>
                  <CardTitle className="text-lg">{useCase.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-gray-600">
                    {useCase.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Popular Providers Section */}
      <section className="py-16 px-4 sm:px-6 lg:px-8 bg-card">
        <div className="max-w-7xl mx-auto">
          <div className="flex justify-between items-center mb-12">
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-4">
                Popular GPU Providers
              </h2>
              <p className="text-lg text-gray-600">
                Top-rated providers with verified performance
              </p>
            </div>
            <Button variant="outline" className="hidden md:block border-2 border-black btn-hover-professional">
              View All Providers
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </div>
          
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[1, 2, 3].map((item) => (
                <Card key={item} className="animate-pulse card-professional">
                  <CardHeader>
                    <div className="h-6 bg-gray-300 rounded mb-2"></div>
                    <div className="h-4 bg-gray-200 rounded"></div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="h-4 bg-gray-200 rounded"></div>
                      <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {providers.map((provider) => (
                <Card key={provider.id} className="card-professional">
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div>
                        <CardTitle className="text-lg font-semibold text-gray-900">
                          {provider.name}
                        </CardTitle>
                        <div className="flex items-center space-x-2 mt-1">
                          <MapPin className="w-4 h-4 text-gray-500" />
                          <span className="text-sm text-gray-600">{provider.location}</span>
                        </div>
                      </div>
                      <Badge className={`badge-professional ${provider.available ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                        {provider.available ? "Available" : "Busy"}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">GPU Model</span>
                        <span className="font-medium text-gray-900">{provider.gpu_model}</span>
                      </div>
                      
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">VRAM</span>
                        <span className="font-medium text-gray-900">
                          {Math.round(provider.gpus.reduce((total, gpu) => total + gpu.vramMb, 0) / 1024)} GB
                        </span>
                      </div>
                      
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-600">Rating</span>
                        <div className="flex items-center space-x-1">
                          <Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
                          <span className="font-medium text-gray-900">{provider.rating}</span>
                        </div>
                      </div>
                      
                      <div className="flex justify-between items-center pt-3 border-t-2 border-black">
                        <span className="text-lg font-bold text-primary">
                          ${provider.hourly_rate}/hour
                        </span>
                        <Button size="sm" className="bg-primary hover:bg-primary/90 border border-black btn-hover-professional">
                          Rent Now
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12 border-t-4 border-black">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="col-span-1 md:col-span-2">
              <div className="flex items-center space-x-3 mb-4">
                <div className="w-8 h-8 logo-container flex items-center justify-center">
                  <Image
                    src="/dantegpu-logo.png"
                    alt="DanteGPU Logo"
                    width={20}
                    height={20}
                    className="w-5 h-5"
                  />
                </div>
                <h3 className="text-xl font-bold">DanteGPU</h3>
              </div>
              <p className="text-gray-400 mb-4 max-w-md">
                The world's first decentralized GPU rental platform powered by Solana blockchain. 
                Democratizing access to high-performance computing resources.
              </p>
              <div className="flex items-center space-x-2">
                <CheckCircle className="w-4 h-4 text-green-400" />
                <span className="text-gray-300 text-sm">Blockchain Secured</span>
              </div>
            </div>
            
            <div>
              <h4 className="text-lg font-semibold mb-4">Platform</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Marketplace</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Providers</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Pricing</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Documentation</Link></li>
              </ul>
            </div>
            
            <div>
              <h4 className="text-lg font-semibold mb-4">Support</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Help Center</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Contact Us</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Status</Link></li>
                <li><Link href="#" className="hover:text-white transition-colors border-b border-transparent hover:border-white">Community</Link></li>
              </ul>
            </div>
          </div>
          
          <div className="border-t border-gray-800 mt-8 pt-8 flex flex-col md:flex-row justify-between items-center">
            <p className="text-gray-400 text-sm">
              © 2024 DanteGPU. All rights reserved.
            </p>
            <div className="flex items-center space-x-6 mt-4 md:mt-0">
              <Link href="#" className="text-gray-400 hover:text-white transition-colors text-sm border-b border-transparent hover:border-white">Privacy Policy</Link>
              <Link href="#" className="text-gray-400 hover:text-white transition-colors text-sm border-b border-transparent hover:border-white">Terms of Service</Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
