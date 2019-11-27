# A better launcher for Screeps private servers

[![CircleCI](https://circleci.com/gh/screepers/screeps-launcher/tree/master.svg?style=shield)](https://circleci.com/gh/screepers/screeps-launcher/tree/master)

## Why?
* The steam private server has a few limitations, one being that getting non-workshop mods to work is a huge headache. 
* The npm version is much better, but requires care in installing everything correctly.

Therefore, the goal of this is to simplify the entire process making it much easier to use. 
No need to manually `npm install` anything, its handled automatically

## Guides
If installing on ubuntu 18.04 or on a Pi with raspbian, theres also a guide on
reddit 
[here](https://www.reddit.com/r/screeps/comments/deyq66/newbiefriendly_ish_privatededicated_server_setup/)
that does a step-by-step setup including mongo, redis, and auto start.

## Usage
1. Download a release from the [Releases](https://github.com/screepers/screeps-launcher/releases) Page
2. Drop into an empty folder or your PATH
3. Get your [Steam API key](https://steamcommunity.com/dev/apikey)
4. Create config.yml (All fields are optional! You can pass STEAM_KEY as an environment variable)
  ```yaml
  steamKey: keyFromStep3
  mods: # Recommended mods
  - screepsmod-auth
  - screepsmod-admin-utils
  - screepsmod-mongo  # You must install and start `mongodb` and `redis` before this mod will work
  bots:
    simplebot: screepsbot-zeswarm
  serverConfig: # This section requires screepsmod-admin-utils to work
    welcomeText:  |
      <h1 style="text-align: center;">My Cool Server</h1>
    constants: # Used to override screeps constants
      TEST_CONSTANT: 123
    tickRate: 1000  # In milliseconds. This is a lower bound. Users reported problems when set too low.
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
* `screeps-launcher backup <file>` Creates a backup
* `screeps-launcher restore <file>` Restores a backup (Warning: Completely replaces existing data)

## docker-compose
There is also an example [docker-compose.yml](docker-compose.yml) that starts a server + mongo.
This is the easiest way to get a private server working on windows and using mongo + redis.

1. Install [docker](https://docs.docker.com/install/) (look on the left to find the correct platform).
2. You might have to fiddle with the docker advanced settings to allow enough CPU to run the server smoothly.
3. Create an empty folder with both a `config.yml` (don't forget to add `screepsmod-mongo`!) and a `docker-compose.yml` (see examples). The `docker-compose.yml` example can be used as-is, but the `config.yml` requires some customization.
4. Open a terminal in that folder. Run `docker-compose up` to start the services. Wait until it is done starting the docker images and settle on mongo status messages.
5. Open another terminal in that folder. Run `docker-compose exec screeps screeps-launcher cli`. This is a command-line interface to control your new private server.
6. In the CLI, run `system.resetAllData()` to initialize the database. Unless you want to poke around, use `Ctrl-d` to exit the cli.
7. Run `docker-compose restart screeps` to reboot the private server.

Your server should be up and running! Connect to it using the steam client:

Choose the _Private Server_ tab and connect using those options:
- Host: _localhost_
- Port: _21025_
- Server password: _<leave blank, unless configured otherwise>_

## Docker
Docker builds are published to Dockerhub as `screepers/screeps-launcher`
Quickstart:
1. Create config file in an empty folder (`/srv/screeps` is used for this example)
2. Run `docker run --restart=unless-stopped --name MyScreepsServer -v /srv/screeps:/screeps -p 21025:21025 screepers/screeps-launcher`
3. Done! 

## Mods
You can easily install mods by adding their names to the `config.yml` file. The `screeps-launcher` takes care of downloading them for you. Mods can be found in the [ScreepsMods github repository](https://github.com/ScreepsMods).

A few mods of interests:
- `screepsmod-mongo` (needed to actually use mongo+redis!)
- `screepsmod-auth`  (use it to change your password by going to [http://localhost:21025/authmod/password])
- `screepsmod-admin-utils`
- `screepsmod-map-tool`
- `screepsmod-history`
- `screepsmod-market`

See each of their documentation on the [ScreepsMods github repository](https://github.com/ScreepsMods).
