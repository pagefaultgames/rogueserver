# rogueserver
Backend Golang server & API for [PokéRogue](https://github.com/pagefaultgames/pokerogue).

# Building
There are 2 main methods to host a local instance of the server: via Docker and as a local install.

## Docker
### Requirements:
- Docker Desktop (downloadable [from their website](https://www.docker.com/products/docker-desktop/)).

Use the [sample docker-compose file](./docker-compose.Development.yml) to set up a docker container to run the server.
On initialization, this will create an empty `mariadb` database with the default username/password combo of pokerogue:pokerogue.

# Self Hosting outside of Docker:
## Required Tools:
- Golang 1.22 or higher (downloadable [here](https://go.dev/dl/))
- Node: **18.3.0** or higher
- npm: [how to install](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)
- Both this repository and the [main repo](https://github.com/pagefaultgames/pokerogue) cloned on your device

## Building
Running the server requires only the `rogueserver` executable (compiled from this repo).
Run the following code from the repository root to create and run it[^1]:
```bash
go build .
./rogueserver --debug --dbuser yourusername --dbpass yourpassword
```
(If on windows, replace `rogueserver` with `rogueserver.exe`.)

Now, go to the main repo root and run `npm run start` to boot it up. With some luck, the frontend should connect to the local backend and run smoothly!

[^1]: After doing this, you shouldn't have to re-build it again unless making changes to backend code.

### Hosting for other computers
Now, if you want to access your local server from _other_ machines using localhost, you will need to configure your device's firewalls to allow inbound connections for the ports running the API and server (8000 & 8001).
An example to allow incoming connections using UFW on Linux:
```bash
sudo ufw allow 8000,8001/tcp
```

This should allow you to reach the game from other computers on the same network.

## Tying to a Domain

If you want to tie the local instance to a _domain_ and make it publicly accessible, there are a few extra steps to be done.

**This is FULLY OPTIONAL.** The first 2 steps should be enough for most users merely wanting to test stuff out.

[Caddy](https://caddyserver.com/docs/install) is recommended for use as a reverse proxy.
After installing it, set up a config file like so:

```
pokerogue.exampledomain.com {
	reverse_proxy localhost:8000
}
pokeapi.exampledomain.com {
	reverse_proxy localhost:8001
}
```
(Replace the URLs with whatever domain name you want to tie the server to.)
Then, set up caddy as a service [as shown here](https://caddyserver.com/docs/running).

Once this is good to go, take your API url (https://pokeapi.exampledomain.com) and paste it into **.env.development** inside the main repo, replacing the prior `0.0.0.0:8001` address.

Make sure that ports 8000 and 8001 are both portforwarded on your router.

Enjoy!
