# DNSCollector Logfile S3 Uploader

`dnscollector-logfile-s3uploader` is a command-line utility written in Go, designed to integrate with the [DNSCollector's `postrotate-command` feature](https://github.com/dmachard/DNS-collector/blob/main/docs/loggers/logger_file.md#postrotate-command). It uploads rotated log files to an S3-compatible object storage bucket.

Configuration is managed entirely through environment variables, making it highly suitable for containerized environments and automated deployments.

## Features

* Seamless integration with [DNSCollector's `postrotate-command`](https://github.com/dmachard/DNS-collector/blob/main/docs/loggers/logger_file.md#postrotate-command).
* Uploads log files to AWS S3 or other S3-compatible services (e.g., MinIO, Ceph).
* Configuration via environment variables for easy setup.
* Supports environment variable expansion in configuration values for dynamic settings.
* Support for custom S3 endpoints and path-style access for S3-compatible storage.
* Allows specifying an object key prefix for better organization within the S3 bucket.
* Optionally deletes the local log file after a successful upload.

## Prerequisites

* A running instance of DNSCollector.
* AWS credentials configured in your environment. The tool uses the default AWS SDK credential chain, which searches for credentials in:
    1.  Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.).
    2.  Shared credentials file (`~/.aws/credentials`).
    3.  IAM role for an EC2 instance or ECS task.

## Installation

You can download a pre-compiled binary from the project's GitHub Releases page.

1.  Download the latest binary for your operating system and architecture.
2.  Place the compiled `dnscollector-logfile-s3uploader` binary in a directory accessible to the DNSCollector process, such as `/usr/local/bin/`.
3.  Make the binary executable:
    ```sh
    chmod +x /usr/local/bin/dnscollector-logfile-s3uploader
    ```

## Building from Source

Alternatively, you can build the binary from the source code.

1.  Clone or download the source code into a directory.

2.  You will need Go (version 1.24 or later) to build the binary. Build the binary using the Go toolchain:

    ```sh
    go build -o dnscollector-logfile-s3uploader .
    ```

3.  Place the compiled `dnscollector-logfile-s3uploader` binary in a directory accessible to the DNSCollector process, such as `/usr/local/bin/`.

## Configuration

The tool is configured using the following environment variables.

| Environment Variable             | Description                                                                                                              | Required |
| :------------------------------- | :----------------------------------------------------------------------------------------------------------------------- | :------- |
| `DNSC_LOGFILE_S3_BUCKET`         | The name of the S3 bucket to upload log files to.                                                                        | **Yes**  |
| `DNSC_LOGFILE_S3_KEY_PREFIX`     | An optional prefix for the S3 object key. The final key will be `<prefix>/<log_filename>`.                               | No       |
| `DNSC_LOGFILE_S3_ENDPOINT_URL`   | The endpoint URL for an S3-compatible storage service. Leave unset for AWS S3.                                           | No       |
| `DNSC_LOGFILE_S3_USE_PATH_STYLE` | Set to `true` to enable path-style bucket access (e.g., `endpoint/bucket/key`). Required by some S3-compatible services. | No       |
| `DNSC_LOGFILE_DELETE_ON_SUCCESS` | Set to `true` to delete the local log file after a successful upload.                                                    | No       |

### Environment Variable Expansion

The string values for `DNSC_LOGFILE_S3_BUCKET`, `DNSC_LOGFILE_S3_KEY_PREFIX`, and `DNSC_LOGFILE_S3_ENDPOINT_URL` support environment variable expansion. The tool replaces placeholders in the format `$VAR` or `${VAR}` with the corresponding values from the environment.

This allows for creating dynamic configurations. For example, you can include the server's hostname in the S3 key prefix to organize logs by the server they originated from.

## Usage with DNSCollector

To use the uploader, specify the path to the binary in the `postrotate-command` option in your DNSCollector configuration file.

**Example `config.yml`:**

```yaml
logfile:
  file-path: /var/log/dnscollector/dns.log
  mode: dnstap
  max-size: 100 # Rotate logs at 100MB

  # [Optional] Enable compression. The postrotate command will run AFTER compression.
  compress: true

  # Path to the compiled uploader program.
  postrotate-command: "/usr/local/bin/dnscollector-logfile-s3uploader"

  # [Optional] Set to true to delete the local log file after the postrotate-command exits successfully.
  # Note: You can use this option OR the uploader's built-in deletion feature via DNSC_LOGFILE_DELETE_ON_SUCCESS.
  # Using both is possible but redundant. It's recommended to choose one method.
  # postrotate-delete-success: true
```

## Workflow

When a log rotation is triggered in DNSCollector, the following process occurs:

1.  The active log file reaches its `max-size` and is closed for rotation.
2.  If `compress` is `true`, the log file is renamed with a `tocompress-` prefix and compressed into a `.gz` file.
3.  The rotated (and possibly compressed) file is renamed with a `toprocess-` prefix.
4.  DNSCollector executes the `postrotate-command`, passing three arguments to the script:
    * Arg 1: The full path to the `toprocess-` file.
    * Arg 2: The directory containing the file.
    * Arg 3: The filename without the `toprocess-` prefix.
5.  `dnscollector-logfile-s3uploader` starts, reads its configuration from environment variables, and uploads the file specified in Arg 1 to your S3 bucket. The object key is constructed using the optional prefix and the filename from Arg 3.
6.  The uploader exits with code `0` on success or a non-zero code on failure.
7.  If the uploader was configured with `DNSC_LOGFILE_DELETE_ON_SUCCESS=true`, it will delete the local `toprocess-` file after a successful upload.
8.  If the script exits successfully and DNSCollector's `postrotate-delete-success` is `true`, DNSCollector will also attempt to delete the local `toprocess-` log file.

## License

This project is licensed under the MIT License.
