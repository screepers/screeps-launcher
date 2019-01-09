# A better launcher for Screeps private servers

## Usage:
1. Drop in an empty folder
2. Get your (Steam API key)[https://steamcommunity.com/dev/apikey]
3. Create config.yml
	```yaml
	env:
		backend:
			STEAM_KEY: keyFromStep2
	mods:
	- screepsmod-auth
	bots:
		simplebot: screepsbot-zeswarm
	```
4. Open a shell to folder
5. Run `screeps-launcher`
6. Done!

You can use `npx screeps cli` in the same folder for CLI access

Note: If using `screepsmod-mongo`, run `system.resetAllData()` in CLI to init the DB

