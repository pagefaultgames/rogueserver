# rogueserver

# Hosting in Docker
It is advised that you host this in a docker container as it will be much easier to manage. 
There is a sample docker-compose file for setting up a docker container to setup this server.

# Self Hosting outside of Docker:
Recommended Tools:
If using Windows: [Chocolatey](https://chocolatey.org/install)

## Required Tools:
- Golang
- Node: **18.3.0**
- npm: [how to install](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

## Installation:
The docker compose file should automatically implement a container with mariadb with an empty database and the default user and password combo of pokerogue:pokerogue

Edit the following files:
### rogueserver.go:34 
Change the 'false' after "debug" to 'true'. This will resolve CORS issues that many users have been having while trying to spin up their own servers. 

### rogueserver.go:37
Change the default port if you need to, I set it to 8001. As of another pull request, this should _NOT_ be necessary.

### rogueserver.go:41-43
You can choose to specify a different dbaddr, dbproto, or dbname if you so choose, instead of passing a flag during execution.
It is advised that you do not store hardcoded credential sets, but rather pass them in as flags during execution like in the commands below. 

### src/utils.ts:224-225 (in pokerogue)
Replace both URLs (one on each line) with the local API server address from rogueserver.go (0.0.0.0:8001) (or whatever port you picked)

# If you are on Windows

Now that all of the files are configured: start up powershell as administrator:
```
powershell -ep bypass
cd C:\api\server\location\
go run . -dbuser [usernamehere] -dbpass [passhere]
```

Then in another run this the first time then run `npm run start` from the rogueserver location from then on:
```
powershell -ep bypass
cd C:\rogue\server\location\
npm install
npm run start
```
You will need to allow the port youre running the API (8001) on and port 8000 to accept inbound connections through the [Windows Advanced Firewall](https://www.youtube.com/watch?v=9llH5_CON-Y).

# If you are on Linux
In whatever shell you prefer, run the following:
```
cd /api/server/location/
go run . -dbuser [usernamehere] -dbpass [passhere] &
cd /rogue/server/location/
npm run start &
```
If you have a firewall running such as ufw on your linux machine, make sure to allow inbound connections on the ports youre running the API and the pokerogue server (8000,8001).
An example to allow incoming connections using UFW:
```
sudo ufw allow 8000,8001/tcp
```

This should allow you to reach the game from other computers on the same network. 

## Tying to a Domain

If you want to tie it to a domain like I did and make it publicly accessible, there is some extra work to be done.

I setup caddy and would recommend using it as a reverse proxy. 
[caddy installation](https://caddyserver.com/docs/install)
once its installed setup a config file for caddy:

```
pokerogue.exampledomain.com {
	reverse_proxy localhost:8000
}
pokeapi.exampledomain.com {
	reverse_proxy localhost:8001
} 
```
Preferably set up caddy as a service from [here.](https://caddyserver.com/docs/running)

Once this is good to go, take your API url (https://pokeapi.exampledomain.com) and paste it on 
### src/utils.ts:224-225 
in place of the previous 0.0.0.0:8001 address

Make sure that both 8000 and 8001 are portforwarded on your router.

Test that the server's game and game authentication works from other machines both in and outside of the network. Once this is complete, enjoy!


