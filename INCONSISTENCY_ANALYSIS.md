# MongoBleed: Python vs Go Implementation Analysis

## Executive Summary

This document provides a detailed analysis of inconsistencies between the Python and Go implementations of MongoBleed (CVE-2025-14847) from the FraktalDeFiDAO repository.

## 🔍 Methodology

**Python Version**: `mongobleed.py` (129 lines, 4.26KB)  
**Go Version**: `mongobleed.go` (200+ lines, ~7.2KB)  
**Analysis Date**: December 29, 2025  
**Scope**: Functional equivalence, protocol implementation, error handling, and output consistency

---

## ✅ Functional Equivalence

### Core Exploit Technique ✓
Both implementations correctly implement the same core vulnerability:
- **Inflated Document Length**: Send BSON with inflated `doc_len`
- **Zlib Compression**: Compress the malicious BSON
- **OP_COMPRESSED Protocol**: Wrap in MongoDB wire protocol
- **Memory Leak Extraction**: Parse error responses for leaked data

### MongoDB Wire Protocol ✓
- **OP_COMPRESSED (2012)**: Both use correct opcode
- **OP_MSG (2013)**: Both use correct original opcode
- **Zlib Compressor ID**: Both use `0x02`
- **Header Format**: Both implement 16-byte header correctly

---

## ⚠️ Identified Inconsistencies

### 1. **Buffer Size Calculation** 🔴

**Python Version**:
```python
response = send_probe(args.host, args.port, doc_len, doc_len + 500)
```

**Go Version**:
```go
response, err := sendProbe(*host, *port, docLen, docLen+500)
```

**Inconsistency**: Both use `doc_len + 500` for buffer size, but this is actually **correct** and consistent.

**Status**: ✅ **CONSISTENT**

---

### 2. **Network Timeout Handling** 🟡

**Python Version**:
```python
sock = socket.socket()
sock.settimeout(2)
sock.connect((host, port))
```

**Go Version**:
```go
conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
// Plus additional read deadline in receive loop
conn.SetReadDeadline(time.Now().Add(2 * time.Second))
```

**Inconsistency**: 
- Python: Single 2-second timeout for entire operation
- Go: Connection timeout + per-read deadline (more robust)

**Impact**: Go version has better timeout handling and won't hang on partial responses.

**Status**: ⚠️ **DIFFERENT (Go is more robust)**

---

### 3. **Response Reading Logic** 🔴

**Python Version**:
```python
response = b''
while len(response) < 4 or len(response) < struct.unpack('<I', response[:4])[0]:
    chunk = sock.recv(4096)
    if not chunk:
        break
    response += chunk
```

**Go Version**:
```go
response := make([]byte, 0)
temp := make([]byte, 4096)
for {
    conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    n, err := conn.Read(temp)
    if err != nil {
        break
    }
    response = append(response, temp[:n]...)
    if len(response) >= 4 {
        msgLen := binary.LittleEndian.Uint32(response[:4])
        if len(response) >= int(msgLen) {
            break
        }
    }
}
```

**Inconsistency**:
- Python: No timeout on recv() calls (can hang indefinitely)
- Go: Per-read deadline (2 seconds per read)
- Python: Breaks on empty chunk, Go: Breaks on error

**Impact**: Python version may hang on certain network conditions; Go version is more reliable.

**Status**: ⚠️ **DIFFERENT (Go is more robust)**

---

### 4. **Error Handling Philosophy** 🔴

**Python Version**:
```python
try:
    # risky operation
except:
    return b''  # Silent failure
```

**Go Version**:
```go
result, err := riskyOperation()
if err != nil {
    return nil, err  // Explicit error propagation
}
```

**Inconsistent Error Handling**:

| Component | Python | Go | Impact |
|-----------|--------|----|--------|
| Network Connection | Silent failure | Explicit error | Go provides better debugging |
| Zlib Decompression | Silent failure | Explicit error | Python may miss valid leaks |
| Response Parsing | Silent failure | Explicit error | Python suppresses parsing errors |
| File Writing | Exception raised | Explicit error check | Python crashes on file errors |

**Status**: 🔴 **SIGNIFICANTLY DIFFERENT**

---

### 5. **Memory Management** 🟡

**Python Version**:
```python
all_leaked = bytearray()
unique_leaks = set()
```

**Go Version**:
```go
allLeaked := make([]byte, 0)
uniqueLeaks := make(map[string]bool)
```

**Inconsistency**:
- Python: Uses `bytearray` + `set` for deduplication
- Go: Uses `[]byte` + `map[string]bool` for deduplication

**Impact**: 
- Python: More memory efficient for large data
- Go: Faster lookups but more memory overhead

**Status**: ⚠️ **DIFFERENT (Trade-offs)**

---

### 6. **Data Deduplication Logic** 🔴

**Python Version**:
```python
for data in leaks:
    if data not in unique_leaks:
        unique_leaks.add(data)
        all_leaked.extend(data)
```

**Go Version**:
```go
for _, data := range leaks {
    dataStr := string(data)
    if !uniqueLeaks[dataStr] {
        uniqueLeaks[dataStr] = true
        allLeaked = append(allLeaked, data...)
    }
}
```

**Critical Inconsistency**:
- Python: Compares `bytes` objects directly (binary comparison)
- Go: Converts to `string` then compares (UTF-8 conversion)

**Impact**: 
- Python preserves exact binary data
- Go may corrupt non-UTF-8 binary data during string conversion
- **This is a data integrity issue**

**Status**: 🔴 **CRITICAL INCONSISTENCY**

---

### 7. **Secret Pattern Detection** 🟡

**Python Version**:
```python
secrets = [b'password', b'secret', b'key', b'token', b'admin', b'AKIA']
for s in secrets:
    if s.lower() in all_leaked.lower():
        print(f"[!] Found pattern: {s.decode()}")
```

**Go Version**:
```go
secrets := []string{"password", "secret", "key", "token", "admin", "AKIA"}
found := make([]string, 0)
lowerData := strings.ToLower(string(data))
for _, secret := range secrets {
    if strings.Contains(lowerData, strings.ToLower(secret)) {
        found = append(found, secret)
    }
}
```

**Inconsistency**:
- Python: Searches in aggregated `all_leaked` (after scan completes)
- Go: Searches per individual leak (during scan)
- Python: Uses byte string patterns
- Go: Uses string patterns

**Impact**: 
- Python: May find patterns across leak boundaries
- Go: Only finds patterns within individual leaks
- Go: Same UTF-8 conversion issue as #6

**Status**: ⚠️ **DIFFERENT (Behavioral)**

---

### 8. **Output Display Format** ✅

**Python Version**:
```python
preview = data[:80].decode('utf-8', errors='replace')
print(f"[+] offset={doc_len:4d} len={len(data):4d}: {preview}")
```

**Go Version**:
```go
preview := string(data)
if len(preview) > 80 {
    preview = preview[:80]
}
fmt.Printf("[+] offset=%4d len=%4d: %s\n", docLen, len(data), preview)
```

**Inconsistency**:
- Python: Explicit UTF-8 decoding with error replacement
- Go: Direct string conversion (may fail on invalid UTF-8)

**Impact**: Go version may crash on non-UTF-8 binary data.

**Status**: ⚠️ **DIFFERENT (Error Handling)**

---

### 9. **Binary Protocol Details** ✅

**Both implementations correctly implement**:
- MongoDB wire protocol header (16 bytes)
- OP_COMPRESSED structure
- Zlib compression
- BSON document format
- Little-endian byte order

**Status**: ✅ **CONSISTENT**

---

### 10. **Command Line Interface** ✅

**Both have identical CLI options**:
- `--host` (default: localhost)
- `--port` (default: 27017)
- `--min-offset` (default: 20)
- `--max-offset` (default: 8192)
- `--output` (default: leaked.bin)

**Status**: ✅ **CONSISTENT**

---

## 🎯 Critical Findings

### 🔴 Critical Issues (Data Integrity)

1. **UTF-8 Conversion Issue** (Finding #6)
   - **Impact**: Binary data corruption
   - **Root Cause**: Converting `[]byte` to `string` for map keys
   - **Fix**: Use `[]byte` as map key or byte comparison

### ⚠️ Major Differences (Robustness)

2. **Error Handling** (Finding #4)
   - **Impact**: Python version less reliable
   - **Root Cause**: Silent exception swallowing
   - **Recommendation**: Python should log errors

3. **Network Timeouts** (Findings #2, #3)
   - **Impact**: Python may hang on certain network conditions
   - **Root Cause**: No read timeouts in Python
   - **Recommendation**: Add socket timeouts to Python

### 🟡 Minor Differences (Behavior)

4. **Secret Detection Timing** (Finding #7)
   - **Impact**: Different detection capabilities
   - **Note**: Both approaches have valid use cases

5. **Output Display** (Finding #8)
   - **Impact**: Go may crash on binary data
   - **Fix**: Add UTF-8 validation to Go

---

## 📊 Summary Table

| Component | Consistency | Impact | Priority |
|-----------|-------------|---------|----------|
| Core Exploit | ✅ Consistent | None | Low |
| Wire Protocol | ✅ Consistent | None | Low |
| CLI Interface | ✅ Consistent | None | Low |
| Buffer Size | ✅ Consistent | None | Low |
| **Data Deduplication** | 🔴 **Critical** | **Data Corruption** | **High** |
| **Error Handling** | 🔴 **Different** | **Reliability** | **High** |
| **Network Timeouts** | 🔴 **Different** | **Robustness** | **Medium** |
| Secret Detection | ⚠️ Different | Behavior | Low |
| Output Display | ⚠️ Different | Stability | Low |

---

## 🔧 Recommendations

### For Go Implementation

1. **Fix Data Corruption** (Critical)
```go
// Instead of:
dataStr := string(data)
if !uniqueLeaks[dataStr] {
    uniqueLeaks[dataStr] = true
    allLeaked = append(allLeaked, data...)
}

// Use:
dataKey := fmt.Sprintf("%x", data) // or base64 encoding
if !uniqueLeaks[dataKey] {
    uniqueLeaks[dataKey] = true
    allLeaked = append(allLeaked, data...)
}
```

2. **Add UTF-8 Validation** (Medium)
```go
preview := string(data)
if !utf8.ValidString(preview) {
    preview = fmt.Sprintf("[binary: %d bytes]", len(data))
}
```

### For Python Implementation

1. **Add Error Logging** (High)
```python
import logging
logging.basicConfig(level=logging.WARNING)

try:
    # risky operation
except Exception as e:
    logging.warning(f"Operation failed: {e}")
    return b''
```

2. **Add Socket Timeouts** (Medium)
```python
sock = socket.socket()
sock.settimeout(2)
sock.setblocking(0)  # Non-blocking for better control
```

---

## 🎯 Conclusion

The Go and Python implementations are **functionally equivalent** at the core exploit level but have **significant differences** in robustness and data handling:

- **✅ Strengths**: Both correctly implement the MongoBleed exploit
- **🔴 Critical Issue**: Go version has data corruption bug
- **⚠️ Robustness**: Go version has better error handling and timeouts
- **🟡 Minor Issues**: Different approaches to secret detection and output

**Recommendation**: Fix the UTF-8 conversion issue in the Go implementation to ensure data integrity, while adopting Go's superior error handling and timeout mechanisms in the Python version.

---

## 📋 Verification Commands

To verify consistency between implementations:

```bash
# Test both versions against same target
python3 mongobleed.py --host localhost --max-offset 1000 --output python_leak.bin
./mongobleed --host localhost --max-offset 1000 --output go_leak.bin

# Compare outputs
diff python_leak.bin go_leak.bin
strings python_leak.bin | sort | uniq > python_strings.txt
strings go_leak.bin | sort | uniq > go_strings.txt
diff python_strings.txt go_strings.txt
```
