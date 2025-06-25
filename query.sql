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
  id = sqlc.narg('id')
LIMIT 1;

-- name: GetEventsBySubject :many
SELECT
  *
FROM
  events
WHERE
  `id` > sqlc.narg('id')
  AND `subject` = sqlc.narg('subject')
LIMIT
  ?;

-- name: GetEventsBySubjectPrefix :many
SELECT
  *
FROM
  events
WHERE
  `id` > sqlc.narg('id')
  AND `subject` LIKE sqlc.narg('subject')
LIMIT ?;

-- name: GetEventsBySubjectAndType :many
SELECT
  *
FROM
  events
WHERE
  `id` > sqlc.narg('id')
  AND `subject` = sqlc.narg('subject')
  AND `type` = sqlc.narg('type')
LIMIT ?;

-- name: GetEventsBySubjectPrefixAndType :many
SELECT
  *
FROM
  events
WHERE
  `id` > sqlc.narg('id')
  AND `subject` LIKE sqlc.narg('subject')
  AND `type` = sqlc.narg('type')
LIMIT ?;


-- name: GetAvailableSubjects :many
SELECT
  DISTINCT `subject`
FROM
  events;

-- name: DeleteFromSubject :exec
DELETE FROM
  events
WHERE
  `subject` = sqlc.narg('subject') AND `id` >= sqlc.narg('id');

-- name: DeleteFromSubjectRecursive :exec
DELETE FROM
  events
WHERE
  `subject` LIKE sqlc.narg('subject') AND `id` >= sqlc.narg('id');

-- name: DeleteFromSubjectWithType :exec
DELETE FROM
  events
WHERE
  `subject` = sqlc.narg('subject') AND `type` = sqlc.narg('type') AND `id` >= sqlc.narg('id');

-- name: DeleteFromSubjectRecursiveWithType :exec
DELETE FROM
  events
WHERE
  `subject` LIKE sqlc.narg('subject') AND `type` = sqlc.narg('type') AND `id` >= sqlc.narg('id');

-- name: HealthCheck :one
SELECT 1;