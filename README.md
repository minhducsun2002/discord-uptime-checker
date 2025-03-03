# Discord Uptime Checker
A Prometheus exporter that checks whether your Discord bot is alive by pinging it regularly.

## How it works
- At intervals (configured timeout or 15 seconds, whichever is greater), the exporter will ping your bot with a pre-defined keyword.
  - If your bot's user ID is `123` and the keyword is `check`, the message would be `<@123> check`
- Your bot must respond before configured timeout by replying to that message using the same keyword e.g. `<@exporter-user-id> check`.
  - You have to enable `GUILD` and `GUILD_MESSAGES` intent on your bot to receive the message
    (the `MESSAGE_CREATE` event).
  - You **do not** need the `MESSAGE_CONTENT` intent, as the check message mentions your bot,
    so its content will always be provided.
  - If you customize `allowed_mentions`, ensure that the reply message mentions the exporter. 

## Requirement
- A Discord bot user. The exporter will run using this user. Note down its token to use later.

## Running up
- Write a config file, see [sample `config.yaml`](./sample.yaml)
- `docker run -e DISCORD_TOKEN=your_bot_token -p 8080:8080 -v /path/to/config.yaml:/config.yaml:ro ghcr.io/minhducsun2002/discord-uptime-checker:master`
