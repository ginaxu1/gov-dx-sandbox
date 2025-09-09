#!/bin/bash
# View logs

SERVICE=${1:-""}
[ -n "$SERVICE" ] && docker compose logs -f "$SERVICE" || docker compose logs -f