# mTLS Configuration

## Overview

The NamespaceManager service now supports mutual TLS (mTLS) authentication in both directions:

1. **Incoming requests**: Accepts and verifies client certificates from glidelet and resource controllers
2. **Outgoing requests**: Presents its own client certificate when making requests to glidelet and resource controllers

## Configuration

### Environment Variables

Set the following environment variables to enable mTLS:

```bash
# Enable mTLS for incoming connections
MTLS_ENABLED=true

# Server certificate and key (also used as client certificate for outgoing requests)
TLS_CERT_PATH=/path/to/server-cert.pem
TLS_KEY_PATH=/path/to/server-key.pem

# CA certificate for verifying peer certificates (both incoming and outgoing)
MTLS_CA_CERT_PATH=/path/to/ca-cert.pem
```

### Certificate Requirements

1. **Server/Client Certificate** (`TLS_CERT_PATH`, `TLS_KEY_PATH`):
   - This certificate is used both as the server certificate for HTTPS and as the client certificate when making outgoing requests
   - Must be signed by the CA that glidelet and resource controllers trust
   - Should include appropriate Common Name (CN) and Subject Alternative Names (SANs)

2. **CA Certificate** (`MTLS_CA_CERT_PATH`):
   - Used to verify incoming client certificates from glidelet/resource controller
   - Used to verify server certificates when connecting to glidelet/resource controller
   - Should be the root or intermediate CA that signed all service certificates

## How It Works

### Incoming Requests (Server Side)

When `MTLS_ENABLED=true`:
1. Server starts with TLS configuration that requests client certificates
2. The `MTLSMiddleware()` verifies that a valid client certificate was provided
3. Routes protected with `middleware.MTLSMiddleware()` require a client certificate
4. The client's Common Name (CN) is extracted and stored in the Gin context

Example from routers.go:
```go
ticketGroup.POST("/updateTicketStatusFromGlidelet", middleware.MTLSMiddleware(), ticketHandler.UpdateTicketStatusFromGlidelet())
```

### Outgoing Requests (Client Side)

The service automatically presents its client certificate when making requests to external services:

1. During application startup, `httpclient.InitMTLSClient()` is called
2. This function loads the client certificate from `TLS_CERT_PATH` and `TLS_KEY_PATH`
3. An HTTP client with TLS configuration is created and stored in `httpclient.MTLSClient`
4. All requests made via `httpclient.SendRequest()` and `utils.SendRequest()` use this shared mTLS-enabled client

**Note**: The `utils` package delegates to `httpclient.MTLSClient` to avoid duplication. There's only one mTLS client initialization.

## Affected Services

The following outgoing requests will present the client certificate:

- Requests to glidelet (resource pools):
  - `POST /api/v1/ticket/createPods` - Creating pods
  - `PATCH /api/v1/ticket/cancelPods` - Cancelling pods
  - `POST /api/v1/ticket/createList` - Creating ticket lists
  - `PATCH /api/v1/ticket/rollbackPods` - Rolling back pods
  - `GET /api/v1/pool/{pool_id}` - Getting pool information

- Any other external service calls using the `httpclient` or `utils` packages

## Testing

### With mTLS Enabled

```bash
# Set environment variables
export MTLS_ENABLED=true
export TLS_CERT_PATH=/path/to/cert.pem
export TLS_KEY_PATH=/path/to/key.pem
export MTLS_CA_CERT_PATH=/path/to/ca.pem

# Start the service
go run cmd/api/main.go
```

### Without mTLS (Development)

For local development without mTLS:
```bash
# Don't set MTLS_ENABLED or set it to false
export MTLS_ENABLED=false

# Start the service
go run cmd/api/main.go
```

The service will:
- Accept both HTTP requests (without client certificates)
- Make outgoing requests without presenting a client certificate
- Still use HTTPS if `TLS_CERT_PATH` and `TLS_KEY_PATH` are set

## Troubleshooting

### Client Certificate Not Being Sent

If outgoing requests fail with certificate errors:
1. Verify `TLS_CERT_PATH` and `TLS_KEY_PATH` are set correctly
2. Check that the certificate is valid and not expired: `openssl x509 -in cert.pem -text -noout`
3. Ensure the certificate is trusted by the receiving service

### Client Certificate Not Accepted by Remote Service

If the remote service rejects your certificate:
1. Verify the certificate was signed by a CA the remote service trusts
2. Check the certificate's Common Name (CN) matches expected values
3. Review the remote service's logs for rejection reasons

### Incoming Client Certificates Rejected

If incoming requests from glidelet/resource controller are rejected:
1. Verify `MTLS_CA_CERT_PATH` contains the correct CA certificate
2. Check that the client's certificate is signed by that CA
3. Review server logs for certificate validation errors

## Security Best Practices

1. **Protect Private Keys**: Ensure `.key` files are readable only by the application user
2. **Certificate Rotation**: Plan for regular certificate renewal before expiration
3. **CA Trust**: Only add necessary CAs to `MTLS_CA_CERT_PATH`
4. **Monitoring**: Monitor certificate expiration dates
5. **Logs**: Review connection logs regularly for authentication failures
