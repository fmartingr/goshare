GoShare
=======

Easy share files on your own terms.

> :warning: **Release advice**: There's no *release* as of yet. The config file may change in the process until the final steps of the [proof of concept](https://github.com/fmartingr/goshare/projects/1) are finished.

## Why?

I started to move away from private cloud services and one feature I missed was the ability to easily share screenshots when I took them.

This made me think, what if I could share easily any kind of file...?

## Installation

TODO

## Usage

TODO (depends on [Installation](#installation))

## How it works

The process is simple and as follow:

1. The script receives a path to a file.
2. A random filename is generated (based on UUID4 spec).
3. An SSH connection is made and the file is uploaded.
4. You will see in the output the URL to your uploaded file.

## Configuration

If you don't have a configuration file the first run will create one for you and then you only need to fill in he correct details.

``` json
{
  "SSH": {
    "User": "ssh username",
    "Host": "ssh host/ip",
    "Key": "~/.ssh/id_rsa",
    "Port": 22
  },
  "RemotePath": "path in the remote server to upload the file",
  "ShareUrl": "http://share.example.com/%s"
}
```

## Authors

- Felipe Martin (fmartingr)
