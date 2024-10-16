# GoPipe
Embed YouTube videos on Telegram, Discord and more!

## How to use:
Replace `www.youtube.com` or `youtu.be` with `y.outube.duckdns.org` to fix embeds for short videos.

https://github.com/birabittoh/FixYouTube-legacy/assets/26506860/e1ad5397-41c8-4073-9b3e-598c66241255

### Advanced usage
Some services require video previews to be smaller than a certain file size. By default, this app selects the best-looking format that fits these criteria.

Enter the `/{videoID}/{formatID}` endpoint.

The default value of  `formatID` is `1`, but you can increase it to cycle through the available formats for a given video or set it to `0` to select the best-looking format and ignore file size bounds altogether.

If the video is too long, there might not be small enough formats. In that case, the app effectively behaves like `formatID` is set to `0`.

## Instructions

First of all, you should create a `.env` file:
```
cp .env.example .env
```

### Docker with reverse proxy
Copy the template config file and make your adjustments. My configuration is based on [DuckDNS](http://duckdns.org/) but you can use whatever provider you find [here](https://docs.linuxserver.io/general/swag#docker-compose).

```
cp swag/swag.env.example swag/swag.env
nano swag/swag.env
```

Finally: `docker compose up -d`.

### Docker without reverse proxy
Just run:
```
docker compose -f docker-compose.simple.yaml up -d
```

## Test and debug locally
```
go test -v ./...
go run .
```
