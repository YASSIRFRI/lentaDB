

# <div align="center" class="nerd-font">Lenta DB</div>
<div align="center">
  <img src="logo.png" alt="Lenta DB Logo" width="200"/>
</div>

Lenta DB is a persistent key-value store implemented in the Go programming language. It is designed to provide efficient and reliable storage for key-value pairs, focusing on performance and simplicity.
## Table of Contents
## Table of Contents
1. <span style="color:black; text-decoration:none;">[**Description**](#description)</span>
2. <span style="color:black; text-decoration:none;">[**Deployment**](#deployment)</span>
    1. <span style="color:black; text-decoration:none;">[**Environment Variables**](#environment-variables)</span>
    2. <span style="color:black; text-decoration:none;">[**Cache Impact on Integrity**](#cache-impact-on-integrity)</span>
        1. <span style="color:black; text-decoration:none;">[**Asynchrony of API Requests**](#asynchrony-of-api-requests)</span>
        2. <span style="color:black; text-decoration:none;">[**Amortized Flush Price**](#amortized-flush-price)</span>
        3. <span style="color:black; text-decoration:none;">[**Read-Heavy Usage**](#read-heavy-usage)</span>
        4. <span style="color:black; text-decoration:none;">[**Crash Recovery**](#crash-recovery)</span>
3. <span style="color:black; text-decoration:none;">[**Architecture**](#architecture)</span>
    1. <span style="color:black; text-decoration:none;">[**Memtable**](#memtable)</span>
    2. <span style="color:black; text-decoration:none;">[**SST Files Structure**](#sst-files-structure)</span>
        1. <span style="color:black; text-decoration:none;">[**Header**](#header)</span>
        2. <span style="color:black; text-decoration:none;">[**Encoding**](#encoding)</span>
    3. <span style="color:black; text-decoration:none;">[**Write-Ahead Log (WAL)**](#write-ahead-log-wal)</span>
4. <span style="color:black; text-decoration:none;">[**Usage**](#usage)</span>
5. <span style="color:black; text-decoration:none;">[**License**](#license)</span>


## Description
This repository contains the source code and technical documentation for the Lenta DB key-value store. The system is built to optimize read and write operations, providing a high-performance solution for persistent storage.

## Deployment
### Environment Variables
To deploy Lenta DB, set up the necessary environment variables in a configuration file (e.g., .env). Specify important parameters such as cache size, max file size, and entry length.

### Cache Impact on Integrity
Carefully configure the cache size to balance memory usage and system performance. A very low cache size may lead to frequent cache evictions, impacting both read and write performance.

#### Asynchrony of API Requests
Consider the asynchrony of API requests processing when determining the optimal cache size. In scenarios with asynchronous API requests, a larger cache size may enhance overall system responsiveness.

#### Amortized Flush Price
Setting a large key-value store with a correspondingly large cache size may result in an expensive amortized flush price. Evaluate the trade-off between cache size and data retrieval latency based on your use case.

#### Read-Heavy Usage
For read-heavy workloads, a lower cache size (not less than 100) may be acceptable, focusing on minimizing memory usage.

#### Crash Recovery
In case of a system crash or unexpected shutdown, the key-value store implements a crash recovery mechanism. The application checks for the presence of a Write-Ahead Log (WAL) file on startup, ensuring data consistency and integrity are maintained.

**Note:** Crash recovery assumes the log file is never corrupted or impacted. Regular monitoring and integrity checks of the log file are advisable.

## Architecture
The architecture of Lenta DB is designed to optimize read and write operations through key components, including the Memtable, SST files, and Write-Ahead Log (WAL).

### Memtable
The Memtable resides in memory and facilitates rapid read and write operations. It acts as an in-memory cache for frequently accessed key-value pairs, providing low-latency writes for write-intensive workloads.

### SST Files Structure
SST files are used for persistent storage, structured for efficient retrieval and storage. The structure includes a header section, encoded key-value pairs, and a final checksum.

#### Header
The header of an SST file contains metadata information crucial for proper file handling and retrieval during read operations.

#### Encoding
Key-value pairs within the SST file are encoded to optimize storage space and facilitate quick decoding during retrieval. Common encoding techniques include variable-length encoding and compression.

### Write-Ahead Log (WAL)
To ensure data durability and recovery in the event of system failures, Lenta DB employs a Write-Ahead Log (WAL). Write operations are first recorded in the WAL before being applied to the Memtable. This sequential log allows for the replaying of operations in case of a crash or unexpected shutdown, ensuring database integrity.

## Usage
Provide instructions on how to use and integrate Lenta DB into different projects.


## License
This project is licensed under the [MIT License](LICENSE).

