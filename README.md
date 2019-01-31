# A better launcher for Screeps private servers

## Usage:
1. Download a release from the [Releases](https://github.com/ags131/screeps-launcher/releases) Page
2. Drop into an empty folder or your PATH
3. Get your [Steam API key](https://steamcommunity.com/dev/apikey)
4. Create config.yml (All fields are optional! You can pass STEAM_KEY as an environment variable)
  ```yaml
  steamKey: keyFromStep2
  mods:
  - screepsmod-auth
  bots:
    simplebot: screepsbot-zeswarm
  ```
5. Open a shell to folder
6. Run `screeps-launcher`
7. Done!

You can use `npx screeps cli` in the same folder for CLI access

Note: If using `screepsmod-mongo`, run `system.resetAllData()` in CLI to init the DB

## Docker

A docker image is also built and published to quay.io

A minimal server can be ran with
```bash
docker run -e STEAM_KEY=<key> --name server quay.io/ags131/screeps-launcher
```

Then just use `docker stop server` and `docker start server` to start and stop it.

You can mount a local folder in to set config.yml or to add local mods
```bash
docker run -e STEAM_KEY=<key> -v $PWD/server:/screeps --name server quay.io/ags131/screeps-launcher
```

You can also bring it up with the included docker-compose.yml