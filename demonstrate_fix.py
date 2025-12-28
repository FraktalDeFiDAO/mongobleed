#!/usr/bin/env python3
"""
Demonstration script showing the MongoBleed Go deduplication fix

This script simulates the old (broken) and new (fixed) behavior
to clearly show the data integrity issue and its resolution.
"""

import binascii

def simulate_old_broken_approach(data_list):
    """
    Simulate the old broken Go approach: converting []byte to string
    """
    print("🔴 OLD (BROKEN) APPROACH:")
    print("   Converting []byte -> string for deduplication keys")
    
    unique_leaks = {}
    all_leaked = bytearray()
    
    for i, data in enumerate(data_list):
        try:
            # This is the problematic conversion
            data_str = data.decode('utf-8')  # May fail or corrupt
            
            if data_str not in unique_leaks:
                unique_leaks[data_str] = True
                all_leaked.extend(data)
                print(f"   [{i}] Added: {repr(data_str[:50])}... (len={len(data)})")
            else:
                print(f"   [{i}] Skipped: {repr(data_str[:50])}... (duplicate)")
                
        except UnicodeDecodeError as e:
            print(f"   [{i}] ❌ CORRUPTION: Cannot decode binary data: {e}")
            # Data is lost!
    
    print(f"   Total unique fragments: {len(unique_leaks)}")
    print(f"   Total leaked bytes: {len(all_leaked)}")
    return bytes(all_leaked)

def simulate_new_fixed_approach(data_list):
    """
    Simulate the new fixed Go approach: hex encoding for binary-safe keys
    """
    print("\n🟢 NEW (FIXED) APPROACH:")
    print("   Using hex encoding for binary-safe deduplication keys")
    
    unique_leaks = {}
    all_leaked = bytearray()
    
    for i, data in enumerate(data_list):
        # Binary-safe key generation
        data_key = binascii.hexlify(data).decode('ascii')
        
        if data_key not in unique_leaks:
            unique_leaks[data_key] = True
            all_leaked.extend(data)
            
            # Show preview based on data type
            if all(b < 128 for b in data):  # ASCII printable
                preview = data.decode('ascii', errors='replace')
                print(f"   [{i}] Added ASCII: {repr(preview[:50])}... (len={len(data)})")
            else:
                print(f"   [{i}] Added BINARY: {data_key[:20]}... (len={len(data)})")
        else:
            preview = data.decode('utf-8', errors='replace') if all(b < 128 for b in data) else data_key[:20]
            print(f"   [{i}] Skipped: {repr(preview[:50])}... (duplicate)")
    
    print(f"   Total unique fragments: {len(unique_leaks)}")
    print(f"   Total leaked bytes: {len(all_leaked)}")
    return bytes(all_leaked)

def demonstrate_preview_formatting():
    """
    Demonstrate the preview formatting improvements
    """
    print("\n" + "="*60)
    print("PREVIEW FORMATTING DEMONSTRATION")
    print("="*60)
    
    test_cases = [
        (b"Hello World", "ASCII text"),
        ("🚀 Unicode test: 中文 & emojis 🎉".encode('utf-8'), "UTF-8 with unicode"),
        (bytes([0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD]), "Binary data"),
        (b"password=secret123", "Contains secret"),
        (bytes([0x41] * 100), "Long ASCII text"),
        (bytes(range(256)), "Full byte range"),
    ]
    
    print("\nData Type Preview Formatting:")
    print("-" * 40)
    
    for data, description in test_cases:
        print(f"\n{description}:")
        print(f"  Raw data: {repr(data[:30])}...")
        
        # Simulate old approach (may fail)
        try:
            old_preview = data.decode('utf-8')[:80]
            print(f"  Old preview: {repr(old_preview)}")
        except UnicodeDecodeError:
            print(f"  Old preview: ❌ UNICODE DECODE ERROR")
        
        # Simulate new approach
        if all(b < 128 and b >= 32 for b in data):  # All printable ASCII
            new_preview = data.decode('ascii')[:80]
            print(f"  New preview: {repr(new_preview)}")
        else:
            if len(data) > 40:
                new_preview = f"[binary: {len(data)} bytes] {binascii.hexlify(data[:20]).decode('ascii')}..."
            else:
                new_preview = f"[binary: {len(data)} bytes] {binascii.hexlify(data).decode('ascii')}"
            print(f"  New preview: {new_preview}")

def main():
    print("MongoBleed Go Implementation - Deduplication Fix Demonstration")
    print("=" * 65)
    
    # Test data that demonstrates the problem
    print("\n" + "="*60)
    print("TEST SCENARIO: Mixed Binary and UTF-8 Memory Dumps")
    print("="*60)
    
    # Simulate memory dumps that would come from MongoDB
    test_data = [
        # Valid UTF-8 strings
        b"MemAvailable: 8554792 kB",
        b"MongoDB connection UUID: 123e4567-e89b-12d3-a456-426614174000",
        b"password=secret123",  # This should trigger secret detection
        
        # Binary data that looks like UTF-8 but isn't
        bytes([0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD]),
        bytes([0x41, 0x42, 0x43, 0x80, 0x81, 0x82]),  # Mixed ASCII/binary
        
        # Duplicates to test deduplication
        b"MemAvailable: 8554792 kB",  # Duplicate of first
        bytes([0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD]),  # Duplicate of binary
        
        # More complex cases
        b"/proc/meminfo data follows...",
        bytes([0xDE, 0xAD, 0xBE, 0xEF] * 10),  # Repeated binary pattern
        b"Docker container path: /var/lib/docker/...",
        
        # Edge cases
        b"",  # Empty
        b"\x00\x01\x02",  # NUL bytes
        "🚀 Unicode: 测试 🎉".encode('utf-8'),  # Unicode
        bytes([i for i in range(256)]),  # Full byte range
    ]
    
    print(f"\nTest data contains {len(test_data)} memory fragments:")
    for i, data in enumerate(test_data):
        if all(b < 128 and b >= 32 for b in data):
            print(f"  [{i:2}] ASCII text: {repr(data[:40])}")
        else:
            print(f"  [{i:2}] Binary data: {len(data)} bytes")
    
    # Run both approaches
    old_result = simulate_old_broken_approach(test_data)
    new_result = simulate_new_fixed_approach(test_data)
    
    # Compare results
    print("\n" + "="*60)
    print("COMPARISON SUMMARY")
    print("="*60)
    
    print(f"\n📊 Data Integrity:")
    print(f"   Old approach leaked: {len(old_result)} bytes")
    print(f"   New approach leaked: {len(new_result)} bytes")
    print(f"   Difference: {len(new_result) - len(old_result)} bytes")
    
    if old_result != new_result:
        print("\n⚠️  DATA INTEGRITY ISSUE DETECTED!")
        print("   The old approach lost or corrupted data.")
        print("   The new approach preserves all binary data correctly.")
    else:
        print("\n✅ Both approaches produced identical results (test case limitation)")
    
    # Demonstrate preview formatting
    demonstrate_preview_formatting()
    
    print("\n" + "="*60)
    print("CONCLUSION")
    print("="*60)
    print("""
The fix addresses a critical data integrity issue where:

🔴 OLD APPROACH:
   • Converted []byte -> string using UTF-8 decoding
   • Could corrupt binary memory dumps
   • Lost data on invalid UTF-8 sequences
   • Caused incorrect deduplication

🟢 NEW APPROACH:
   • Uses hex encoding for binary-safe keys
   • Preserves exact byte sequences
   • Handles any binary data correctly
   • Matches Python implementation behavior

✅ RESULT:
   The Go implementation now maintains data integrity while preserving
   Go's superior error handling, network resilience, and performance.
""")

if __name__ == "__main__":
    main()
