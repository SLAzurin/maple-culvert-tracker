# Maplayer One secondary characters

Allow a tracked character to have zero or more secondary characters. When Maplayer One screenshots show a secondary, `/parse-images` should use the tracked character’s current name in the score export. The export remains compatible with `/submit-scores` and `/submit-scores-from-attachment`.

## Database

1. Add `4_secondary_characters.up.sql` and `4_secondary_characters.down.sql` under `db_migrations`. Create `secondary_characters` with a `character_id` foreign key to `characters(id)` and a `maple_character_name` for the secondary. Index `character_id` and enforce global, case-insensitive uniqueness for secondary names. Use `ON DELETE CASCADE`. Linking by ID preserves secondaries when a tracked character is renamed.
2. Reject a secondary name that is already tracked or that OCR cannot distinguish from a name belonging to another tracked character. Use the same normalization as OCR for the latter check. Check again when loading the active OCR roster because tracking and renaming can introduce a conflict after a secondary is added. Report the conflicting names and withhold the score export if the mapping is ambiguous.

## Discord commands

| Command | Required options | Behavior |
| --- | --- | --- |
| `/add-secondary-character` | `character-name`, `secondary-character-name` | Resolve the parent case-insensitively from `characters`, including inactive parents. Trim input, verify the secondary with `FetchCharacterData` in the configured MapleStory region, and store the name returned by Nexon. The optional `skip-name-check` flag bypasses the Nexon lookup and keeps the supplied spelling. Duplicate and OCR conflict checks still apply. |
| `/remove-secondary-character` | `secondary-character-name` | Find the globally unique secondary case-insensitively and remove it. Confirm the parent it was linked to, or report that the secondary was not found. |
| `/list-secondary-characters` | None | Use `GetActiveCharacters` to select parents. Upload `secondary_characters.json`, keyed by each active parent’s current `maple_character_name`. Sort the secondary names within each array. Omit parents with no secondaries and return `{}` if none qualify. |

Register the commands and handlers alongside the existing roster commands. Restrict add and remove to roster managers because mappings affect who receives a score.

Example list attachment with two secondaries for `MainA`:

```json
{
  "MainA": ["AltA", "AltB"]
}
```

## `/parse-images`

1. Load active characters through `GetActiveCharacters`. Fetch their secondary mappings in one query and build a shared OCR roster containing active and secondary names. Keep a lookup from each secondary to the active parent’s current name.
2. Resolve OCR matches to active parent names before merging entries, checking for unmatched names, and building `gpq_scores.json`. Retain the original matched name and score until duplicate checks finish. If an OCR match ties between different parents, report the row for review instead of assigning its score.
3. Require each tracked character ID to appear at most once across all matched rows. Warn in the interaction response and withhold `gpq_scores.json` whenever two rows share a parent ID. This covers a main and secondary, two secondaries, and repeated rows under the same name. Attach the matched names and scores, even when the scores agree. Count zero-score rows only when their decoded name exactly matches after OCR normalization.
4. Keep the unmatched-name report and score-order warning.

## Verification and rollout

- Exercise the PostgreSQL migration up and down. Verify that parent renames preserve mappings, parent deletion removes them, and duplicate secondary names cannot be inserted.
- Check the command permissions, add and remove behavior, and active-only JSON output. Include inactive parents, parents without secondaries, and an empty export.
- Check `/parse-images` with a secondary by itself and with repeated parent IDs within one screenshot and across attachments. Verify the warning and withheld JSON for both equal and different scores, and run the existing OCR fixtures and Go test suite.
- Apply the migration through bot startup before command registration. To roll back the application, deploy the prior bot version with the additive table still present. Running the down migration drops the stored mappings.
