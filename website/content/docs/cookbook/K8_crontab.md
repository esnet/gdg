---
title: "Kubernetes Crontab Example"
weight: 52
---

This is a bit beyond the scope of GDG, but wanted to provide an example of how to setup GDG to run on a regular cadence and take backups of your dashboards.  Example provided courtesy of [arnecls](https://github.com/arnecls)

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  labels:
    helm.sh/chart: grafana-operator-package-5.6.1
  name: grafana-backup
  namespace: grafana-operator
spec:
  concurrencyPolicy: Forbid
  failedJobsHistoryLimit: 1
  schedule: 0 * * * *
  successfulJobsHistoryLimit: 1
  jobTemplate:
    spec:
      backoffLimit: 0
      completions: 1
      template:
        spec:
          containers:
            - args:
                - -c
                - /etc/config/gdg.yaml
                - backup
                - dashboards
                - upload
                - --skip-confirmation
              image: ***/ghcr-io-mirror/esnet/gdg:0.7.1
              name: grafana-backup-dashboards
              resources:
                limits:
                  memory: 256Mi
                requests:
                  cpu: 100m
                  memory: 256Mi
              securityContext:
                allowPrivilegeEscalation: false
                capabilities:
                  drop:
                    - ALL
                readOnlyRootFilesystem: true
                runAsGroup: 65532
                runAsNonRoot: true
                runAsUser: 65532
              volumeMounts:
                - mountPath: /etc/config/
                  name: config
                  readOnly: true
                - mountPath: /app/backup # <--- this
                  name: scratch
                  readOnly: false
          restartPolicy: Never
          serviceAccountName: grafana-operator-grafana-sa
          volumes:
            - configMap:
                defaultMode: 0444
                name: grafana-backup
              name: config
            - emptyDir: {}
              name: scratch

```
