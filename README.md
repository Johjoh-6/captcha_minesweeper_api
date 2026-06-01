![TEST](https://github.com/Johjoh-6/captcha_minesweeper_api/actions/workflows/ci.yml/badge.svg)
![DEPLOY](https://github.com/Johjoh-6/captcha_minesweeper_api/actions/workflows/cd.yml/badge.svg)

# Captcha Sweeper

**A backend API for the Captcha Sweeper App**, built with Go, PostgreSQL, and modern tooling (sqlc, pgx, Goose).

It's a simple API that allows you based on session to get a captcha game. The game is Minesweeper with different difficulty levels.

The SDK is available for the frontend to integrate with the API.
[SDK](https://github.com/Johjoh-6/captcha-minesweeper-sdk)

If the session captcha is set as `bot`, the API will increase the difficulty level automatically. It will increase the difficulty level by 1 every time the bot makes a request with never a chance of being detected as a human.
> This is for the future updates

# Motivation

I wanted a home solution for a captcha that could be integrated into any web application. Since most of the traditional captcha use images recognition, I wanted to use a different approach. The idea to use minesweeper as a captcha game come from the popular game Minesweeper (i use to play a lot during my childhood). But I wanted to build a backend API for it, so I could integrate it into any web application. There is also a frontend SDK available to integrate with the API.

# Usage 

After setting your Postgres database, run:

```bash
make migrations/up
```

After just run it with 
```bash
make run
```

## API Documentation

> Add the version in the front of the endpoint, currently `v1`.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/status` | Returns the status of the API |
| GET    | `/metrics` | Returns the metrics of the API |
| GET    | `/captcha` | Returns the current captcha game, if reload is needed |
| POST   | `/captcha/new` | Creates a new captcha game |
| POST   | `/captcha/move` | Makes a move in the captcha game |

### GET `/status`

Returns the status of the API.

JSON response:

**200 OK**
```json
{
    "status": "ok",
}
```

> If the API is not available, you will get a **500 Internal Server Error** response.

### GET `/metrics`

Returns the metrics of the API.

Only available with the basic auth. Mostly used for monitoring and debugging. You need to set the `BASIC_AUTH_USER` and `BASIC_AUTH_PASSWORD` environment variables.

> **Important:** The basic auth password is hashed using bcrypt. There is a bug where the `.env` variable who contains a *$* got stripped. For avoid this bug you should double it (e.g. `$$` instead of `$`).

JSON response:

**200 OK**
```json
{
{
  "app_metrics": {
    "captchas_last_hour": 0,
    "captchas_solved": 0,
    "captchas_total": 10,
    "captchas_unsolved": 10,
    "sessions_bots": 1,
    "sessions_non_bots": 0,
    "sessions_total": 1
  },
  "cmdline": [
    "./bin/api"
  ],
  "database": {
    "acquire_count": 2,
    "acquire_duration": "114.852124ms",
    "acquired_conns": 0,
    "canceled_acquire_count": 0,
    "constructing_conns": 0,
    "empty_acquire_count": 1,
    "empty_acquire_wait_time": "114.851541ms",
    "idle_conns": 9,
    "max_conns": 25,
    "max_idle_destroy_count": 0,
    "max_lifetime_destroy_count": 0,
    "new_conns_count": 25,
    "total_conns": 9
  },
  "goroutines": 6,
  "memstats": {
    "Alloc": 1534856,
    "TotalAlloc": 4624920,
    "Sys": 19663112,
    "Lookups": 0,
    "Mallocs": 22222,
    "Frees": 18939,
    "HeapAlloc": 1534856,
    "HeapSys": 11206656,
    "HeapIdle": 7348224,
    "HeapInuse": 3858432,
    "HeapReleased": 5824512,
    "HeapObjects": 3283,
    "StackInuse": 1376256,
    "StackSys": 1376256,
    "MSpanInuse": 145920,
    "MSpanSys": 179520,
    "MCacheInuse": 22960,
    "MCacheSys": 32144,
    "BuckHashSys": 1444327,
    "GCSys": 2977008,
    "OtherSys": 2447201,
    "NextGC": 4194304,
    "LastGC": 1779379424359357000,
    "PauseTotalNs": 413376,
    "PauseNs": [
      79000,
      ...
    ],
    "PauseEnd": [
      1779379424290321000,
      ...
    ],
    "NumGC": 5,
    "NumForcedGC": 0,
    "GCCPUFraction": 0.004927479399894521,
    "EnableGC": true,
    "DebugGC": false,
    "BySize": [
      {
        "Size": 0,
        "Mallocs": 0,
        "Frees": 0
      },
      ...
    ]
  },
  "timestamp": 1779379428,
  "version": "x.x.x"
} 
}
```

**401 Unauthorized**
```json
{
    "Error": "Unauthorized"
}
```

### GET `/captcha`

Returns the current captcha game, if reload is needed.
It stores the captcha ID in the session, so you don't need to pass it in the request body.
> Actually only one captcha is active per session. Return the active captcha if one exists, otherwise return a 404 Not Found response.

**200 OK**
```json
{
  "captcha_id": "adfc5e7c-091b-439e-ae9b-ec3bf2fbbc04",
  "solved": false,
  "game_over": false,
  "mine_hit": false,
  "grid_size": 6,
  "grid_revealed": [
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ]
  ],
  "difficulty_level": 2,
  "status": "unsolved"
}
```

**404 Not Found**
```json
{
  "Error": "The requested resource could not be found"
}
```


### POST `/captcha/new`

Creates a new captcha game.

Body:
```json
{
  "difficulty_level": 2
}
```

**201 Created**
```json
{
  "captcha_id": "adfc5e7c-091b-439e-ae9b-ec3bf2fbbc04",
  "solved": false,
  "game_over": false,
  "mine_hit": false,
  "grid_size": 6,
  "grid_revealed": [
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ]
  ],
  "difficulty_level": 2,
  "status": "unsolved"
}
```

**404 Not Found**
```json
{
    "Error": "The requested resource could not be found"
}
```

### POST `/captcha/move`

Moves the captcha game to the next state.

Body:
```json
{
  "row": 1,
  "col": 1
}
```

**200 OK**
```json
{
  "captcha_id": "adfc5e7c-091b-439e-ae9b-ec3bf2fbbc04",
  "solved": false,
  "game_over": true,
  "mine_hit": true,
  "grid_size": 6,
  "grid_revealed": [
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -1,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ],
    [
      -2,
      -2,
      -2,
      -2,
      -2,
      -2
    ]
  ],
  "difficulty_level": 2,
  "status": "mine_hit"
}
```

**404 Not Found**
```json
{
    "Error": "The requested resource could not be found"
}
```

## Logic of the API

You need first to create a new captcha game using the `/captcha/new` endpoint, then you can move it to the next state using the `/captcha/move` endpoint.
Based on the current state, the API will return the appropriate response.

You can reload the captcha using the `/captcha` endpoint, which will return the current captcha game if one exists. (usefull for reload the webpage)


## CI / CD
 Important, we use CI on `pull request` to the `main` branch. This will trigger: 
 	- Go test
	- GoSec
	- Style with go fmt and staticcheck

The CD will be trigger only on `push` to the `main` branch, this will triger the build of the docker image, push to the artifact (GCP), deploy to the service (GCP) and make it available for the web.
