# Remote Code Execution Microservices (Go + C++)

This project provides a minimal microservice architecture for remote code execution supporting Go and C++.

## Components

- **api-gateway**: Accepts submissions, exposes status/result endpoints.
- **executor-go**: Consumes SQS messages for Go jobs, runs `go run` inside container, uploads output to S3, updates DynamoDB.
- **executor-cpp**: Consumes SQS messages for C++ jobs, compiles with `g++ -std=c++17 -O2`, runs, uploads output to S3, updates DynamoDB.
- **shared**: Common models, config, store (DynamoDB helpers), and queue helpers.

## Data Flow

1. Client POSTs `/submit` with `{ userId, language: go|cpp, code, input }`.
2. `api-gateway` stores job (status=queued) in DynamoDB and pushes minimal message `{executionId, language}` to SQS.
3. Appropriate executor polls SQS, fetches job from DynamoDB, marks `running`.
4. Code executed with timeout (`EXEC_TIMEOUT_SEC`, default 10s). Output + stderr combined -> S3 object `outputs/<executionId>.txt`.
5. DynamoDB updated with `status` (completed|failed|timeout), `outputPath`, `stdoutPreview`, timestamps, and error if any.
6. Client polls `/status?id=<executionId>` or requests `/result?id=<executionId>` for a presigned URL + preview.

## DynamoDB Item Fields

| Field                                           | Description                               |
| ----------------------------------------------- | ----------------------------------------- | ------- | --------- | ------ | ------- |
| executionId (PK)                                | Unique job id                             |
| userId                                          | Arbitrary user identifier                 |
| language                                        | go or cpp                                 |
| code                                            | Raw source code                           |
| input                                           | Optional stdin content                    |
| status                                          | queued                                    | running | completed | failed | timeout |
| error                                           | Error/truncated stderr message            |
| outputPath                                      | S3 URI of combined output file            |
| stdoutPreview                                   | First ~500 chars of stdout                |
| createdAt / updatedAt / startedAt / completedAt | Timestamps (RFC3339)                      |
| execDurationMs                                  | Milliseconds runtime (excludes S3 upload) |

## Environment Variables

Set these for every service (compose passes through):

- `AWS_REGION`
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` (dummy acceptable for LocalStack)
- `DYNAMODB_TABLE`
- `CODE_EXEC_BUCKET`
- `SQS_QUEUE_URL`
- `EXEC_TIMEOUT_SEC` (executors only, optional)

## Local Development

You can use real AWS resources or LocalStack (not yet wired here—future enhancement).

### Build & Run (Docker Compose)

```bash
# Ensure env vars exported or placed in a .env file for docker-compose
docker compose build
docker compose up -d
```

### Submit a Job

```bash
curl -X POST http://localhost:8080/submit \
  -H 'Content-Type: application/json' \
  -d '{"userId":"u1","language":"go","code":"package main\nimport (\n\t\"fmt\"\n)\nfunc main(){fmt.Println(\"hi\")}","input":""}'
```

Response:

```json
{
  "executionId": "...",
  "status": "queued",
  "language": "go",
  ...
}
```

### Check Status

```bash
curl "http://localhost:8080/status?id=<executionId>"
```

Will include: status, stdoutPreview (after completion), error, outputPath.

### Get Result (Presigned URL)

```bash
curl "http://localhost:8080/result?id=<executionId>"
```

Returns JSON with `url` and `stdoutPreview`.

## C++ Example

```bash
curl -X POST http://localhost:8080/submit \
  -H 'Content-Type: application/json' \
  -d '{"userId":"u1","language":"cpp","code":"#include <bits/stdc++.h>\nusing namespace std; int main(){string s; if(!(cin>>s)) return 0; cout<<s<<\\n;}","input":"hello"}'
```

## Timeout Handling

If execution exceeds `EXEC_TIMEOUT_SEC`, status becomes `timeout` and error field contains a message.

## Security Notes (Further Hardening Needed)

- Current containers run arbitrary code with full process privileges of container.
- Recommended enhancements: cgroup CPU/mem limits, seccomp profiles, gVisor/Firecracker isolation, restrict network egress, sanitize code size.
- Consider separate per-job ephemeral containers instead of in-process `go run` / compiled binary reuse.

## Future Enhancements

- LocalStack docker-compose integration.
- Per-language SQS queues / SNS fan-out.
- Rate limiting & auth (API keys / JWT).
- Output size streaming & pagination.
- Persistent logs & metrics (CloudWatch / OpenTelemetry).
- Websocket / SSE for real-time status updates.

## Cleanup & Costs

Ensure you delete S3 objects and DynamoDB items for test jobs to control costs in real AWS.

## Architecture Diagram (Text)

```
Client -> API Gateway -> DynamoDB (create item)
                    |-> SQS (enqueue {id,lang})
SQS -> executor-go (filter lang=go) -> fetch item -> run -> S3 upload -> DynamoDB update
SQS -> executor-cpp (filter lang=cpp) -> fetch item -> run -> S3 upload -> DynamoDB update
Client -> /status -> DynamoDB
Client -> /result -> DynamoDB -> (presign) S3
```

## Troubleshooting

- Stuck in queued: check executors logs & SQS queue size.
- Missing outputPath: verify S3 bucket exists & permissions.
- Timeout quickly: adjust `EXEC_TIMEOUT_SEC`.
- Docker build issues: ensure relative `shared` module path matches build context.

---

This README covers how to run, extend, and harden the system.
