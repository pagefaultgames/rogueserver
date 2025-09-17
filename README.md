<!--
SPDX-FileCopyrightText: 2024-2025 Pagefault Games

SPDX-License-Identifier: CC-BY-NC-SA-4.0
-->

# rogueserver

## Table of Contents

- [Quickstart (Linux, Podman/Docker)](#quickstart-linuxwsl-podmandocker)
- [Quickstart (Windows, Podman)](#quickstart-windows-podmandocker)
- [Running without Podman/Docker](#running-without-podmandocker)
- [Self Hosting](#self-hosting)
- [Podman/Docker Mini Primer](#podmandocker-mini-primer)

## Quickstart (Linux/WSL, Podman/Docker)

These are the exact steps to get this repo working from a fresh clone on a typical Linux system using Podman (a Docker alternative). This is the recommended way for development and local testing.
Note that you can also use Docker as a drop-in replacement for podman, though the steps are a bit different.

For hosting the server for purposes outside of aiding in the development of PokéRogue, see [Self Hosting outside of Docker](#self-hosting-outside-of-docker)

#### 1. Install Required Packages

You will need the following packages:

- **Podman** ([install guide](https://podman.io/getting-started/installation))
- **Podman Compose** ([install guide](https://github.com/containers/podman-compose#installation))

You do not need Go (Golang) installed on your host system for this setup—the container image will handle all Go builds internally.

Install dependencies using your system's package manager:

**Using apt (Debian/Ubuntu and derivatives):**
```sh
sudo apt update
sudo apt install -y podman podman-compose
```

**Using dnf (Fedora and derivatives):**
```sh
sudo dnf install -y podman podman-compose
```

If your distribution does not provide podman-compose as a package, you can install it with pipx or pip:

**First, install [pipx](https://pipx.pypa.io/stable/):**
```sh
# Using apt
sudo apt install pipx
# Or using dnf
sudo dnf install pipx
# Or, with pip (if not available as a package):
python3 -m pip install --user pipx
python3 -m pipx ensurepath
```

**Then, install podman-compose:**
```sh
# Using pipx (recommended for isolated installs)
pipx install podman-compose
# Or, using pip (installs to user site-packages)
pip install --user podman-compose
```

#### 2. Clone the Repository

```sh
git clone <this-repo-url>
cd rogueserver
```

#### 3. Build the Go Server Image

```sh
podman build -t rogueserver:dev .
```

#### 4. Start the Development Environment

```sh
podman-compose -f docker-compose.Development.yml up
```
This will start both the MariaDB database and the Go server. The first startup will initialize the database schema automatically (dev only).

#### 5. Access the Server

- The API will be available at: http://localhost:8001
- The game server (if running) will be at: http://localhost:8000

#### 6. (Optional) Stopping and Cleaning Up

To stop the services:
```sh
podman-compose -f docker-compose.Development.yml down
```
To remove all containers and volumes:
```sh
podman-compose -f docker-compose.Development.yml down -v
```

---


## Quickstart (Windows, Podman/Docker)

These are the steps to get this repo working from a fresh clone on Windows using Podman Desktop and podman-compose. This is the recommended way for development and local testing on Windows.
Docker can also be used instead of Podman.

#### 1. Install Required Software

- **Podman Desktop** ([download here](https://podman-desktop.io/))
- **podman-compose** (included with Podman Desktop, or install via pipx/pip if needed)

#### 2. Clone the Repository

Open PowerShell or Command Prompt and run:
```powershell
git clone <this-repo-url>
cd rogueserver
```

#### 3. Build the Go Server Image

Open a terminal in the project directory and run:
```powershell
podman build -t rogueserver:dev .
```

#### 4. Start the Development Environment

```powershell
podman-compose -f docker-compose.Development.yml up
```
This will start both the MariaDB database and the Go server. The first startup will initialize the database schema automatically (dev only).

#### 5. Access the Server

- The API will be available at: http://localhost:8001
- The game server (if running) will be at: http://localhost:8000

#### 6. (Optional) Stopping and Cleaning Up

To stop the services:
```powershell
podman-compose -f docker-compose.Development.yml down
```
To remove all containers and volumes:
```powershell
podman-compose -f docker-compose.Development.yml down -v
```


## Running without Podman/Docker
### Required Tools:
- Golang
- mariadb [how to install](https://mariadb.org/download/)

You must install and configure MariaDB yourself when not using Podman/Docker.

#### Linux / WSL
```sh
sudo apt update
sudo apt install mariadb-server
sudo systemctl start mariadb
sudo systemctl enable mariadb
```

Then, secure your installation and set up the database and user:
```sh
sudo mysql_secure_installation
sudo mysql -u root -p
```
In the MariaDB shell, run:
```sql
CREATE DATABASE pokerogue;
CREATE USER 'pokerogue'@'localhost' IDENTIFIED BY 'pokerogue';
GRANT ALL PRIVILEGES ON pokerogue.* TO 'pokerogue'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

#### Windows
- Download and install MariaDB from the [official site](https://mariadb.org/download/).
- Use the MariaDB command line or a GUI tool to create a database named `pokerogue` and a user `pokerogue` with password `pokerogue`, and grant that user all privileges on the database.

You can use different credentials, but you must update your server's configuration/flags accordingly.

---

### Buliding and Running
Now that all of the files and dependencies are configured, the next step is to build the server and then run it.

##### Windows
```powershell
cd C:\api\server\location\
go build . -tags=devsetup
.\rogueserver.exe
```
You may need to adjust environment variables to adjust things like the database user and password, and setup debug mode.
See [rogueserver.go](./rogueserver.go)

### Linux
In whatever shell you prefer, run the following:
```sh
go build . -tags=devsetup
dbuser=pokerogue dbpass=pokerogue debug=1 ./rogueserver &
```


## Self Hosting
You can host your own rogueserver and allow other machines to connect to it.

You will need npm: [how to install](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

The first steps are the same as the above section, but have an additional step to expose the server to outside connections.

### Windows

##### The following step is only necessary to host the server for external use, and is unnecessary for development
Then in another run this the first time then run `npm run start` from the rogueserver location from then on:
```powershell
powershell -ep bypass
cd C:\server\location\
npm install
npm run start
```
You will need to allow the port you're running the API (8001) on and port 8000 to accept inbound connections through the [Windows Advanced Firewall](https://www.youtube.com/watch?v=9llH5_CON-Y).

### Linux
In whatever shell you prefer, run the following:
```sh
go build -tags=devsetup
dbuser=pokerogue dbpass=pokerogue debug=1 ./rogueserver &
```
you can replace the dbuser=..., etc., with whatever authentication you chose for mariadb. These just set the environment variables for the running process. Alternatively, load them into their own environment variables before running `./rogueserver`.
`debug` is obviously optional.

If you have a firewall running such as ufw on your linux machine, make sure to allow inbound connections on the ports youre running the API and the pokerogue server (8000,8001).
An example to allow incoming connections using UFW:
```
sudo ufw allow 8000,8001/tcp
```

This should allow you to reach the game from other computers on the same network. 

### Tying to a Domain

If you want to tie it to a domain and make it publicly accessible, there is some extra work to be done.

It is recommended to setup caddy and use it as a reverse proxy. 
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

Once this is good to go, take your API url (https://pokeapi.exampledomain.com) and paste it in the apprporiate `.env` file for PokéRogue in place of the previous address.

Make sure that both 8000 and 8001 are portforwarded on your router.

Test that the server's game and game authentication works from other machines both in and outside of the network. Once this is complete, enjoy!


## Podman/Docker Mini Primer

### What are Images, Containers, and Compose?

- **Image:** A snapshot or template of an application and its environment (e.g., `mariadb:11`, `rogueserver:dev`). You build or pull images, and then run containers from them. Think of an image as a "blueprint" for a program. Images can be sourced from an existing provider (like docker.io) or built locally.
- **Container:** A running instance of an image. It’s isolated from your host system, but can be allowed to connect to networks and volumes. Containers are temporary by default, but you can persist data using named volumes.
- **Compose:** A tool and YAML file format (e.g., `docker-compose.Development.yml`) for defining and running multi-container applications. Compose lets you start, stop, and manage groups of containers (services), their networks, and volumes with a single command. The compose file provides instructions for how to run the images and what commands to take.

### Why are the steps above needed?

1. **Build the image:**
	- You only need to do this when the code or Dockerfile changes, or when setting up for the first time. This creates the `rogueserver:dev` image locally.
2. **Create the network:**
	- This is needed once, unless you delete the network. It allows your containers to communicate securely and privately.
3. **Start the environment (compose up):**
	- This launches the containers as defined in your compose file. You do this every time you want to start the server/database.
4. **Stop the environment (compose down):**
	- This stops and removes the containers, but keeps your data (unless you use `-v`).
5. **Rebuilding:**
	- If you change your code or Dockerfile, rebuild the image and restart the environment.

### Which steps do I repeat?

- **Daily development:**
  - Run `podman-compose -f path/to/compose-file.yml up` to start
    - You can pass the `-d` flag to allow it to run in the background
  - Run `podman-compose -f /path/to/compose-file.yml down` to stop
- **If rogueserver changes:**
  - Pull the latest changes from rogueserver via git.
  - Rebuild the image (`podman build -t rogueserver:dev .`), then restart the environment.
- **If you want to reset everything:**
  - See below for clearing out all Podman data.

### How do I clear out all Podman containers, images, and volumes?

**Warning:** This will delete all containers, images, and volumes managed by Podman on your system!

```sh
podman stop --all
podman rm --all
podman rmi --all
podman volume rm --all
```

To only remove the ones for rogueserver, pass explicit names. You can see the images, for example, via
```sh
podman ps
```

To delete everything at once, do
```sh
podman system prune -a
```

This is useful if you want to start completely fresh or reclaim disk space.

---

In short: **Images** are blueprints, **containers** are running apps, and **compose** is how you manage multi-container setups easily. The steps above help you build, run, and manage your development environment in a repeatable and isolated way.