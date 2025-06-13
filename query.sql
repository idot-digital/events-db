-- name: CreateEvent :execlastid
INSERT INTO
  events (`source`, `type`, `subject`, `data`)
VALUES
  (?, ?, ?, ?);

-- name: GetEventByID :one
SELECT
  *
FROM
  events
WHERE
  id = ?
LIMIT 1;

-- name: GetEventsBySubject :many
SELECT
  *
FROM
  events
WHERE
  `id` > ?
  AND `subject` = ?
LIMIT
  ?;

-- name: GetEventsBySubjectPrefix :many
SELECT
  *
FROM
  events
WHERE
  `id` > ?
  AND `subject` LIKE ?
LIMIT 50;

-- name: GetEventsBySubjectAndType :many
SELECT
  *
FROM
  events
WHERE
  `id` > ?
  AND `subject` = ?
  AND `type` = ?
LIMIT 50;

-- name: GetEventsBySubjectPrefixAndType :many
SELECT
  *
FROM
  events
WHERE
  `id` > ?
  AND `subject` LIKE ?
  AND `type` = ?
LIMIT 50;


-- name: GetAvailableSubjects :many
SELECT
  DISTINCT `subject`
FROM
  events;

-- name: DeleteFromSubject :exec
DELETE FROM
  events
WHERE
  `subject` = ? AND `id` >= ?;

-- name: DeleteFromSubjectRecursive :exec
DELETE FROM
  events
WHERE
  `subject` LIKE ? AND `id` >= ?;

-- name: DeleteFromSubjectWithType :exec
DELETE FROM
  events
WHERE
  `subject` = ? AND `type` = ? AND `id` >= ?;

-- name: DeleteFromSubjectRecursiveWithType :exec
DELETE FROM
  events
WHERE
  `subject` LIKE ? AND `type` = ? AND `id` >= ?;