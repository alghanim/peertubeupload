# PeerTubeUpload

PeerTubeUpload is a Go application designed to upload media files to a PeerTube instance. PeerTube is a decentralized, federated video platform powered by ActivityPub and WebTorrent.

## Overview

The application loads a configuration from a `config.json` file, which includes details about the PeerTube API (URL and port), the number of threads to use for processing, and details about where to load media files from (a folder or a database).

It uses an `Authenticator` interface to handle login prerequisites and token management for the PeerTube API. Depending on the configuration, it either gathers paths to media files from a specified folder or from a database. These paths are sent to a channel for processing.

The application then loops over the channel of file paths, uploading each file to the PeerTube instance. It uses a semaphore to limit the number of concurrent uploads based on the configured number of threads. After each upload, it logs the result to either a database or a file, depending on the configuration.

## Prerequisites

Before running the application, you need to install the following packages:

```bash
sudo apt install ffmpeg
sudo apt install build-essential
```

Please note that these commands are for Debian-based systems like Ubuntu. If you're using a different Linux distribution, you'll need to use the appropriate package manager commands.

## Configuration

You need to provide a `config.json` file with the following details:

- `APIConfig`: Details about the PeerTube API, including the URL and port.
- `ProccessConfig`: The number of threads to use for processing.
- `LoadType`: Details about where to load media files from (a folder or a database).
- `FolderConfig`: If loading from a folder, the path to the folder.
- `DBConfig`: If loading from a database, the database configuration details.

## Running the Application

To run the application, use the `go run` command:

```bash
go run main.go
```

Please ensure that you have the Go programming language installed and correctly set up on your system.

## Contributing

Contributions are welcome! Please feel free to submit a pull request.