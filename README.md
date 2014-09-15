authorized-keys
===============

AuthorizedKeysCommand


Create user and group authorizator
Install to /usr/local/bin/authorized-keys with owner root and group authorizator, permissions 0750


Set in sshd_config:

LogLevel VERBOSE
AuthorizedKeysFile /nonexistent
AuthorizedKeysCommand /usr/local/bin/authorized-keys
AuthorizedKeysCommandUser authorizator

