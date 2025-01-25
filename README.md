# go-w

A Go implementation of the Linux `w` utility, displaying information about logged-in users and system load averages.

## Features

- Displays current time, system uptime, and load averages.
- Lists logged-in users, their TTYs, and session details.
- Colorful output for better readability.
- Lightweight and fast.

## Installation

### Prerequisites

- Go 1.20 or higher.
- Docker (optional, for containerization).

### Using Go

1. Clone the repository:
   ```
   git clone https://github.com/ericmanlol/go-w.git
   cd go-w
   ```

2. Build the project:
   ```
   make build
   ```

3. Install the binary:
   ```
   make install
   ```

4. Run the program:
   ```
   go-w
   ```

### Using Docker

1. Build the Docker image:
   ```
   make docker-build
   ```

2. Run the Docker container:
   ```
   make docker-run
   ```

## Usage

Run the program directly:
```
go-w
```

Example output:
```
 14:30:45 up 1:23,  load average: 0.15, 0.10, 0.05 (using /var/run/utmp)
USER     TTY      FROM             LOGIN@   IDLE   JCPU   PCPU WHAT
john     tty1     :0               14:00    .      0.00s  0.00s -
jane     pts/0    192.168.1.100    14:15    5m     0.00s  0.00s -
```

## Testing

To run the tests:
```
make test
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Commit your changes.
4. Submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Enjoy using `go-w`! 🚀