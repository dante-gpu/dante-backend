# Billing & Payment Service (Dante Backend)

This Go service handles all financial transactions for the Dante GPU Platform, including dGPU token integration via Solana blockchain, usage tracking, pricing calculations, and provider payouts.

## Responsibilities

- **dGPU Token Integration**: Manage transactions using the dGPU token (7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump) on Solana blockchain
- **Usage Tracking**: Monitor GPU rental time, VRAM allocation, and power consumption
- **Dynamic Pricing**: Calculate hourly rates based on GPU specifications, power consumption, and market demand
- **Wallet Management**: Handle user and provider dGPU token wallets
- **Payment Processing**: Process rental payments and provider payouts in dGPU tokens
- **Billing History**: Maintain transaction records and usage reports
- **VRAM Allocation Pricing**: Calculate costs based on allocated VRAM portions

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL (for transaction records and usage tracking)
- **Blockchain**: Solana (for dGPU token transactions)
- **Message Queue**: NATS (for real-time usage updates)
- **Service Discovery**: Consul
- **Monitoring**: Prometheus metrics

## Key Features

### dGPU Token Integration
- Real-time token balance checking
- Secure transaction processing
- Multi-signature wallet support
- Transaction fee optimization

### Dynamic Pricing Engine
- Base rates by GPU model and VRAM
- Power consumption multipliers
- Market demand adjustments
- Provider-set minimum rates

### VRAM Allocation Management
- Fractional VRAM rental (e.g., 50% of 24GB = 12GB)
- Per-GB pricing calculations
- Real-time allocation tracking
- Automatic billing adjustments

### Usage Monitoring
- Real-time GPU utilization tracking
- Precise billing down to the minute
- Power consumption monitoring
- Automatic session termination on insufficient funds

## API Endpoints

### User Wallet Management
- `GET /api/v1/wallet/balance` - Get user's dGPU token balance
- `POST /api/v1/wallet/deposit` - Initiate token deposit
- `POST /api/v1/wallet/withdraw` - Request token withdrawal
- `GET /api/v1/wallet/transactions` - Get transaction history

### Billing & Usage
- `POST /api/v1/billing/start-session` - Start GPU rental session
- `POST /api/v1/billing/end-session` - End GPU rental session
- `GET /api/v1/billing/current-usage` - Get current session costs
- `GET /api/v1/billing/history` - Get billing history

### Provider Payouts
- `GET /api/v1/provider/earnings` - Get provider earnings
- `POST /api/v1/provider/payout` - Request payout
- `GET /api/v1/provider/rates` - Get/set provider rates

### Pricing
- `GET /api/v1/pricing/rates` - Get current GPU rental rates
- `POST /api/v1/pricing/calculate` - Calculate cost for specific requirements

## Configuration

```yaml
# Solana Configuration
solana:
  rpc_url: "https://api.mainnet-beta.solana.com"
  dgpu_token_address: "7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump"
  platform_wallet: "YOUR_PLATFORM_WALLET_ADDRESS"
  private_key_path: "/secrets/solana_private_key"

# Database Configuration
database:
  url: "postgres://user:password@localhost:5432/dante_billing"
  max_connections: 25
  connection_timeout: "30s"

# Pricing Configuration
pricing:
  base_rates:
    nvidia_rtx_4090: 0.50  # dGPU tokens per hour
    nvidia_a100: 2.00
    nvidia_h100: 4.00
  power_multiplier: 0.001  # Additional cost per watt
  platform_fee_percent: 5.0
  minimum_session_minutes: 1

# NATS Configuration
nats:
  address: "nats://localhost:4222"
  usage_updates_subject: "dante.billing.usage"
  payment_events_subject: "dante.billing.payments"

# Service Configuration
server:
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

# Consul Configuration
consul:
  address: "localhost:8500"
  service_name: "billing-payment-service"
  health_check_interval: "10s"
```

## Database Schema

### Tables
- `wallets` - User and provider dGPU token wallets
- `rental_sessions` - Active and completed GPU rental sessions
- `transactions` - All dGPU token transactions
- `usage_records` - Detailed usage tracking
- `provider_rates` - Custom provider pricing
- `billing_history` - Aggregated billing records

## Security Considerations

- Private keys stored in secure key management system
- Multi-signature transactions for large amounts
- Rate limiting on API endpoints
- Audit logging for all financial operations
- Real-time fraud detection

## Integration Points

- **Provider Registry Service**: Get GPU specifications and availability
- **Scheduler Service**: Receive job start/end notifications
- **Auth Service**: Validate user permissions
- **Monitoring Service**: Track service health and performance
- **Solana Blockchain**: Execute dGPU token transactions

## Development Status

- [x] Service structure and configuration
- [x] Database schema design
- [x] Solana integration planning
- [ ] Core billing logic implementation
- [ ] dGPU token transaction handling
- [ ] Dynamic pricing engine
- [ ] VRAM allocation pricing
- [ ] Provider payout system
- [ ] API endpoint implementation
- [ ] Integration testing
- [ ] Security audit
