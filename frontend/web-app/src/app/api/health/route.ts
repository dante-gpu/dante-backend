import { NextResponse } from 'next/server';

export async function GET() {
  try {
    const healthData = {
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'Dante GPU Rental Platform - Frontend',
      version: '1.0.0',
      environment: process.env.NEXT_PUBLIC_ENVIRONMENT || 'development',
      api_url: process.env.NEXT_PUBLIC_API_URL,
      checks: {
        memory: process.memoryUsage(),
        uptime: process.uptime(),
      }
    };

    return NextResponse.json(healthData, { status: 200 });
  } catch (error) {
    return NextResponse.json(
      {
        status: 'unhealthy',
        timestamp: new Date().toISOString(),
        error: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 500 }
    );
  }
} 