# GoPipe
Embed YouTube videos on Telegram, Discord and more!

## How to use:
Replace `www.youtube.com` or `youtu.be` with `y.outube.duckdns.org` to fix embeds for short videos.

https://github.com/birabittoh/FixYouTube-legacy/assets/26506860/e1ad5397-41c8-4073-9b3e-598c66241255

### Advanced usage
Some services require video previews to be smaller than a certain file size. This app gets the best-looking format by default, so it's not uncommon to have previews that are too big, thus being ignored by crawlers.

Enter the `/{videoID}/{formatID}` endpoint.

You can change `formatID` to cycle through the available formats for a given video. These formats are also filtered based on file size, in a way that makes them viable Telegram previews.

In short, if `/{videoID}` does not generate a preview, you can try `/{videoID}/1`, which should be the best-looking format within Telegram's filesize bounds.

The next values of `formatID` (2 to infinity) are even smaller formats within those bounds.

If the video is too long, there might not be small enough formats. In that case, the app returns error 500.

## Instructions

First of all, you should create a `.env` file:
```
cp .env.example .env
```

### Docker with reverse proxy
Copy the template config file and make your adjustments. My configuration is based on [DuckDNS](http://duckdns.org/) but you can use whatever provider you find [here](https://docs.linuxserver.io/general/swag#docker-compose).

```
cp docker/swag.env.example docker/swag.env
nano docker/swag.env
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
