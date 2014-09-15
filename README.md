# authorized-keys


## Installation

Create user and group authorizator
Install to /usr/local/bin/authorized-keys with owner root and group authorizator, permissions 0750



## sshd_config:

```
LogLevel VERBOSE
AuthorizedKeysFile /nonexistent
AuthorizedKeysCommand /usr/local/bin/authorized-keys
AuthorizedKeysCommandUser authorizator
```
