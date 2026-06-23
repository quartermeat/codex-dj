# Codex DJ

`codex-dj` plays short local generated WAV sounds. It does not open a browser,
does not depend on YouTube or Spotify, and does not need an API key.

On Windows it uses the built-in .NET `System.Media.SoundPlayer`, so the sound
format is deliberately plain WAV.

## Usage

```powershell
go run ./cmd/codex-dj
go run ./cmd/codex-dj play doit
go run ./cmd/codex-dj success
go run ./cmd/codex-dj list
go run ./cmd/codex-dj render doit .\doit.wav
```

Build:

```powershell
go build -o codex-dj.exe ./cmd/codex-dj
```

After building, run the wrapper:

```powershell
.\Codex-DJ.bat
.\Codex-DJ.bat success
.\Codex-DJ.bat reload
.\Codex-DJ.bat what
.\Codex-DJ.bat loki
.\Codex-DJ.bat glow
```

## Sounds

- `doit`
- `success`
- `error`
- `reload`
- `what`
- `loki`
- `glow`

## Codex Sound Protocol

Use these short sounds as a local interaction language:

- `doit`: action acknowledged; Codex is about to execute the requested move.
- `what`: surprise, confusion, or "that result is not what I expected."
- `loki`: knowing mischief; the same shape as `what`, but less startled and more sly.
- `success`: task completed or a smoke test passed.
- `error`: command failed, blocker hit, or user attention needed.
- `reload`: Bitwig/plugin reload or restart handoff is needed.

Do not describe these as music playback. They are local generated WAV cues.

## State

Generated sounds are cached at:

```text
%APPDATA%\CodexTools\codex-dj\sounds
```

The files are generated from Go code and can be deleted at any time.
