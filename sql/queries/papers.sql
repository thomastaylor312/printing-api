-- name: GetPapers :many
SELECT * FROM papers;

-- name: AddPaper :one
INSERT INTO papers (name, cost_per_square_inch, finish) VALUES ($name, $cost_per_square_inch, $finish) RETURNING *;

-- name: UpdatePaper :exec
UPDATE papers SET name = $name, cost_per_square_inch = $cost_per_square_inch, finish = $finish WHERE id = $id;

-- name: DeletePaper :exec
DELETE FROM papers WHERE id = $id;
