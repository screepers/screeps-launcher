To setup and use a local Grafana via docker-compose, follow these steps:

- Copy the `docker-compose.yml` from the stats folder to the parent folder. (Take care not to overwrite any changes you have made)
- Edit setup.json to fit your needs. See [https://github.com/ScreepsPlus/node-hosted-agent] for details. The example reads from segment 98 every 15 seconds.
- Start everything with `docker-compose up -d`
- Access Grafana at [http://localhost:3000/]. You'll find your stats in the default data source, under `screeps.privateserver`.
