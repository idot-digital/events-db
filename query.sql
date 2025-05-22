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
  AND MATCH(`subject`) AGAINST (? IN BOOLEAN MODE)
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
  AND MATCH(`subject`) AGAINST ($2 IN BOOLEAN MODE)
  AND `type` = $3
LIMIT 50;
