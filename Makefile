.PHONY: up down reset test grade format format-check lint

up:
	docker compose up --build

down:
	docker compose down

reset:
	docker compose down -v
	docker compose up -d db

test:
	docker compose run --rm api pytest -q

# Internal acceptance suite. Excluded from submission archives.
grade:
	docker compose run --rm api pytest -q grading

format:
	docker compose run --rm --no-deps api ruff format .
	docker compose run --rm --no-deps api ruff check --fix .

format-check:
	docker compose run --rm --no-deps api ruff format --check .
	docker compose run --rm --no-deps api ruff check .
