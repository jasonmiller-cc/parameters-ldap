# parameters-ldap

REST API service for LDAP directory management (FreeIPA / OpenLDAP), built on
[parameters-core](https://github.com/jasonmiller-cc/parameters-core).

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users` | List users (`?search=`, `?limit=`, `?offset=`) |
| GET | `/api/v1/users/{uid}` | Get user by uid |
| POST | `/api/v1/users` | Create user |
| PUT | `/api/v1/users/{uid}` | Update user attributes |
| DELETE | `/api/v1/users/{uid}` | Delete user |
| GET | `/api/v1/groups` | List groups |
| GET | `/api/v1/groups/{cn}` | Get group by cn |
| POST | `/api/v1/groups` | Create group |
| POST | `/api/v1/groups/{cn}/members` | Add member to group |
| DELETE | `/api/v1/groups/{cn}/members/{uid}` | Remove member from group |
| POST | `/api/v1/auth/bind` | Test LDAP bind / authenticate |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe (LDAP ping) |
| GET | `/metrics` | Prometheus metrics |

## Configuration

```yaml
server:
  host: 0.0.0.0
  port: 8080

ldap:
  url: ldaps://ipa-01.parameters.cc
  bind_dn: uid=svc-ldap,cn=sysaccounts,cn=etc,dc=parameters,dc=cc
  bind_password: ""        # set via PARAMS_LDAP_BIND_PASSWORD env var
  base_dn: dc=parameters,dc=cc
  user_base_dn: cn=users,cn=accounts,dc=parameters,dc=cc
  group_base_dn: cn=groups,cn=accounts,dc=parameters,dc=cc
  tls_skip_verify: false

log:
  level: info
  format: json

metrics:
  enabled: true
  path: /metrics
```

Environment variable overrides use the `PARAMS_LDAP_` prefix (e.g.
`PARAMS_LDAP_CONFIG=/etc/parameters/ldap/config.yaml`).

## Development

```bash
# Build
make build

# Run tests
make test

# Run locally (requires a reachable LDAP server or stubs)
make run

# Docker image
make docker-build
```

## Module layout

```
cmd/server/         entry point
internal/
  config/           Config embedding BaseConfig + LDAPConfig
  api/              HTTP handlers (net/http ServeMux)
  service/          LDAP operations (go-ldap/ldap/v3 stubs)
```
