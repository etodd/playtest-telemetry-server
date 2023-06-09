# Playtest Telemetry

A Go server for receiving playtest telemetry from Godot.

## Setup

1. Create a linux server with a public IP address. Make sure ports 80 and 443 are publicly accessible on it. Give it a domain name.

2. Install Docker.

3. Run this command:
	```
    USERNAME=<username> PASSWORD=<password> API_KEY=<key> DOMAIN=<example.com> docker-compose up -d
	```