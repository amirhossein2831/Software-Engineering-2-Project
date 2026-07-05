#!/bin/sh
#
# register-runner.sh <glrt-token>
# Registers the gitlab-runner container against the local GitLab instance.

TOKEN="$1"

if [ -z "$TOKEN" ]; then
  echo "Usage: $0 <glrt-token>"
  exit 1
fi

docker exec -it gitlab-runner gitlab-runner register \
  --non-interactive \
  --url "http://gitlab.local" \
  --token "$TOKEN" \
  --executor "docker" \
  --docker-image "alpine:latest" \
  --description "docker-runner"
