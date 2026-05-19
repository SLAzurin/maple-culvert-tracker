# Discord Embed Character Limits & UTF-16 Counting

## The Problem

Discord's API enforces character limits on embed fields:

| Field | Documented Limit | Confirmed Limit | Notes |
|-------|-----------------|-----------------|-------|
| Title | 256 characters | 256 runes | - |
| Description | 4096 characters | **4096 runes** | Client visually truncates ~3300 rendered chars |
| Footer text | 2048 characters | 2048 runes | - |
| Total (all fields combined) | 6000 characters | - | - |

> **Confirmed (May 2026):** Sending a 5400-rune description returned HTTP 400 with `BASE_TYPE_MAX_LENGTH: Must be 4096 or fewer in length.` The API limit is **4096 runes** (Go's `len([]rune(s))`).

> **Client rendering:** Even within the 4096 rune limit, the Discord client visually truncates the display at around 3300–3400 rendered characters. This appears to be a client-side UI limit, not an API limit. Messages sent within the 4096 rune API limit are accepted and stored correctly, but the client may not show all of them.

**Discord counts characters using UTF-16 code units**, matching JavaScript's `String.length`. This is important because Go's `len()` counts bytes (UTF-8) and `utf8.RuneCountInString()` counts Unicode codepoints (runes) — neither matches Discord's counting.

## Why It Matters

Most ASCII characters are 1 unit in all encodings. The difference shows up with emoji:

| Character | Go `len()` (bytes) | Go runes | UTF-16 code units (Discord) |
|-----------|--------------------|---------|-----------------------------|
| `A`       | 1                  | 1       | 1                           |
| `→` (U+2192) | 3             | 1       | 1                           |
| `⭐` (U+2B50) | 3             | 1       | 1                           |
| `🥇` (U+1F947) | 4            | 1       | **2**                       |
| `🏆` (U+1F3C6) | 4            | 1       | **2**                       |
| `📈` (U+1F4C8) | 4            | 1       | **2**                       |
| `💀` (U+1F480) | 4            | 1       | **2**                       |

Any Unicode codepoint at or above `U+10000` requires a **surrogate pair** in UTF-16, counting as **2** units instead of 1. Using Go's rune count will undercount these characters, potentially exceeding Discord's limits.

## The Fix

Use `unicode/utf16` to count characters the same way Discord does:

```go
import "unicode/utf16"

// discordLen counts UTF-16 code units, matching Discord's character counting
func discordLen(s string) int {
    return len(utf16.Encode([]rune(s)))
}

// truncateDiscord truncates a string to fit within a UTF-16 code unit limit
func truncateDiscord(s string, maxLen int) string {
    runes := []rune(s)
    count := 0
    for i, r := range runes {
        units := 1
        if r >= 0x10000 {
            units = 2
        }
        if count+units > maxLen {
            return string(runes[:i])
        }
        count += units
    }
    return s
}
```

## Where This Is Used

- `cmd/culvert_score_monthly_improvements/` — Monthly improvements Discord embed
