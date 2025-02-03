# 🧊 IcyProxy

A simple reverse proxy that muxes ICY metadata into an audio stream.

## Why?

Some webradios don't expose the current song metadata in the audio stream,
relying instead on a separate API. This prevents third party players from
displaying the information. IcyProxy acts as a man-in-the-middle proxy that
fetches both the audio stream and the metadata, and muxes them together.

```
┌──────────────────────┐
│ Audio stream without │
│    ICY metadata      │───┐                  ┌───────────────────┐
└──────────────────────┘   │   ┌──────────┐   │ Audio stream with │
                           ├──>│ IcyProxy │──>│   ICY metadata    │
     ┌─────────────────┐   │   └──────────┘   └───────────────────┘
     │ Webradio custom │   │
     │  metadata API   │───┘
     │  (HTTP/JSON)    │
     └─────────────────┘
```

## Installation

### Using Go

Run `go build -o . ./cmd/icyproxy`

The `icyproxy` binary will be in the root directory.

### With Nix

Run `nix build`

The `icyproxy` binary will be in `result/bin`.

## Usage

IcyProxy relies on a JSON configuration file to know how to fetch the audio and
metadata for a given station. Several stations can be configured on a same
server. There is an example config file in
[examples/sources.json](examples/sources.json).

You can then simply run `icyproxy sources.json` to start the server.

Run `icyproxy -help` for more command line options.
