# PolyID

A cloud-native identity platform providing Passkey (FIDO2/WebAuthn) authentication and multi-factor authentication (MFA).

## Features

- Passkey Authentication (FIDO2/WebAuthn)
- Multi-Factor Authentication (App-link, TOTP, SMS)
- High Performance (<50ms P99 latency at 100k RPM)
- Global Scale (Multi-region active/active)
- Security First (Hardware-backed key attestation)
- Privacy Focused (Privacy budget enforcement)

## Architecture

- Backend: Go microservices on Kubernetes
- Data Storage: Distributed NoSQL database
- Caching: Redis edge cache
- Event Processing: Kafka + gRPC
- Authentication: OIDC & OAuth 2.0
- Device Registry: X.509 attestation

## Quick Start

```bash
# Clone the repository
git clone https://github.com/yourusername/polyid-auth.git

# Install dependencies
go mod download

# Run the development environment
make dev
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- Kubernetes cluster (for production)
- Cloud provider account (for managed services)

### Local Development

1. Install dependencies:
   ```bash
   make deps
   ```

2. Start local services:
   ```bash
   make local
   ```

3. Run tests:
   ```bash
   make test
   ```

## Security

- Credentials encrypted at rest using cloud KMS
- Hardware-backed key attestation for passkeys
- Automated threat intelligence integration
- Privacy budget enforcement for device fingerprinting

## License

MIT License - see LICENSE file for details