# A better launcher for Screeps private servers

## Why?
* The steam private server has a few limitations, one being that getting non-workshop mods to work is a huge headache. 
* The npm version is much better, but requires care in installing everything correctly.

Therefore, the goal of this is to simplify the entire process making it much easier to use. 
No need to manually `npm install` anything, its handled automatically

## Usage
1. Download a release from the [Releases](https://github.com/ags131/screeps-launcher/releases) Page
2. Drop into an empty folder or your PATH
3. Get your [Steam API key](https://steamcommunity.com/dev/apikey)
4. Create config.yml (All fields are optional! You can pass STEAM_KEY as an environment variable)
  ```yaml
  steamKey: keyFromStep3
  mods: # Recommended mods
  - screepsmod-auth
  - screepsmod-tickrate
  - screepsmod-mongo  # You must install and start `mongodb` and `redis` before this mod will work
  bots:
    simplebot: screepsbot-zeswarm
  ```
5. Open a shell to folder
6. Run `screeps-launcher`
7. If you installed `screepsmod-mongo`, run `screeps-launcher cli` in another shell, and run `system.resetAllData()` to init the DB. It completes instantly, restart the server after.
8. Done!

You can use `screeps-launcher cli` in the same folder for CLI access

### Other options

There are several extra arguments that can be used to manage the install:
* `screeps-launcher apply` Applies the current config.yml without starting the server.
* `screeps-launcher upgrade` Upgrades all packages (screeps, mods, bots, etc)
* `screeps-launcher cli` Launch a screeps cli

## Docker

If you don't use Docker, you probably want the regular [usage](#usage) section above

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
