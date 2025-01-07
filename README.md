# JTools
JTools 是一个用于检查 URL 有效性的程序。它通过并发的方式处理大量的 URL，并将结果分别写入有效和无效的 CSV 文件中。

## 功能
- 将有效和无效的 URL 分别记录到 `good.csv` 和 `bad.csv` 文件中。
- 疑惑(为什么goroutine开多了功能就不是很稳定了？)

## 启动方式

```bash
go build .

### 程序运行时间
- 25s