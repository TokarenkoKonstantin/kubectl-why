# kubectl-why 🔍

> **One command that tells you _why_ a pod isn't running** — instead of scrolling through 200 lines of `kubectl describe`.

![status](https://img.shields.io/badge/status-active%20development-yellow)
![go](https://img.shields.io/badge/v0.2-Go-00ADD8?logo=go&logoColor=white)
![license](https://img.shields.io/badge/license-MIT-blue)

<p align="center">
  <img src="demo.gif" alt="kubectl-why diagnosing a broken pod" width="820">
</p>

---

## The problem

Your pod is `Pending` / `CrashLoopBackOff` / `ImagePullBackOff`… and you dig through `kubectl describe pod` and a wall of events hunting for the actual reason. Every Kubernetes engineer knows this pain.

## The solution

```console
$ kubectl why payment-api

🔍 Diagnosing pod "payment-api" (phase: Pending)…

❌ Container "payment-api": ImagePullBackOff
   └─ Kubernetes can't pull the image.
   └─ Check: image name/tag typo, or private registry needs an imagePullSecret.
```

One clear diagnosis with a concrete next step — instead of reading raw YAML.

### What it detects

- ✅ Healthy pods (and says so)
- ❌ `CrashLoopBackOff` → points you to the right logs
- ❌ `ImagePullBackOff` / `ErrImagePull` → image typo or missing pull secret
- ❌ `CreateContainerConfigError` → names the missing ConfigMap/Secret
- ❌ `Pending` → the scheduler's actual reason (resources, taints, PVC)
- ❌ `OOMKilled` (exit 137) → out of memory

---

## Install

### Go (any platform — single static binary, no dependencies)

```bash
go install github.com/TokarenkoKonstantin/kubectl-why@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`, then call it as `kubectl why`.

### Pre-built binary

Grab the archive for your OS from [Releases](https://github.com/TokarenkoKonstantin/kubectl-why/releases), unpack it, and move `kubectl-why` into your `PATH`.

> 🚧 Krew distribution coming in v1.0 — `kubectl krew install why`

### Bash version (no Go; needs `jq`)

```bash
curl -sLO https://raw.githubusercontent.com/TokarenkoKonstantin/kubectl-why/main/legacy/kubectl-why
chmod +x kubectl-why && sudo mv kubectl-why /usr/local/bin/
```

## Usage

```bash
kubectl why <pod> [-n namespace]
```

---

## Roadmap

- [x] **v0.1** — bash MVP (Pending, CrashLoop, ImagePull, OOMKilled)
- [x] **v0.2** — rewritten in Go with `client-go`; cross-platform binaries via GoReleaser
- [ ] **v0.3** — diagnose Deployments/StatefulSets, not just pods
- [ ] **v1.0** — submit to the [Krew](https://krew.sigs.k8s.io/) index

---

## Build from source

```bash
git clone https://github.com/TokarenkoKonstantin/kubectl-why
cd kubectl-why
go build -o kubectl-why .
```

## Author

**Konstantin Tokarenko** — DevOps Engineer
[GitHub](https://github.com/TokarenkoKonstantin) · [Telegram](https://t.me/KonstantinTokar)

Contributions and ideas welcome — open an issue!
