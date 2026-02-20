# Cierge Configuration

All configuration values across the server and the CLI can be overridden with environment variables that have the `CIERGE_` prefix (e.g. `CIERGE_DATABASE_HOST`).

## Command Line Interface

| Field | Description |
| ------------- | -------------- |
| `host_url` | URL of the Cierge server |
| `api_key` | User's API key |


## Server

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `environment` | `dev` | Environment of the server (`dev` or `prod`) |
| `log_level` | `info` | Logging level |
| `server` |  | Host and TLS configuration |
| `database` |  | Database configuration |
| `token_store_path` | `./data/token_store` | Path for the BadgerDB token store key,value database |
| `auth` |  | Authentication configuration |
| `cloud` |  | Cloud providers |
| `notification` |  | Notification providers |
| `default_admin` |  | Credentials of the default administrator (used if no user exists) |


### Server

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `host` | `localhost` | Hostname of the server |
| `external_host` | `` | External hostname of the server |
| `port` | `8080` | Port number to run the server |
| `tls.enabled` | `false` | Whether TLS should be enabled |
| `tls.cert_file` | `` | Certificate file for TLS |
| `tls.key_file` | `` | Key file for TLS |
| `trusted_proxies` | `` | List of trusted proxy IP addresses |
| `cors_origins` | `` | List of allowed CORS origins |


### Database

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `host` | `localhost` | Database hostname |
| `port` | `5432` | Database port |
| `user` | `cierge` | Database user |
| `password` | `` | Database password |
| `database` | `cierge` | Database name |
| `ssl_mode` | `disable` | SSL mode |
| `auto_migrate` | `true` | Automatically run migrations on startup |
| `timeout` | `30s` | Database connection timeout |


### Auth

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `method` | `local` | Authentication method (`local` or `oidc`) |
| `jwt_secret` | `` | Secret key for signing JWTs (minimum 64 characters recommended for production) |
| `jwt_issuer` | `cierge` | Issuer field for JWT tokens |
| `access_token_expiry` | `15m` | Access token expiry duration |
| `refresh_token_expiry` | `168h` | Refresh token expiry duration |
| `rate_limit_requests` | `3` | Maximum login attempts per window (local auth only) |
| `rate_limit_window` | `5m` | Rate limit window duration (local auth only) |
| `oidc_providers` | `{}` | Map of OIDC provider configurations keyed by provider name |

#### OIDC Provider

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `client_id` | `` | OIDC client ID |
| `client_secret` | `` | OIDC client secret |
| `issuer_url` | `` | OIDC issuer URL |
| `redirect_url` | `` | OIDC redirect URL |
| `scopes` | `` | List of OIDC scopes to request |
| `backchannel_logout` | `false` | Whether to enable OIDC backchannel logout |


### Cloud

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `provider` | `aws` | Cloud provider (`local` or `aws`) |
| `config` | | Provider-specific configuration |

#### AWS Config

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `region` | `` | AWS region |
| `kms_key_id` | `` | KMS key ID for encryption |
| `lambda_arn` | `` | ARN of the Lambda function for job execution |
| `scheduler_role_arn` | `` | ARN of the IAM role for the EventBridge scheduler |
| `schedule_group_name` | `` | EventBridge Scheduler schedule group name |
| `cold_start_buffer` | `1m` | Buffer duration to account for Lambda cold starts |
| `access_key_id` | `` | AWS access key ID |
| `secret_access_key` | `` | AWS secret access key |


### Notification

Notification is a list of notification provider configurations.

| Field | Default | Description |
| --------------- | --------------- | --------------- |
| `name` | `` | Notification provider name (`email`, `sms`, or `webhook`) |
| `enabled` | `false` | Whether the notification provider is enabled |
| `config` | `{}` | Provider-specific configuration |


### Default Admin

| Field | Description |
| -------------- | --------------- |
| `email` | Email of the default administrator |
| `password` | Password of the default administrator |
