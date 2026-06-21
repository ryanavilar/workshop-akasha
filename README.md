# Workshop Akasha

Hands-on workshop for Kubernetes infrastructure testing on the Akasha platform.

## About Akasha

Akasha is a private cloud platform that provides a web console for managing virtualized infrastructure and Kubernetes clusters. Built on top of Incus (the container and VM hypervisor), it gives teams a single interface to operate compute, storage, and networking resources across bare-metal hosts — without relying on public cloud providers.

### Instances

Instances are the core compute units in Akasha. They can be either system containers (lightweight, sharing the host kernel) or full virtual machines (running their own OS kernel). Through the web console you can create, start, stop, snapshot, migrate, and monitor instances. Each instance gets its own terminal and graphical console accessible directly from the browser, along with real-time CPU, memory, and filesystem metrics.

### Kubernetes

Akasha integrates Cluster API (CAPI) with an Incus provider to provision and manage Kubernetes clusters as tenant workloads. From the web console you can create new clusters, inspect nodes, manage Helm releases, and open a browser-based terminal with `kubectl` and `k9s` pre-configured. Each tenant cluster runs on VMs inside its own Incus project, fully isolated from other tenants.

### Profiles

Profiles are reusable configuration templates applied to instances at creation time. A profile defines defaults for devices (disks, NICs), resource limits (CPU, memory), and cloud-init configuration. Instances can inherit from multiple profiles, and changes to a profile propagate to all instances using it — making it easy to enforce consistent configuration across fleets of containers or VMs.

### Networking

Akasha supports multiple network types: OVN overlay networks for tenant isolation with built-in DHCP, DNS, and load balancers; bridged networks for direct host connectivity; and managed network forwards, ACLs, and peering between networks. The web console provides a topology view, IPAM (IP address management), lease tracking, and configuration of load balancers and forwarding rules.

### Images

Images are the base OS templates used to launch instances. Akasha can pull images from public remote servers (e.g., Ubuntu, Debian, Alpine) or from custom OCI registries. You can also upload custom ISOs for VM installation, create images from existing instance snapshots, and manage a local image library. Images are cached locally and can be replicated across cluster members.

### Storage

Akasha manages storage through pools and volumes. Storage pools abstract the underlying driver (ZFS, LVM, Btrfs, Ceph, or directory-backed) and expose volumes that instances attach as root disks or additional data disks. The web console lets you create pools, provision and resize custom volumes, take volume snapshots, migrate volumes between cluster members, and manage S3-compatible storage buckets.

## Workshop Steps

### Step 1: Create a Network

Before creating any clusters, you need an OVN network for your tenant.

1. Open the Akasha web console and navigate to your project
2. Go to **Networks** → **Create network**
3. Configure the network:
   - **Name**: `workshop` (or your preferred name)
   - **Type**: OVN
   - **IPv4 address**: `auto` (assigns a subnet automatically)
   - **IPv6 address**: `none`
4. Click **Create**

The OVN network provides tenant-isolated L2/L3 networking with built-in DHCP, DNS, and support for load balancers and forwarding rules.

### Step 2: Launch a VM (Optional)

To explore Akasha's instance management before spinning up a full cluster:

1. Go to **Instances** → **Create instance**
2. Select **VM** as the instance type
3. Choose a base image (e.g., Ubuntu 24.04)
4. Select a profile with appropriate resource limits
5. Attach the network you created in Step 1
6. Click **Create and start**

From the instance detail page you can open a terminal, view the graphical console, monitor resource usage, and take snapshots.

### Step 3: Create a Kubernetes Cluster

This is the main step — provision a Kubernetes cluster that will run the workshop workloads.

1. Go to **Kubernetes** → **Create cluster**
2. Fill in the cluster settings:
   - **Cluster name**: `workshop-todo`
   - **Network**: select the network from Step 1
   - **Root disk pool**: select an available storage pool
3. Select **Add-ons**:
   - **Ingress** (installs cert-manager + Traefik)
   - **Database** (installs cert-manager + CNPG)
   - **CI & CD** (installs cert-manager + Flux + Zot registry)
   - **Storage** is auto-enabled when Database or CI & CD is selected
4. Configure **storage**:
   - **Storage type**: Longhorn (distributed block storage) or OpenEBS-LVM (local)
   - **Data disk size**: 50 GiB or more
5. Configure **control plane**:
   - **CP size**: `c2-m4` (2 vCPU, 4 GB) minimum, `c4-m8` recommended with add-ons
   - **CP replicas**: 1 for workshop, 3 for HA
6. Configure **workers**:
   - **Worker size**: `c4-m8` (4 vCPU, 8 GB) recommended
   - **Worker count**: 0 for now (we'll add worker pools with specific roles next)
7. Click **Create Cluster**

Wait for the cluster phase to change from "Provisioning" to "Provisioned".

### Step 4: Add Worker Node Pools

The workshop workloads need node pools with specific labels and taints. From the cluster detail page, go to the **Nodes** tab and add each pool:

**Pool 1 — Worker** (general app workloads):
- **Name**: `worker`
- **Size**: `c4-m8`
- **Replicas**: 2
- **Node labels**: `node-role.kubernetes.io/worker: worker`
- No taints

**Pool 2 — Database** (dedicated PostgreSQL nodes):
- **Name**: `database`
- **Size**: `c4-m8`
- **Replicas**: 1
- **Node labels**: `workload: db`
- **Restrict scheduling** (taint): key `workload`, value `db`, effect `NoSchedule`

**Pool 3 — Storage** (Longhorn replica nodes):
- **Name**: `storage`
- **Size**: `c2-m4`
- **Replicas**: 2
- **Node labels**: `node.longhorn.io/create-default-disk: true`
- No taints

Wait for all machines to reach "Running" phase before proceeding.

### Step 5: Access the Cluster (Browser Terminal)

Once all nodes are ready, access the cluster terminal directly from the web console:

1. From the cluster detail page, click the **Terminal** tab
2. A browser-based terminal opens with `kubectl` and `k9s` pre-configured
3. Verify the nodes are ready:

```bash
kubectl get nodes
```

You should see all 5 nodes (1 CP + 2 worker + 1 database + 2 storage) in `Ready` status.

### Step 6: Download Kubeconfig (Access from Local Machine)

To manage the cluster from your local terminal instead of the browser:

1. From the cluster detail page, click the **download** icon (⬇) in the header
2. Select **Kubeconfig**
3. Enter the **API Server Address** — this is the IP or domain of the network load balancer that fronts the cluster's API server (e.g., `10.99.47.6:6443` or `workshop-todo.example.com:6443`). Leave empty to use the internal cluster IP.
4. Click **Download** — saves as `workshop-todo-kubeconfig.yaml`

Use it locally:

```bash
# Option A: set for the current shell session
export KUBECONFIG=~/Downloads/workshop-todo-kubeconfig.yaml
kubectl get nodes

# Option B: merge into your default kubeconfig
cp ~/.kube/config ~/.kube/config.bak
KUBECONFIG=~/.kube/config:~/Downloads/workshop-todo-kubeconfig.yaml \
  kubectl config view --flatten > /tmp/merged-config
mv /tmp/merged-config ~/.kube/config

# Switch context to the workshop cluster
kubectl config get-contexts
kubectl config use-context workshop-todo
kubectl get nodes
```

> **Note**: Your local machine must be able to reach the API server address you entered. If the cluster runs behind a private network, you may need VPN or SSH tunnel access to the host.

### Step 7: Deploy the GitOps Stack

With the cluster ready and Flux installed (via the CI & CD add-on), point Flux at this repository:

```bash
# Create Flux GitRepository
flux create source git workshop-akasha \
  --url=https://github.com/<YOUR_ORG>/workshop-akasha \
  --branch=main \
  --interval=1m

# Create root Flux Kustomization
flux create kustomization workshop-akasha \
  --source=workshop-akasha \
  --path=./clusters/workshop-todo \
  --prune=false \
  --interval=1m
```

Watch the deployment progress:

```bash
kubectl -n flux-system get kustomizations
```

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
        │   ├── cluster-issuer.yaml         # Let's Encrypt + DNS-01 solver
        │   ├── external-dns.yaml           # Auto DNS records via Cloudflare
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

- Access to the Akasha web console with a project assigned
- A domain managed via Cloudflare DNS (for TLS and automatic DNS records)
- Search and replace all `cl8s.dev` references across manifests with your actual domain
- Replace all `CHANGE_ME` placeholders:
  - `infra-services/cluster-issuer.yaml` — Cloudflare API token + email (cert-manager)
  - `infra-services/external-dns.yaml` — Cloudflare API token (DNS records)
  - `apps/workshop-db-user.yaml` — database password
  - `apps/workshop-db-backup-s3.yaml` — R2/S3 backup credentials
  - `apps/github-runner.yaml` — GitHub personal access token
