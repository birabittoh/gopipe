# GoPipe
Embed YouTube videos on Telegram, Discord and more!

## How to use:
Replace `www.youtube.com` or `youtu.be` with `y.outube.duckdns.org` to fix embeds for short videos.

https://github.com/birabittoh/FixYouTube-legacy/assets/26506860/e1ad5397-41c8-4073-9b3e-598c66241255

## Instructions

First of all, you should set the API key to whatever you want:
```
cp .env.example .env
nano .env
```

### Docker with reverse proxy
Copy the template config file and make your adjustments. My configuration is based on [DuckDNS](http://duckdns.org/) but you can use whatever provider you find [here](https://docs.linuxserver.io/general/swag#docker-compose).

```
cp docker/swag.env.example docker/swag.env
nano docker/swag.env
```

Finally: `docker-compose up -d`.

### Docker without reverse proxy
Just run:
```
docker-compose -f docker-compose.simple.yaml up -d
```

## Test and debug locally
```
go test -v ./...
go run .
```
