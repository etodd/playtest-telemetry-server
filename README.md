# Playtest Telemetry Server

A Go server for collecting playtest telemetry.

## SDKs

- [Godot](https://github.com/etodd/playtest-telemetry-godot)

## Setup

1. Create a linux server with a public IP address. Make sure ports 80 and 443 are publicly accessible on it. Give it a domain name.

2. Install Docker and curl.

3. Run these commands, inserting your desired credentials:
	```
	curl -o compose.yaml https://raw.githubusercontent.com/etodd/playtest-telemetry-server/main/compose.yaml
    USERNAME=<username> PASSWORD=<password> API_KEY=<key> DOMAIN=<example.com> docker-compose up -d
	```
