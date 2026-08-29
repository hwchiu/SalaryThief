# Test Matrix

## Unit / component tests

| ID | Area | Test | Expected |
|---|---|---|---|
| U01 | Cache | concurrent read/write | no race/panic |
| U02 | Cache | failed refresh after successful data | old data retained, freshness increases |
| U03 | Scheduler | due targets dispatched | all due targets eventually dispatched |
| U04 | Scheduler | context cancellation | workers stop cleanly |
| U05 | Worker pool | blocked endpoints | max in-flight <= configured workers |
| U06 | Errors | request timeout | classified as timeout |
| U07 | Errors | 401/403 | classified as auth |
| U08 | Errors | 500 | classified as HTTP |
| U09 | Errors | invalid payload | classified as schema |
| U10 | Partial | storage fails, thermal succeeds | storage error only |
| U11 | Metrics | scrape handler | no Redfish HTTP request occurs |
| U12 | Freshness | last success in past | data_age grows monotonically |

## Local integration tests

| ID | Scenario | Action | Expected |
|---|---|---|---|
| I01 | Healthy | start lab | healthy target exports current metrics |
| I02 | Unreachable | stop one mock | only that target goes down/stale |
| I03 | Scrape isolation | repeatedly curl `/metrics` while mock is down | responses remain fast |
| I04 | Slow BMC | inject delay > timeout | timeout/backoff visible, no scrape coupling |
| I05 | Partial failure | storage endpoint returns error | BMC stays up, storage unknown/error |
| I06 | Recovery | restore failed mock | target returns to up/current |
| I07 | Multi-target | one broken + several healthy | healthy targets continue updating |
| I08 | Restart | restart collector | scheduler/cache initialize cleanly |

## Later-phase tests already anticipated

| ID | Phase | Scenario |
|---|---|---|
| L01 | Inventory | Bay 03 Serial AAA -> BBB => replacement |
| L02 | Events | duplicate EventLog entry => one indexed event |
| L03 | OpenSearch | storage outage does not break metrics |
| L04 | Scale | 200/1k/2k/5k logical BMCs |
