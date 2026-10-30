---
title: "Version 0.10"
description: "Release notes for GDG v0.10, including 12-factor config compliance: environment variables now take precedence over secure auth files."
date: 2026-10-30T00:00:00
draft: false
images: [ ]
weight: 1
toc: true
---

## Release Notes for v0.10.0

**Release Date: TBD**

---

### Breaking Changes (v0.10.0)

#### Environment Variables Now Take Precedence Over Secure Auth Files

GDG now follows [12-factor app](https://12factor.net/config) configuration
principles. Environment variables always win over values stored in secure auth
files, on a per-field basis.

##### Grafana Authentication

Previously, `GDG_CONTEXTS__<CTX>__TOKEN` and `GDG_CONTEXTS__<CTX>__PASSWORD`
were only consulted when the secure auth file (`auth_<context>.yaml`) was
missing or unreadable. From v0.10.0 onward they are **always** evaluated and
take precedence over the file.

Field-level merging applies: if you set only `TOKEN` via env, the `PASSWORD`
from the file still fills in, and vice-versa.

**Old precedence (≤ 0.9.x):**
1. Secure auth file (`auth_<context>.yaml`)
2. Env vars (`GDG_CONTEXTS__<CTX>__TOKEN` / `PASSWORD`) — only checked on file failure
3. Anonymous access

**New precedence (≥ 0.10.0):**
1. Env vars (`GDG_CONTEXTS__<CTX>__TOKEN` / `PASSWORD`) — always checked, always win
2. Secure auth file (`auth_<context>.yaml`)
3. Anonymous access

**Migration:** If you have auth files in your deployment *and* env vars set,
the env vars will now win. Ensure your env vars contain the credentials you
intend to use, or unset them to fall back to the file.

##### Cloud / S3 Storage Authentication

GDG now exposes standard GDG-namespaced environment variables for overriding
custom S3-compatible storage credentials:

```
GDG_STORAGE_ENGINE__<STORAGE_NAME>__ACCESS_ID=myaccesskey
GDG_STORAGE_ENGINE__<STORAGE_NAME>__SECRET_KEY=mysecretkey
```

For example, for a storage configuration named `super_cloud`:

```bash
export GDG_STORAGE_ENGINE__SUPER_CLOUD__ACCESS_ID=myaccesskey
export GDG_STORAGE_ENGINE__SUPER_CLOUD__SECRET_KEY=mysecretkey
```

These env vars follow the same precedence rule: they win over the secure cloud
auth file (`s3_<storage>.yaml`) on a per-field basis.

#### Removal of `AWS_ACCESS_KEY` / `AWS_SECRET_KEY` Support

`AWS_ACCESS_KEY` and `AWS_SECRET_KEY` were **never officially supported** by
GDG — they were an undocumented side effect of an internal implementation
detail. They are now explicitly **unset** by GDG at runtime so they cannot
cause unexpected behavior in downstream code.

{{% notice style="warning" title="Action Required" %}}
If you were relying on `AWS_ACCESS_KEY` / `AWS_SECRET_KEY` to supply custom
S3 credentials, migrate to the new GDG-namespaced variables:

```bash
# Before (unsupported, no longer works)
export AWS_ACCESS_KEY=mykey
export AWS_SECRET_KEY=mysecret

# After
export GDG_STORAGE_ENGINE__<STORAGE_NAME>__ACCESS_ID=mykey
export GDG_STORAGE_ENGINE__<STORAGE_NAME>__SECRET_KEY=mysecret
```

Standard AWS SDK environment variables (`AWS_ACCESS_KEY_ID`,
`AWS_SECRET_ACCESS_KEY`) used for native AWS S3 buckets are **not affected**
by this change.
{{% /notice %}}
