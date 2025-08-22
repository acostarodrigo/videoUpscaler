#!/usr/bin/env bash
set -euo pipefail

# Defaults
SCALE=2              # scale ratio (maps to --scale-ratio)
NOISE=-1             # -1 disables denoise; 0..3 maps to --noise-level
FAST=0               # internal: if 1, we use more threads (non-deterministic)
TIME=0               # -ss start time in seconds
DURATION=""          # -t duration in seconds (empty = full)
FRAME_STEP=1         # extract every Nth frame
FRAME_ONLY=""        # if set, extract only that exact frame index (0-based)
OUTPUT=""            # output file (video) or dir (when --frame is used)
INPUT=""             # input video

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    -i) INPUT="$2"; shift 2 ;;
    -o) OUTPUT="$2"; shift 2 ;;
    -s) SCALE="$2"; shift 2 ;;         # will map to --scale-ratio
    -n) NOISE="$2"; shift 2 ;;         # will map to --noise-level
    --fast) FAST=1; shift ;;
    --time) TIME="$2"; shift 2 ;;
    --duration) DURATION="$2"; shift 2 ;;
    --frame-step) FRAME_STEP="$2"; shift 2 ;;
    --frame) FRAME_ONLY="$2"; shift 2 ;;  # single frame index (0-based)
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$INPUT" || -z "$OUTPUT" ]]; then
  echo "Usage:"
  echo "  $0 -i input.mp4 -o output.mp4 [-s 2] [-n -1|0..3] [--fast] [--time SS] [--duration SS] [--frame-step N]"
  echo "  $0 -i input.mp4 -o /out/dir --frame 1234 [-s 2] [-n -1|0..3] [--fast]"
  echo "Notes:"
  echo "  • In --frame mode, -o must be a directory (PNG frames will be saved there as frame_%08d.png)."
  exit 1
fi

if [[ ! -f "$INPUT" ]]; then
  echo "Input not found: $INPUT" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
FRAMES_IN="$TMP_DIR/frames_in"
FRAMES_OUT="$TMP_DIR/frames_out"
mkdir -p "$FRAMES_IN" "$FRAMES_OUT"

# ---------- [1] Extract frames ----------
echo "[1/3] Extracting frames..."
if [[ -n "$FRAME_ONLY" ]]; then
  # Single exact frame
  PADDED_FRAME=$(printf "%08d" "$FRAME_ONLY")
  ffmpeg -hide_banner -loglevel error -y \
    -i "$INPUT" \
    -vf "select=eq(n\,${FRAME_ONLY})" -vsync vfr \
    "$FRAMES_IN/frame_${PADDED_FRAME}.png"
else
  # Range / stepped extraction
  if [[ -n "$DURATION" ]]; then
    DURATION_ARG=(-t "$DURATION")
  else
    DURATION_ARG=()
  fi
  ffmpeg -hide_banner -loglevel error -y \
    -ss "$TIME" "${DURATION_ARG[@]}" \
    -i "$INPUT" \
    -vf "select=not(mod(n\,$FRAME_STEP))" -vsync vfr \
    "$FRAMES_IN/frame_%08d.png"
fi

# ---------- [2] Waifu2x upscale ----------
echo "[2/3] Running waifu2x upscale..."

# Build waifu2x args using the correct long flags for this build
W2X_ARGS=(
  -i "$FRAMES_IN"
  -o "$FRAMES_OUT"
  --scale-ratio "$SCALE"
  --png-compression 0
  --image-quality 100
  --model-dir /usr/local/share/waifu2x-converter-cpp/models_rgb
  # --verbose
)

# Noise handling and mode
if [[ "$NOISE" -ge 0 ]]; then
  W2X_ARGS+=(--noise-level "$NOISE" -m noise_scale)
else
  W2X_ARGS+=(-m scale)
fi

# Threads: default 1 (deterministic). If --fast, use all cores (faster, may be non-deterministic).
if [[ "$FAST" -eq 1 ]]; then
  W2X_ARGS+=(-j "$(nproc)")
else
  W2X_ARGS+=(-j 1)
fi

waifu2x-converter-cpp "${W2X_ARGS[@]}"

# ---------- [3] Save output ----------
echo "[3/3] Saving..."
if [[ -n "$FRAME_ONLY" ]]; then
  # single-frame mode → OUTPUT must be a directory
  mkdir -p "$OUTPUT"
  cp "$FRAMES_OUT/"*.png "$OUTPUT/"
else
  # re-encode PNG frames to a video; adjust framerate if you need original FPS preservation
  ffmpeg -hide_banner -loglevel error -y \
    -framerate 30 \
    -i "$FRAMES_OUT/frame_%08d.png" \
    -c:v libx264 -pix_fmt yuv420p -preset medium -crf 18 \
    "$OUTPUT"
fi

rm -rf "$TMP_DIR"
echo "Done!"
