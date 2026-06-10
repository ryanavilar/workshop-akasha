# Workshop Akasha

Hands-on workshop for Kubernetes infrastructure testing on the Akasha platform.

## What We're Testing

### 1. App Changes & GitOps Pipeline

Modify the Vikunja todo app, rebuild, and watch Flux roll it out.

**Example: Rename Vikunja to "Tokunja"**

```bash
# Edit apps/vikunja/Dockerfile to customize the app
vi apps/vikunja/Dockerfile

# Commit and push
git add . && git commit -m "rebrand vikunja to tokunja" && git push
```

The pipeline:

```
Code change → git push → GitHub Runner → buildah build → Zot registry → Flux CD → Deployment rollout
```

### 2. Database IOPS (pg_bench)

Benchmark PostgreSQL 16 (CNPG) performance on the cluster:

```bash
# Connect to the database pod
kubectl exec -it workshop-db-1 -- bash

# Initialize pgbench
pgbench -i -s 10 vikunja

# Run benchmark (read/write mix)
pgbench -c 10 -j 2 -T 60 vikunja

# Read-only benchmark
pgbench -c 10 -j 2 -T 60 -S vikunja
```

### 3. Storage IOPS (fio)

Benchmark Longhorn storage performance:

```bash
# Create a test pod with a Longhorn PVC
kubectl run fio-test --image=nixery.dev/fio --restart=Never --overrides='
{
  "spec": {
    "containers": [{
      "name": "fio-test",
      "image": "nixery.dev/fio",
      "command": ["sleep", "3600"],
      "volumeMounts": [{"name": "test-vol", "mountPath": "/data"}]
    }],
    "volumes": [{
      "name": "test-vol",
      "persistentVolumeClaim": {"claimName": "vikunja-files"}
    }]
  }
}'

# Random read/write IOPS
kubectl exec fio-test -- fio --name=randwrite --ioengine=libaio --rw=randwrite \
  --bs=4k --direct=1 --size=1G --numjobs=4 --time_based --runtime=60 \
  --group_reporting --directory=/data

kubectl exec fio-test -- fio --name=randread --ioengine=libaio --rw=randread \
  --bs=4k --direct=1 --size=1G --numjobs=4 --time_based --runtime=60 \
  --group_reporting --directory=/data

# Cleanup
kubectl delete pod fio-test
```

### 4. App Test Locally (Vikunja)

Test the Vikunja todo app before deploying:

```bash
# Run locally
cd apps/vikunja
docker run -p 3456:3456 vikunja/vikunja:latest

# Open http://localhost:3456
# - Create an account
# - Create a project and tasks
# - Test file upload on a task (attachments)
```

## Repository Structure

```
workshop-akasha/
├── README.md
├── apps/
│   └── vikunja/
│       └── Dockerfile                      # Customizable Vikunja image
├── .github/
│   └── workflows/
│       └── build-vikunja.yaml              # CI: build + push to Zot on app changes
└── clusters/
    └── workshop-todo/
        ├── kustomization.yaml              # Root → Flux Kustomizations
        ├── flux/
        │   ├── infra.yaml                  # Layer 1: operators (cert-manager, CNPG, Longhorn, Traefik)
        │   ├── infra-services.yaml         # Layer 2: ClusterIssuer, external-dns, Zot (depends on Layer 1)
        │   └── apps.yaml                   # Layer 3: app workloads (depends on Layer 2)
        ├── infra/                          # Operators & CRDs
        │   ├── cert-manager.yaml           # TLS certificate operator
        │   ├── cnpg.yaml                   # CloudNative PostgreSQL operator
        │   ├── longhorn.yaml               # Distributed block storage
        │   └── traefik.yaml                # Ingress controller
        ├── infra-services/                 # CRD-dependent services
        │   ├── cluster-issuer.yaml         # Let's Encrypt + Cloudflare DNS-01 solver
        │   ├── external-dns.yaml           # Auto DNS records: Ingress → Cloudflare cl8s.dev
        │   └── zot.yaml                    # OCI container registry
        └── apps/                           # Application workloads
            ├── vikunja.yaml                # Todo app (Deployment, Service, Ingress, PVC, ConfigMap)
            ├── workshop-db.yaml            # PostgreSQL 16 cluster (CNPG)
            ├── github-runner.yaml          # Self-hosted GitHub Actions runner
            ├── workshop-db-user.yaml       # DB credentials (CHANGE_ME)
            ├── workshop-db-backup-s3.yaml  # R2 backup credentials (CHANGE_ME)
            └── workshop-db-scheduled-backup.yaml
```

### Dependency Chain

```
infra (operators/CRDs)
  └─► infra-services (ClusterIssuer, external-dns, Zot)
        └─► apps (Vikunja, PostgreSQL, GitHub Runner)
```

## Setup

### Prerequisites

- Kubernetes cluster with Flux CD installed
- Cloudflare API token for `cl8s.dev` (DNS + TLS)
- Replace all `CHANGE_ME` placeholders:
  - `infra-services/cluster-issuer.yaml` — Cloudflare API token (cert-manager)
  - `infra-services/external-dns.yaml` — Cloudflare API token (DNS records)
  - `apps/workshop-db-user.yaml` — database password
  - `apps/workshop-db-backup-s3.yaml` — R2/S3 backup credentials
  - `apps/github-runner.yaml` — GitHub personal access token

### Node Requirements

The cluster needs at least 3 node pools with specific labels and taints:

| Node Pool | Min Nodes | Labels | Taints | Workloads |
|-----------|-----------|--------|--------|-----------|
| **Worker** | 2 | `node-role.kubernetes.io/worker: worker` | — | Vikunja app, GitHub Runner, Traefik |
| **Database** | 1 | `workload: db` | `workload=db:NoSchedule` | PostgreSQL (CNPG) |
| **Storage** | 2 | `node.longhorn.io/create-default-disk: true` | — | Longhorn replicas |

**Worker nodes** — general app workloads. Vikunja, GitHub Actions runner, and Traefik are scheduled here via `nodeSelector: {node-role.kubernetes.io/worker: worker}`.

**Database nodes** — dedicated to PostgreSQL. The DB pod uses `nodeSelector: {workload: db}` and tolerates the `workload=db:NoSchedule` taint, so only database workloads run on these nodes.

**Storage nodes** — Longhorn stores volume replicas on nodes labeled `node.longhorn.io/create-default-disk: true`. These can overlap with worker nodes if needed, but dedicated storage nodes are recommended for IOPS isolation.

### Deploy

```bash
# Create Flux GitRepository
flux create source git workshop-akasha \
  --url=https://github.com/ryanavilar/workshop-akasha \
  --branch=main \
  --interval=1m

# Create root Flux Kustomization
flux create kustomization workshop-akasha \
  --source=workshop-akasha \
  --path=./clusters/workshop-todo \
  --prune=false \
  --interval=1m
```
