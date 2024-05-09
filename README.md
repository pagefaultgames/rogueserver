# rogueserver

Hosting in Docker
It is advised that you host this in a docker container as it will be much easier to manage. 
There is a sample docker-compose file for setting up a docker container to setup this server.

Self Hosting outside of Docker:
Windows:
Recommended Tools:
Chocolatey (https://chocolatey.org/install)

Required Tools:
Golang
Node: 18.3.0
npm: how to install (https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

Installation:

Once you have all the prerequisites you will need to setup a database. 
There are tons of database services you can choose from:
mysql - getting started (https://dev.mysql.com/doc/mysql-getting-started/en/)
mariadb - getting started (https://mariadb.com/get-started-with-mariadb/)
etc

I went with MySQL. Once the database is setup, make sure that you can authenticate to the database.
After being able to login to the database, create a database/schema called pokeroguedb.
Select it as the default database and then run the sql queries from sqlqueries.sql. You should now be able to see all of the empty tables. 

Edit the following files:
##### rogueserver.go:34 
Change the 'false' after "debug" to 'true'. This will resolve CORS issues that many users have been having while trying to spin up their own servers.

##### rogueserver.go:37
Change the default port if you need to, I set it to 8001.

##### rogueserver.go:39-43
Make sure that the credentials that you tested to login to the database are put in here. 

##### src/utils.ts:224-225 (in pokerogue)
Replace both URLs (one on each line) with the local API server address from rogueserver.go (0.0.0.0:8001) (or whatever port you picked)

Now that all of the files are configured: start up powershell as administrator (or save the following as a ps1):
```
powershell -ep bypass
cd C:\server\location\
go run .
```

Then in another run this the first time, then remove the npm install if you plan on running this as a powershell script in the future:
```
powershell -ep bypass
cd C:\server\location\
npm install
npm run start
```

You will need to allow the port youre running the API (8001) on and port 8000 to accept inbound connections through the windows advanced firewall.

This should allow you to reach the game from other parts on the network. If you want to tie it to a domain like I did, there is some extra work to be done.
I setup caddy and would recommend using it as a reverse proxy. 
caddy installation - (https://caddyserver.com/docs/install)
once its installed setup a config file for caddy:

pokerogue.exampledomain.com {
	reverse_proxy localhost:8000
}
pokeapi.exampledomain.com {
	reverse_proxy localhost:8001
} 

Preferably set up caddy as a service from here: https://caddyserver.com/docs/running

Once this is good to go, take your API url (https://pokeapi.exampledomain.com) and paste it on 
##### src/utils.ts:224-225 
in place of the previous 0.0.0.0:8001 address

Make sure that both 8000 and 8001 are portforwarded on your router.

Test that the server's game and game authentication works from other machines both in and outside of the network. Once this is complete, enjoy!
