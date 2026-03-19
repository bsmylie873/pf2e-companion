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
