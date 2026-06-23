package dj

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const sampleRate = 44100

type sound struct {
	Name  string
	Notes []note
}

type note struct {
	Frequency float64
	Seconds   float64
	Volume    float64
}

var sounds = map[string]sound{
	"doit": {
		Name: "doit",
		Notes: []note{
			{Frequency: 523.25, Seconds: 0.09, Volume: 0.30},
			{Frequency: 659.25, Seconds: 0.09, Volume: 0.30},
			{Frequency: 783.99, Seconds: 0.16, Volume: 0.34},
		},
	},
	"success": {
		Name: "success",
		Notes: []note{
			{Frequency: 659.25, Seconds: 0.08, Volume: 0.28},
			{Frequency: 783.99, Seconds: 0.08, Volume: 0.30},
			{Frequency: 1046.50, Seconds: 0.18, Volume: 0.32},
		},
	},
	"error": {
		Name: "error",
		Notes: []note{
			{Frequency: 220.00, Seconds: 0.12, Volume: 0.30},
			{Frequency: 164.81, Seconds: 0.18, Volume: 0.30},
		},
	},
	"reload": {
		Name: "reload",
		Notes: []note{
			{Frequency: 392.00, Seconds: 0.07, Volume: 0.25},
			{Frequency: 0, Seconds: 0.04, Volume: 0},
			{Frequency: 392.00, Seconds: 0.07, Volume: 0.25},
			{Frequency: 587.33, Seconds: 0.14, Volume: 0.30},
		},
	},
	"what": {
		Name: "what",
		Notes: []note{
			{Frequency: 349.23, Seconds: 0.08, Volume: 0.30},
			{Frequency: 0, Seconds: 0.035, Volume: 0},
			{Frequency: 466.16, Seconds: 0.10, Volume: 0.33},
			{Frequency: 698.46, Seconds: 0.20, Volume: 0.36},
		},
	},
	"glow": {
		Name: "glow",
		Notes: []note{
			{Frequency: 440.00, Seconds: 0.10, Volume: 0.24},
			{Frequency: 554.37, Seconds: 0.10, Volume: 0.25},
			{Frequency: 659.25, Seconds: 0.10, Volume: 0.27},
			{Frequency: 880.00, Seconds: 0.24, Volume: 0.24},
		},
	},
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		args = []string{"play", "doit"}
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	case "list":
		printSounds(stdout)
		return nil
	case "play":
		name := "doit"
		if len(args) > 1 {
			name = args[1]
		}
		path, err := ensureSound(name)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "playing %s\n%s\n", name, path)
		return playWAV(ctx, path)
	case "render":
		if len(args) < 2 || len(args) > 3 {
			return errors.New("usage: codex-dj render <sound> [path]")
		}
		path := ""
		if len(args) == 3 {
			path = args[2]
		}
		rendered, err := renderSound(args[1], path)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, rendered)
		return nil
	default:
		name := args[0]
		path, err := ensureSound(name)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "playing %s\n%s\n", name, path)
		return playWAV(ctx, path)
	}
}

func ensureSound(name string) (string, error) {
	return renderSound(name, "")
}

func renderSound(name, outPath string) (string, error) {
	s, ok := sounds[normalize(name)]
	if !ok {
		return "", fmt.Errorf("unknown sound %q; run codex-dj list", name)
	}
	if outPath == "" {
		outPath = filepath.Join(soundDir(), s.Name+".wav")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
		return "", err
	}
	data := synthesize(s)
	if err := os.WriteFile(outPath, data, 0600); err != nil {
		return "", err
	}
	return outPath, nil
}

func synthesize(s sound) []byte {
	var samples []int16
	for _, n := range s.Notes {
		count := int(n.Seconds * sampleRate)
		for i := 0; i < count; i++ {
			if n.Frequency <= 0 || n.Volume <= 0 {
				samples = append(samples, 0)
				continue
			}
			t := float64(i) / sampleRate
			envelope := 1.0
			attack := int(math.Round(0.008 * float64(sampleRate)))
			release := int(math.Round(0.025 * float64(sampleRate)))
			if attack > 0 && i < attack {
				envelope = float64(i) / float64(attack)
			}
			if release > 0 && count-i < release {
				envelope *= float64(count-i) / float64(release)
			}
			wave := math.Sin(2 * math.Pi * n.Frequency * t)
			samples = append(samples, int16(wave*n.Volume*envelope*math.MaxInt16))
		}
	}
	return wav(samples)
}

func wav(samples []int16) []byte {
	dataSize := uint32(len(samples) * 2)
	riffSize := uint32(36) + dataSize
	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], riffSize)
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], 1)
	binary.LittleEndian.PutUint32(buf[24:28], sampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], sampleRate*2)
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], dataSize)
	offset := 44
	for _, sample := range samples {
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(sample))
		offset += 2
	}
	return buf
}

func playWAV(ctx context.Context, path string) error {
	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(`$p = %q; $player = New-Object System.Media.SoundPlayer $p; $player.PlaySync()`, path)
		return exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", script).Run()
	case "darwin":
		return exec.CommandContext(ctx, "afplay", path).Run()
	default:
		if _, err := exec.LookPath("paplay"); err == nil {
			return exec.CommandContext(ctx, "paplay", path).Run()
		}
		if _, err := exec.LookPath("aplay"); err == nil {
			return exec.CommandContext(ctx, "aplay", path).Run()
		}
		return errors.New("no WAV player found; install paplay, aplay, or use Windows SoundPlayer")
	}
}

func soundDir() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "CodexTools", "codex-dj", "sounds")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".codex-tools", "codex-dj", "sounds")
	}
	return filepath.Join(os.TempDir(), "codex-dj-sounds")
}

func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func printSounds(w io.Writer) {
	keys := make([]string, 0, len(sounds))
	for key := range sounds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintln(w, key)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `codex-dj plays local generated WAV sounds.

Usage:
  codex-dj
  codex-dj play [sound]
  codex-dj <sound>
  codex-dj list
  codex-dj render <sound> [path]

Sounds:
  doit, success, error, reload, what, glow`)
}
