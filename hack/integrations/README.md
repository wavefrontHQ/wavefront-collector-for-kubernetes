# scripts

This directory contains helpful scripts to aid in dashboard and alert development.

To start working on an existing dashboard:

cd integrations 


./scripts/start-dashboard-development.sh  -t $WAVEFRONT_TOKEN -d <dasboard url>


update dashboard on WF cluster (defaults to nimba)

when down run:

./scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dasboard url>