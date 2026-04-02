# pf2e-companion

## Local Development

### Quick Start

```bash
docker compose up        # starts db → runs migrations → seeds data
```

### Clean Reset

```bash
docker compose down -v && docker compose up
```

This tears down all containers and volumes (destroying DB data), then re-creates everything from scratch.

### Seed User Credentials

All seeded users share the password: **`password123`**

| Username         | Email                | Role   |
|------------------|----------------------|--------|
| `gm_valen`       | `valen@example.com`  | GM     |
| `player_elara`   | `elara@example.com`  | Player |
| `player_dorn`    | `dorn@example.com`   | Player |

---

## Production Deployment

### Architecture

- **Backend**: Go Echo API on [Fly.io](https://fly.io) (shared-cpu-1x, 512 MB RAM)
- **Database**: Fly Postgres (shared-cpu-1x, 256 MB RAM, 1 GB volume)
- **Frontend**: React SPA on [Cloudflare Pages](https://pages.cloudflare.com)
- **File uploads**: Fly Volume mounted at `/app/uploads/maps`

### One-Time Provisioning

```bash
# 1. Authenticate
fly auth login

# 2. Create application
fly apps create pf2e-companion

# 3. Create Postgres cluster
fly postgres create --name pf2e-companion-db --region iad \
  --vm-size shared-cpu-1x --volume-size 1 --initial-cluster-size 1

# 4. Attach Postgres to the app
fly postgres attach pf2e-companion-db -a pf2e-companion

# 5. Create persistent volume for map uploads
fly volumes create uploads_vol --region iad --size 1 -a pf2e-companion

# 6. Set application secrets
fly secrets set \
  JWT_SECRET=$(openssl rand -hex 32) \
  CORS_ALLOW_ORIGIN=https://pf2e-companion.pages.dev \
  -a pf2e-companion

# 7. (Optional) Allocate dedicated IPv4 (+$2/mo)
fly ips allocate-v4 -a pf2e-companion
```

### Cloudflare Pages Setup

1. Connect the repository in the Cloudflare dashboard
2. Set root directory: `ui/`
3. Build command: `npm run build`
4. Output directory: `dist`
5. Add environment variable: `VITE_API_BASE_URL=https://pf2e-companion.fly.dev`

### GitHub Configuration

**Secrets** (Settings → Secrets → Actions):

| Secret | Purpose |
|---|---|
| `FLY_API_TOKEN` | Fly.io deploy token (`fly tokens create deploy`) |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token with Pages edit permission |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID |
| `POSTGRES_USER` | Database user for Flyway migrations |
| `POSTGRES_PASSWORD` | Database password for Flyway migrations |
| `POSTGRES_DB` | Database name for Flyway migrations |

**Variables** (Settings → Variables → Actions):

| Variable | Purpose |
|---|---|
| `VITE_API_BASE_URL` | Backend URL (e.g., `https://pf2e-companion.fly.dev`) |

### CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/deploy.yml`) triggers on every push to `master`:

1. **deploy-backend** — Builds the Docker image and deploys to Fly.io
2. **migrate-database** — Runs Flyway migrations (only after backend deploys successfully)
3. **deploy-frontend** — Builds the React SPA and deploys to Cloudflare Pages (runs in parallel with backend)

Migrations only run after the backend health check passes, protecting against partially-migrated database states.

### Cost Estimate

| Component | Monthly Cost |
|---|---|
| Fly Machine (shared-cpu-1x, 512 MB) | ~$3.32 |
| Fly Volume (1 GB) | ~$0.15 |
| Fly Postgres (shared-cpu-1x, 256 MB, 1 GB vol) | ~$2.17 |
| Dedicated IPv4 (optional) | $2.00 |
| Cloudflare Pages | $0.00 |
| **Total** | **~$7.64** |
