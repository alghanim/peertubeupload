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

The application requires a `config.json` file in the root directory. This file should contain the following sections:

- `APIConfig`: Details about the PeerTube API, including the URL, port, username, password, channel ID, and various settings related to downloads, comments, privacy, and transcoding.

- `LoadType`: Specifies where to load media files from (a folder or a database), whether to convert audio to MP3, the temporary folder to use, and the log type. If specific extensions are to be loaded, they can be specified here.

- `FolderConfig`: If loading from a folder, this contains the path to the folder.

- `DBConfig`: If loading from a database, this contains the database configuration details, including the type of database, username, password, port, host, database name, table name, and column names for the title, description, and file path. It also specifies whether to update the same table and any reference columns.

- `ProccessConfig`: Specifies the number of threads to use for processing.

If the `config.json` file does not exist when you run the application, a sample `config.json` file will be created with default values. You should then modify this file with your actual configuration details before running the application again.

## Running the Application

To run the application, use the `go run` command:

```bash
go run main.go
```

Please ensure that you have the Go programming language installed and correctly set up on your system.

## Contributing

Contributions are welcome! Please feel free to submit a pull request.