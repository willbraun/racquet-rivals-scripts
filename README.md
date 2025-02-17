# Overview

This repo is for a Go script that gathers results of pro tennis tournaments for [racquetrivals.com](https://www.racquetrivals.com). It makes requests to the ATP and WTA websites for each draw, scrapes the player names, seeds, match scores, and their position in the draw as the tournament progresses, and loads new data to the Pocketbase backend for the application. The script is executed by a cron job on the Digital Ocean droplet hosting the app.

Why didn't I use an API? I would have loved to, but I couldn't find a free/cheap one that met my needs. If you know of one, let me know!

For more information on Racquet Rivals, see this [repo](https://github.com/willbraun/tennis-bracket-frontend).

# Technologies Used

Go, Goquery, Pocketbase, Linux cron job on Digital Ocean Droplet, Bright Data Web Unlocker
