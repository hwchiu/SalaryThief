# 維運資料模型

收集目標不是「多抓幾個 sensor」，而是讓團隊能回答實體機器日常維運問題。

實作細節（路徑、指標、事件、硬碟非效能原則）見 [`design.md`](design.md)。

| 面向 | 問題 | Redfish 來源 | 去哪裡看 |
|---|---|---|---|
| 存活 | BMC 通不通、這台開不開機 | ServiceRoot, System.PowerState | Prometheus `bmc_up`, `bmc_system_power_on` |
| 身分 | 這台是哪一款、序號、資產編號 | System / Chassis identity | `bmc_system_info` + OpenSearch inventory |
| 組成 | CPU / DIMM / 碟 / NIC 有哪些、健不健康 | Processors, Memory, Storage/Drives, EthernetInterfaces | health gauges + inventory JSON |
| 環境 | 熱不熱、風扇、功耗、PSU | Thermal, Power | temp / fan / watt metrics |
| 韌體 | BIOS、BMC、裝置版號、能不能更新 | UpdateService FirmwareInventory | `bmc_firmware_info` |
| 開機 | 現在 override 去 PXE 還是 Disk | System.Boot | `bmc_boot_info` |
| 事件 | SEL / lifecycle 警告 | LogServices Entries | OpenSearch `bmc-events` |
| 預測故障 | SSD 壽命、FailurePredicted | Drive | `bmc_drive_life_left_percent`, `bmc_drive_failure_predicted` |

Prometheus 只放能畫圖、能告警的數字。  
序號、MAC、完整 snapshot、SEL 原文放 OpenSearch。
