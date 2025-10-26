## This markdown file serves as development notes for the `/submit-scores` discord command

user interaction flow
1. paste scores as attachment in discord msg and copy `message-id` as msg link or msg-id.
2. use `/submit-scores` and use options mandatory `message-id`, optional `date`, optional `overwrite`

That's it.


---

code side validation and execution flow
1. command options must contain following:
    - option message-id must be a message url link or message id from same text channel
    - option date must be a culvert reset date, aka wednesday, in format YYYY-MM-DD, or if it is a new week, auto-default the date for this week, default optional
    - option overwrite must be set to true if values for this date already exist, default false

2. input validation + build scores to submit
    - validate if date is culvert day of the week aka wednesday
    - query if there are scores (limit 1 query select stmt)
        - check if overwrite is enabled or error
    - validate message-id content
        - message must only have 1 attachment
        - attachment must be in file format .txt or .json
        - attachment must be <2mb
    - download message attachment
        - parse json if schema matches `map[string]int` => attachmentMap
    - query all tracked characters in db, left join with scores, default null or score int => trackedCharacterScores: []struct{ MapleCharacterID: int, Score: int64 }
    - loop through trackedCharacterScores -> check if characterName is in map attachmentMap[v]
        - error and break loop if name is not found, suggest use /track-character command
        - assign culvert score from attachmentMap[v] to newMapIsNew[v] if score is nil, and to newMapIsNotNew[v] if score not nil
            - omit newMapIsNotNew if score is same as before
    - at this point, after these validations, newMapIsNew will contain scores to insert, and newMapIsNotNew will contain scores to update.
    - generate jwt key with ttl 5m, and send POST to `localhost:${BACKEND_HTTP_PORT ?? 8080}/api/maple/characters/culvert`
        - separate post request to isNew: true for new inserts, versus isNew: false for update statements
        - only POST request if len(map) > 0
        - 2 api calls total, minimum 1 or 0
    - finally reply to command interaction success or error
