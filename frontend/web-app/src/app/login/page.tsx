'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import Image from 'next/image';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useAuth } from '@/hooks/useAuth';
import { 
  Eye, 
  EyeOff, 
  Shield, 
  Zap, 
  Globe, 
  Users,
  TrendingUp,
  DollarSign,
  ArrowRight,
  Sparkles,
  Lock,
  Mail,
  CheckCircle,
  Star
} from 'lucide-react';

export default function LoginPage() {
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();
  const router = useRouter();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    if (error) setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const success = await login(formData);
      if (success) {
        router.push('/dashboard');
      } else {
        setError('Invalid username or password. Please try again.');
      }
    } catch (error) {
      console.error('Login error:', error);
      setError('An error occurred during login. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-professional flex">
      {/* Left Panel - Login Form */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-md space-y-8">
          {/* Logo and Title */}
          <div className="text-center">
            <div className="flex items-center justify-center space-x-3 mb-6">
              <div className="w-14 h-14 logo-container flex items-center justify-center">
                <Image 
                  src="/dantegpu-logo.png" 
                  alt="DanteGPU Logo" 
                  width={32} 
                  height={32}
                  className="w-8 h-8"
                />
              </div>
              <div className="text-left">
                <h1 className="text-3xl font-bold text-gray-900">DanteGPU</h1>
                <p className="text-sm text-gray-600">Decentralized GPU Platform</p>
              </div>
            </div>
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              Welcome Back
            </h2>
            <p className="text-gray-600">
              Sign in to access your GPU computing dashboard
            </p>
          </div>

          {/* Login Form */}
          <Card className="card-professional">
            <CardHeader className="space-y-1">
              <CardTitle className="text-xl text-center text-gray-900">Sign In</CardTitle>
              <CardDescription className="text-center text-gray-600">
                Enter your credentials to continue
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="username" className="text-sm font-semibold text-gray-700">
                    Username or Email
                  </Label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      id="username"
                      name="username"
                      type="text"
                      placeholder="Enter your username or email"
                      value={formData.username}
                      onChange={handleInputChange}
                      className="pl-10 form-input"
                      required
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password" className="text-sm font-semibold text-gray-700">
                    Password
                  </Label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      id="password"
                      name="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="Enter your password"
                      value={formData.password}
                      onChange={handleInputChange}
                      className="pl-10 pr-10 form-input"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 hover:text-gray-600"
                    >
                      {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </div>

                {error && (
                  <Alert className="bg-red-50 border-2 border-red-500">
                    <AlertDescription className="text-red-700">
                      {error}
                    </AlertDescription>
                  </Alert>
                )}

                <Button 
                  type="submit" 
                  className="w-full bg-primary hover:bg-primary/90 text-white border-2 border-black btn-hover-professional" 
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <div className="flex items-center space-x-2">
                      <div className="loading-spinner w-4 h-4"></div>
                      <span>Signing In...</span>
                    </div>
                  ) : (
                    <div className="flex items-center space-x-2">
                      <span>Sign In</span>
                      <ArrowRight className="w-4 h-4" />
                    </div>
                  )}
                </Button>
              </form>
            </CardContent>
            <CardFooter className="flex flex-col space-y-4">
              <div className="text-center text-sm text-gray-600">
                <Link href="#" className="text-primary hover:text-primary/80 font-medium border-b border-transparent hover:border-primary">
                  Forgot your password?
                </Link>
              </div>
              <div className="text-center text-sm text-gray-600">
                Don't have an account?{' '}
                <Link href="/register" className="text-primary hover:text-primary/80 font-medium border-b border-transparent hover:border-primary">
                  Sign up
                </Link>
              </div>
            </CardFooter>
          </Card>

          {/* Demo Credentials */}
          <Card className="border-2 border-black bg-secondary/50 shadow-professional">
            <CardContent className="pt-6">
              <div className="text-center">
                <div className="flex items-center justify-center space-x-2 mb-3">
                  <Sparkles className="w-5 h-5 text-primary" />
                  <h3 className="text-lg font-semibold text-gray-900">Demo Access</h3>
                </div>
                <p className="text-sm text-gray-700 mb-4">
                  Try DanteGPU with demo credentials
                </p>
                <div className="bg-card border-2 border-black rounded-lg p-4 space-y-2">
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-gray-700">Username:</span>
                    <code className="bg-secondary border border-black px-2 py-1 rounded text-sm text-gray-800">demo</code>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm font-medium text-gray-700">Password:</span>
                    <code className="bg-secondary border border-black px-2 py-1 rounded text-sm text-gray-800">demo123456</code>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Back to Home */}
          <div className="text-center">
            <Link 
              href="/"
              className="text-gray-600 hover:text-gray-900 text-sm font-medium inline-flex items-center space-x-1 border-b border-transparent hover:border-gray-600"
            >
              <span>← Back to Home</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Right Panel - Features Showcase */}
      <div className="hidden lg:flex flex-1 bg-gradient-to-br from-primary via-primary/90 to-primary/80 text-white p-8 items-center justify-center relative overflow-hidden border-l-4 border-black">
        <div className="absolute inset-0 bg-gradient-to-br from-primary/95 via-primary/90 to-primary/85"></div>
        <div className="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg%20width%3D%2260%22%20height%3D%2260%22%20viewBox%3D%220%200%2060%2060%22%20xmlns%3D%22http%3A//www.w3.org/2000/svg%22%3E%3Cg%20fill%3D%22none%22%20fill-rule%3D%22evenodd%22%3E%3Cg%20fill%3D%22%23ffffff%22%20fill-opacity%3D%220.1%22%3E%3Cpath%20d%3D%22M36%2034v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6%2034v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6%204V0H4v4H0v2h4v4h2V6h4V4H6z%22/%3E%3C/g%3E%3C/g%3E%3C/svg%3E')] opacity-20"></div>
        
        <div className="relative z-10 max-w-lg">
          <div className="mb-8">
            <h2 className="text-4xl font-bold mb-4">
              The Future of <span className="text-yellow-300">GPU Computing</span>
            </h2>
            <p className="text-xl text-red-100 mb-6">
              Join thousands of developers, researchers, and creators using DanteGPU for their computing needs.
            </p>
          </div>

          {/* Features List */}
          <div className="space-y-6">
            <div className="flex items-start space-x-4">
              <div className="flex-shrink-0 w-12 h-12 bg-white/20 border-2 border-white/30 rounded-lg flex items-center justify-center">
                <Shield className="w-6 h-6" />
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-1">Blockchain Secured</h3>
                <p className="text-red-100">
                  All transactions secured by Solana blockchain with dGPU token payments
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-4">
              <div className="flex-shrink-0 w-12 h-12 bg-white/20 border-2 border-white/30 rounded-lg flex items-center justify-center">
                <Zap className="w-6 h-6" />
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-1">Lightning Fast</h3>
                <p className="text-red-100">
                  Access high-performance GPUs in seconds, not hours
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-4">
              <div className="flex-shrink-0 w-12 h-12 bg-white/20 border-2 border-white/30 rounded-lg flex items-center justify-center">
                <Globe className="w-6 h-6" />
              </div>
              <div>
                <h3 className="text-lg font-semibold mb-1">Global Network</h3>
                <p className="text-red-100">
                  Choose from providers worldwide for optimal performance
                </p>
              </div>
            </div>
          </div>

          {/* Stats */}
          <div className="mt-12 grid grid-cols-3 gap-4 pt-8 border-t-2 border-white/20">
            <div className="text-center p-3 bg-white/10 border border-white/20 rounded-lg">
              <div className="text-2xl font-bold text-yellow-300">1000+</div>
              <div className="text-sm text-red-100">Available GPUs</div>
            </div>
            <div className="text-center p-3 bg-white/10 border border-white/20 rounded-lg">
              <div className="text-2xl font-bold text-yellow-300">99.9%</div>
              <div className="text-sm text-red-100">Uptime</div>
            </div>
            <div className="text-center p-3 bg-white/10 border border-white/20 rounded-lg">
              <div className="text-2xl font-bold text-yellow-300">24/7</div>
              <div className="text-sm text-red-100">Support</div>
            </div>
          </div>

          {/* Testimonial */}
          <div className="mt-8 bg-white/10 border-2 border-white/20 rounded-lg p-6 backdrop-blur-sm shadow-professional">
            <div className="flex items-center space-x-1 mb-3">
              {[...Array(5)].map((_, i) => (
                <Star key={i} className="w-4 h-4 fill-yellow-300 text-yellow-300" />
              ))}
            </div>
            <p className="text-red-100 mb-4">
              "DanteGPU revolutionized our AI training workflow. The decentralized approach 
              gives us access to cutting-edge hardware at competitive prices."
            </p>
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-yellow-300 to-orange-400 border-2 border-white rounded-full flex items-center justify-center">
                <span className="text-sm font-bold text-gray-900">AI</span>
              </div>
              <div>
                <div className="font-semibold text-white">Alex Chen</div>
                <div className="text-sm text-red-200">AI Researcher, TechCorp</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 