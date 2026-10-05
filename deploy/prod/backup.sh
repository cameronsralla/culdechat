#!/bin/sh
# Nightly backup sidecar: pg_dump + media tarball into /backups, pruned by age.
# Runs at ~03:00 container time. Restore: see docs/admin-guide.md.
set -eu

KEEP_DAYS="${BACKUP_KEEP_DAYS:-14}"

run_backup() {
  stamp=$(date +%Y%m%d-%H%M%S)
  echo "[backup] starting $stamp"
  pg_dump --format=custom --no-owner --file="/backups/db-$stamp.dump"
  if [ -d /data/media ] && [ -n "$(ls -A /data/media 2>/dev/null)" ]; then
    tar -czf "/backups/media-$stamp.tar.gz" -C /data media
  fi
  find /backups -type f \( -name 'db-*.dump' -o -name 'media-*.tar.gz' \) -mtime "+$KEEP_DAYS" -delete
  echo "[backup] done $stamp"
}

# Run once on start so a fresh install has a baseline, then nightly.
run_backup || echo "[backup] initial backup failed"
while :; do
  now=$(date +%s)
  target=$(date -d "tomorrow 03:00" +%s 2>/dev/null || echo $((now + 86400)))
  sleep $((target - now))
  run_backup || echo "[backup] nightly backup failed"
done
