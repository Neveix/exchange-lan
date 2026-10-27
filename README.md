# Exchange in Lan

Program that helps you to communicate between computers in LAN.

Capabilities:
- Send files
- Receive files
- Send text
- Receive text
- Send text via `stdin` mode

## Run

Clone the repository, and run:

```bash
go run .
```

## Build

```bash
./build_linux.sh
```

## Usage

```bash
# Get help
build/exchange-lan --help

# Add text to send to all clients
build/exchange-lan --text "text"

# Send text from stdin to all clients
echo "text" | build/exchange-lan --stdin

# Apply files, send them to all of clients
build/exchange-lan file1.txt file2.txt
```

