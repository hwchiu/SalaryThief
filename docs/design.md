# SalaryThief 設計文件

這份文件寫的是「為什麼這樣收、每個 collector 做什麼、資料落到哪」。
實作對應 `internal/collect`、`internal/redfish`、`internal/metrics`、`internal/opensearch`。

相關短文：

- 管線與 Kind lab：[`architecture.md`](architecture.md)
- 維運問題對照表：[`ops-model.md`](ops-model.md)

---

## 1. 目標與非目標

### 目標

讓維運團隊從 BMC / Redfish 掌握實體機器，而不必登入每台 iDRAC / iLO / XCC。

要能回答的問題：

- 這台 BMC 現在通不通、主機開不開機
- 這台是哪一款、序號 / SKU / 資產編號是什麼
- CPU、DIMM、碟、NIC 有哪些，健不健康
- 機箱熱不熱、風扇轉不轉、功耗與 PSU 狀態
- BIOS / BMC / 裝置韌體是哪個版本
- 目前 boot override 指到哪
- SEL / lifecycle 最近有沒有 Warning / Critical
- SSD 剩餘壽命、BMC 有沒有宣告預測故障

### 非目標

BMC 不是主機上的 node exporter，也不是儲存陣列的效能探針。

明確不做：

- 碟 IOPS、throughput、latency、queue depth
- SMART 原始屬性（重配置扇區、CRC、Raw Read Error）
- 作業系統檔案系統、mount、inode
- 應用層 SLI / 請求延遲
- 從 collector 對機器下電、改 boot、掛 virtual media（本期只讀）

效能與 SMART 細節若需要，另外用 OS agent 或廠牌 telemetry，不要塞進 Redfish scrape。

---

## 2. 資料怎麼分

| 問題型態 | 後端 | 理由 |
|---|---|---|
| 現在健不健康、過去 6 小時熱不熱 | Prometheus | 固定 cardinality 的 gauge，適合告警與圖 |
| 這顆碟序號、MAC、完整 snapshot | OpenSearch `bmc-inventory` | 文件型、欄位多、變動慢 |
| 什麼時候變成 Critical、SEL 原文 | OpenSearch `bmc-events` | 事件流，不是時間序列 |

一輪 scrape 產出一份 `collect.Snapshot`：

1. `metrics.Registry.Observe` 更新 Prometheus gauge（先依 `target` 清舊 series，再寫新值）
2. 若 OpenSearch 開啟，整份 snapshot 進 `bmc-inventory`，`EventsFrom(snapshot)` 進 `bmc-events`

Prometheus **不**直接 scrape 各 BMC。Collector 自己擁有 target 清單並對 BMC 做 Redfish GET，再對外暴露 `:9100/metrics`。這樣 BMC 連線數可控，也符合「一份機器清單、很多台機器」的維運習慣。

---

## 3. 抓取模型

從 Service Root 出發，只 follow `@odata.id`。`vendor:` 先當 label。子資源失敗不打掉整輪。本期 Basic auth。

## 4. Health 編碼

Unknown=0 OK=1 Warning=2 Critical=3。預測故障即使 health 仍是 OK 也要出 event。

## 5–6. Collectors

`systems` `chassis` `managers` `thermal` `power` `inventory` `storage` `network` `firmware` `logs` `boot`

Storage 只做維運盤點與預警，不收 IOPS / SMART / 碟溫 / RAID rebuild。沒有 life-left 就不寫 0。

Logs 每個 service 最多 40 筆，只進 OpenSearch。Firmware 用 info metric。Boot 只觀察不下發。

完整節次見 repo 內同檔本檔第 1–10 節。
