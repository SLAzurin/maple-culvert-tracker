# `/parse-images` Command Implementation

## User Guide

### Quick Start

1. **Enable Developer Mode** (one-time setup):
   - Discord → User Settings → Advanced → Developer Mode → Toggle ON
   - This enables "Copy Message ID" and "Copy Channel ID" right-click options

2. **Find your IDs**:
   - Channel ID: Right-click channel → Copy Channel ID
   - Message ID: Right-click message with images → Copy Message ID

3. **Run command**:
   ```
   /parse-images channel-id:1234567890 message-id-1:9876543210
   ```
   - Up to 5 message IDs supported (message-id-2 through message-id-5)
   - Accepts raw IDs or Discord mentions: `<#1234567890>`

4. **Get results**:
   - JSON file (`gpq_scores.json`) with character scores in leaderboard order
   - If scores are out of order: warning appended to message (but JSON still attached)
   - If character names don't match active roster: `unmatched.txt` ASCII table with names and parsed scores sent instead

### Requirements

- Bot has `View Channel` + `Read Message History` + `Message Content Intent` enabled
- Images are PNG/JPEG in "small" GPQ table style (pre-cropped leaderboard, ~447×413px)
- All parsed character names must exist in the tracked active guild

### Troubleshooting

| Problem | Solution |
|---------|----------|
| `Copy Message ID` missing | Enable Developer Mode (see step 1) |
| `No image attachments` | Attach PNG/JPEG to the message |
| `Names don't match` | Track missing characters first via `/track-character` |
| ⚠️ Order warning | Images may be out of sequence; inspect JSON and retry |

---

## Overview

Discord slash command that parses GPQ score table images and outputs character scores as JSON. Uses pixel-template-matching OCR (no Tesseract), embedded glyph templates, and parallelized image processing.

## Files

- `internal/commands/parseImages.go` — Handler, validation, JSON marshaling
- `internal/commands/helpers/gpq_ocr.go` — Image decode pipeline, name reconciliation
- `internal/commands/helpers/gpq_ocr_test.go` — 12 provided tests (all passing)
- `internal/commands/helpers/gpq_ocr_glyph_test.go` — Direct pixel-decode verification
- `internal/commands/helpers/font/` — 225 Arial 12px no-AA glyph templates (embedded via `//go:embed`)

## Key Design Decisions

### Pixel Matching (No OCR)
- Ported lossless path from Python `gpq.py` + `font_match.py`
- Binarizes game UI text (`#FFFFFF` and `#B3B3B3` → ink) and crops name/score columns
- Template sliding-window matching with `d_in + d_out` distance; normalizes by glyph area
- Tie-break toward widest glyph (prevents 'i' matching 'n' edge)
- Tolerance: 0.12 (12% pixel mismatch)

### Name Reconciliation
- Folds names (NFD lowercase, drop combining marks) for accent-insensitive matching
- Three tiers: exact prefix (1.0), folded prefix (0.97), Ratcliff/Obershelp fuzzy ratio (threshold 0.7)
- Unifies `l` ↔ `i` (identical glyphs in Arial bitmap font)
- Preserves accent-intact matches over plain-ASCII (e.g., "Mïnäh" beats "Minah")

### Order Preservation
- Returns `[]ScoreEntry` (not `map`) to preserve top-to-bottom row order
- Merges multiple images in download order; duplicate names keep first-seen position, scores overwrite
- Custom JSON marshaler (`marshalOrderedScores`) emits 4-space indentation without sorting keys

### Non-Fatal Validation
- After parsing: checks if scores are descending (non-increasing)
- **If violated**: appends warning with offending pair to Discord message but still attaches JSON
- Catches likely OCR ordering issues without blocking delivery

### Font Coverage
- Shipped 112 templates (digits, ASCII, Latin-1 accents)
- Added 113 more (full Latin Extended-A accents: Č č Š Ž ž É È Ñ ñ etc.)
- Total: **225 templates**
- Generated via Arial 12px no-antialiasing pipeline (verified pixel-perfect vs. shipped set)

## Command Parameters

```
/parse-images 
  channel-id:<channel>      (required; accepts raw ID or <#ID>)
  message-id-1:<id>        (required)
  message-id-2..5:<id>     (optional; up to 5 messages)
```

## Discord Bot Setup

- Requires `MESSAGE_CONTENT` privileged intent (attachment data hidden without it)
- Bot needs `VIEW_CHANNEL` + `READ_MESSAGE_HISTORY` in target channel
- Enabled in `create_bot_session.go`: `discordgo.IntentsGuilds | discordgo.IntentMessageContent`

## Output

- JSON: `{"name": score, ...}` in **row order**, 4-space indent
- Attachment: `gpq_scores.json`
- If unmatched names (not in active guild): sends an `unmatched.txt` ASCII table with names and parsed scores, command fails
- If order violation: warning appended to message; JSON still sent

## Testing

- `TestParseSmallImageAgainstProvided`: 12 provided images (1–10 with scores, 11–12 empty)
  - Verifies exact score match AND row-order preservation vs. file key order
- `TestNewGlyphsDecodedFromPixels`: Confirms `è` and `š` templates decode directly (empty roster)

## Processing Flow

1. Parse `channel-id` and collect `message-id-1..5`
2. Fetch messages via `ChannelMessage()` (requires MESSAGE_CONTENT intent)
3. Collect image URLs from attachments (ContentType or `.png`/`.jpg`/`.jpeg`/`.gif`/`.webp` suffix)
4. **Parallel download** each image into memory (no disk writes)
5. **Parallel parse** each image via `ParseSmallImage(imgData, memberNames, font)`
   - Binarize → crop name/score columns → detect rows → decode via template matching → reconcile names
6. Merge results: preserve image order, keep first-seen position for duplicate names
7. Validate: all names in active guild; scores in descending order (warn if not)
8. Output: ordered JSON attachment + Discord message

## Limitations & Future

- Only "small" style images (pre-cropped table, ~447×413px). Full screenshot support requires additional `binarizeCrop` offsets
- Single country/font (Arial). Multi-language requires additional glyph sets
- Image quality: noisy/compressed images may increase font mismatch distance; retry with clearer images
