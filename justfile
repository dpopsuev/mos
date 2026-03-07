default: ci

ci:
    ./bin/mgate ci

ci-fast:
    ./bin/mgate ci --fast

ci-fix:
    ./bin/mgate ci --fix

build:
    go build -o ./bin/mos ./cmd/mos
    go build -o ./bin/mgov ./cmd/mgov
    go build -o ./bin/mvcs ./cmd/mvcs
    go build -o ./bin/mgate ./cmd/mgate
    go build -o ./bin/mtrace ./cmd/mtrace
    go build -o ./bin/mstore ./cmd/mstore

hook:
    ./bin/mgate hook install

# --- convenience aliases ---

test:
    go test -short ./...

test-all:
    go test ./...

bench:
    go test -bench=. -run=^$ ./testkit/stressgen/

lint:
    ./bin/mgate lint

audit:
    ./bin/mgate audit

harness:
    ./bin/mgate harness run

fmt:
    ./bin/mgov fmt .

status:
    ./bin/mgov status
