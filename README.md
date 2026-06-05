# kubectl-why 🔍

> **One command that tells you _why_ a pod isn't running** — instead of scrolling through 200 lines of `kubectl describe`.

![status](https://img.shields.io/badge/status-active%20development-yellow)
![shell](https://img.shields.io/badge/v0.1-bash-4EAA25?logo=gnu-bash&logoColor=white)
![license](https://img.shields.io/badge/license-MIT-blue)

---

## The problem

Your pod is `Pending` / `CrashLoopBackOff` / `ImagePullBackOff`… and you dig through `kubectl describe pod` and a wall of events hunting for the actual reason. Every Kubernetes engineer knows this pain.

## The solution

```console
$ kubectl why my-pod

❌ Pod "my-pod" is Pending — not scheduled to any node.
   └─ Scheduler says: 0/4 nodes available: insufficient memory.
   └─ Common causes: insufficient CPU/memory, node taints, unbound PVC.
```

One clear diagnosis with a concrete next step — instead of reading raw YAML.

### What it detects (v0.1)

- ✅ Healthy pods (and says so)
- ❌ `CrashLoopBackOff` → points you to the right logs
- ❌ `ImagePullBackOff` / `ErrImagePull` → image typo or missing pull secret
- ❌ `CreateContainerConfigError` → missing ConfigMap/Secret
- ❌ `Pending` → scheduling reason (resources, taints, PVC)
- ❌ `OOMKilled` (exit 137) → out of memory

---

## Install

> 🚧 Krew distribution coming in v1.0 — `kubectl krew install why`

**Now (bash MVP):**

```bash
curl -sLO https://raw.githubusercontent.com/TokarenkoKonstantin/kubectl-why/main/kubectl-why
chmod +x kubectl-why
sudo mv kubectl-why /usr/local/bin/
```

Requires `kubectl` and `jq`.

## Usage

```bash
kubectl why <pod> [-n namespace]
```

---

## Roadmap

- [x] **v0.1** — bash MVP (Pending, CrashLoop, ImagePull, OOMKilled)
- [ ] **v0.2** — rewrite in Go with `client-go`, cross-platform binaries
- [ ] **v0.3** — diagnose Deployments/StatefulSets, not just pods
- [ ] **v1.0** — submit to the [Krew](https://krew.sigs.k8s.io/) index

---

## Author

**Konstantin Tokarenko** — DevOps Engineer
[GitHub](https://github.com/TokarenkoKonstantin) · [Telegram](https://t.me/KonstantinTokar)

Contributions and ideas welcome — open an issue!
