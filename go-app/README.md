# MongoBleed - Go Implementation

A Go implementation of the MongoBleed exploit for CVE-2025-14847 - MongoDB Unauthenticated Memory Leak.

Original Python version by Joe Desimone (@dez_)

## Description

This tool exploits a vulnerability in MongoDB's zlib message decompression that allows unauthenticated attackers to leak sensitive server memory. The bug causes MongoDB to return uninitialized memory contents when processing specially crafted compressed messages.

## Vulnerability Details

- **CVE**: CVE-2025-14847
- **Affected Versions**:
  - 8.2.x: 8.2.0 - 8.2.2 (Fixed in 8.2.3)
  - 8.0.x: 8.0.0 - 8.0.16 (Fixed in 8.0.17)
  - 7.0.x: 7.0.0 - 7.0.27 (Fixed in 7.0.28)
  - 6.0.x: 6.0.0 - 6.0.26 (Fixed in 6.0.27)
  - 5.0.x: 5.0.0 - 5.0.31 (Fixed in 5.0.32)

## How It Works

1. Sends a compressed message with an inflated `uncompressedSize` claim
2. MongoDB allocates a large buffer based on the attacker's claim
3. zlib decompresses actual data into the start of the buffer
4. The bug causes MongoDB to treat the entire buffer as valid data
5. BSON parsing reads "field names" from uninitialized memory until null bytes

Leaked data may include:
- MongoDB internal logs and state
- WiredTiger storage engine configuration
- System `/proc` data (meminfo, network stats)
- Docker container paths
- Connection UUIDs and client IPs

## Installation

### Prerequisites

- Go 1.18 or higher

### Build

```bash
# Clone or download the files
go build -o mongobleed mongobleed.go

# Or using make
make
```

## Usage

### Basic Scan

```bash
# Basic scan (offsets 20-8192)
./mongobleed --host <target>

# Default: localhost:27017
./mongobleed
```

### Advanced Options

```bash
# Deep scan for more data
./mongobleed --host <target> --max-offset 50000

# Custom range
./mongobleed --host <target> --min-offset 100 --max-offset 20000

# Custom port
./mongobleed --host <target> --port 27017

# Custom output file
./mongobleed --host <target> --output custom.bin
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--host` | localhost | Target MongoDB host |
| `--port` | 27017 | Target MongoDB port |
| `--min-offset` | 20 | Minimum document length to probe |
| `--max-offset` | 8192 | Maximum document length to probe |
| `--output` | leaked.bin | Output file for leaked data |

## Example Output

```
[*] mongobleed - CVE-2025-14847 MongoDB Memory Leak
[*] Author: Joe Desimone - x.com/dez_
[*] Target: localhost:27017
[*] Scanning offsets 20-8192

[+] offset=  117 len=  39: ssions^\x01�r��*YDr���
[+] offset=16582 len=1552: MemAvailable:    8554792 kB\nBuffers: ...
[+] offset=18731 len=3908: Recv SyncookiesFailed EmbryonicRsts ...

[*] Total leaked: 8748 bytes
[*] Unique fragments: 42
[*] Saved to: leaked.bin
```

## Testing with Docker

Use the original Docker Compose setup from the Python version to test against a vulnerable MongoDB instance:

```bash
# Get the docker-compose.yml from the original repository
curl -O https://raw.githubusercontent.com/joe-desimone/mongobleed/main/docker-compose.yml

# Start vulnerable MongoDB
docker-compose up -d

# Run the exploit
./mongobleed --host localhost
```

## Comparison with Python Version

| Feature | Go Version | Python Version |
|---------|------------|----------------|
| Performance | Faster (compiled) | Slower (interpreted) |
| Binary Size | Single static binary | Requires Python runtime |
| Memory Usage | Lower | Higher |
| Cross-platform | Easy cross-compilation | Python dependency |
| Dependencies | Standard library only | zlib, struct, socket, re, argparse |

## Technical Implementation

### Key Components

1. **Network Communication**: Direct TCP socket operations using Go's `net` package
2. **BSON Manipulation**: Manual construction of BSON documents with inflated lengths
3. **Zlib Compression**: Using Go's `compress/zlib` package
4. **Memory Leak Extraction**: Regex-based extraction from error responses
5. **Binary Protocol**: Full implementation of MongoDB wire protocol

### MongoDB Wire Protocol

The exploit uses MongoDB's wire protocol operations:
- **OP_COMPRESSED** (2012): Compressed message opcode
- **OP_MSG** (2013): Modern message format
- **Zlib Compression**: Compressor ID 2

## Legal Notice

**This tool is for authorized security testing only. Unauthorized access to computer systems is illegal.**

- Use only on systems you own or have explicit permission to test
- Follow responsible disclosure practices
- Comply with all applicable laws and regulations
- Educational and authorized penetration testing use only

## Credits

- **Original Python Version**: Joe Desimone (@dez_)
- **CVE Discovery**: OX Security
- **Go Implementation**: This version

## References

- [Original MongoBleed Repository](https://github.com/joe-desimone/mongobleed)
- [CVE-2025-14847](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2025-14847)
- [OX Security Advisory](https://www.ox.security/)
- [MongoDB Security](https://www.mongodb.com/security/)

## License

This tool is provided for educational and authorized security testing purposes only.
