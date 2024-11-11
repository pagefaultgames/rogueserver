# rogueserver

# Hosting in Docker
It is advised that you host this in a docker container as it will be much easier to manage. 
There is a sample docker-compose file for setting up a docker container to setup this server.

# Self Hosting outside of Docker:
## Required Tools:
- Golang
- Node: **18.3.0**
- npm: [how to install](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

## Installation:
### First Steps
- Edit beta.env
	- Setting VITE_BYPASS_LOGIN to 0 helps provide access to PokeRogue's accounts features
	- If testing locally without an S3 instance, set local to true. 
	- gameurl should reference the IP of your development machine
	- callbackurl should reference the IP of your server
- Edit docker-compose.Example.yml
	- Under services->server, add the following lines right above image: rogueserver:latest
	```
	env_file:
		- beta.env
	```
	- Under services->db, add the port the database will use.
	```
	ports:
		- "3306:3306"
	```
### Booting up Rogueserver
- First, compile the code with
```
go build .
```
- Then run the command
```
docker build ./ -t rogueserver
```
- Finally, run the command with the Docker file you just configured!
```
docker-compose -f docker-compose.Example.yml up -d
```
### Connecting your PokeRogue to RogueServer
- Find .env.development in your PokeRogue repo and update it with the following changes
	- To access PokeRogue's account features, you need to set VITE_BYPASS_LOGIN to 0 here too
	- Update VITE_SERVER_URL with the correct machine if your server is running on a different machine than your development machine. 
- In utils.ts, around lines 280-300, remove the Secure headers from the document.cookie variables. For example:
```
document.cookie = `${cName}=${cValue};Secure;SameSite=Strict;Domain=${window.location.hostname};Path=/;Expires=${expiration.toUTCString()}`;
```
should be changed to
```
document.cookie = `${cName}=${cValue};SameSite=Strict;Domain=${window.location.hostname};Path=/;Expires=${expiration.toUTCString()}`;
``` 

The docker compose file should automatically implement a container with mariadb with an empty database and the default user and password combo of pokerogue:pokerogue

# If you are on Windows
You will need to allow the port youre running the API (8001) on and port 8000 to accept inbound connections through the [Windows Advanced Firewall](https://www.youtube.com/watch?v=9llH5_CON-Y).

# If you are on Linux
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

### Contributors
- Instructions by Opaquer


