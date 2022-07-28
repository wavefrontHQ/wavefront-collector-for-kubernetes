# scripts

This directory contains helpful scripts to aid in dashboard and alert development.

1. To start working on an existing dashboard:
cd integrations
./scripts/start-dashboard-development.sh  -t $WAVEFRONT_TOKEN -d <dasboard url>

2. PM or engg member iterate on dev dashboard returned from start-dashboard-development.sh

3. When dev dashboard is ready, merge to branch in integration repo:
./scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dasboard url>