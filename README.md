# toller.link

Ein simpler Bookmark-Manager. Geschrieben, weil ich keine Lust hatte, einen anderen zu installieren und mal Go und Svelte angucken wollte.

## Features

- Hinzufügen von neuen Links per API
- Hintergrundprozess, der automatisch Titel und Inhalt der Webseite läd
- Alles landet in RediSearch und ist komplett durchsuchbar
- Frontend, das die Suchergebnisse anzeigt

## Setup

Aktuell gibt es keinen Release-Build, den man sich einfach installieren kann. Für den unwahrscheinlichen Fall, dass man das zum weiterentwickeln einrichten will, hier eine kleine Anleitung:

Folgendes installieren:

- Go
- NodeJS
- Redis mit RediSearch-Modul (z.B. per Docker: `docker run -p 127.0.0.1:6379:6379 redislabs/redisearch:latest`)

Backend ausführen: `go run main.go`, Frontend ausführen: `cd frontend && npm run dev`

## Warnung

Das hier ist in wenigen Stunden zusammengeklöppelt, als Lernübung in ein paar freien Minuten zwischendurch. Der Code ist furchtbar. Vielleicht mache ich ihn irgendwann hübsch.


## Lizenz

MIT
