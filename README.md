## Go RS2D

#### RuneScape 317 Private Server written in golang!

Features:
* JAGGRAB protocol support
* Update/cache server support
* Pluggable player configuration source
* Configurable RSA encryption for security (requries modifying the client)



Currently the client can connect, download all required resources (and additional files) and send a login request.

To return the correct login response a PlayerStore needs to be implemented and assigned to Player.PlayerStore...

From here, everything else needs doing!