# docker-it
[![Build Status](https://travis-ci.org/cloud-42/docker-it.svg?branch=master)](https://travis-ci.org/cloud-42/docker-it)
[![Coverage Status](https://coveralls.io/repos/github/cloud-42/docker-it/badge.svg?branch=master)](https://coveralls.io/github/cloud-42/docker-it?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloud-42/docker-it)](https://goreportcard.com/report/github.com/cloud-42/docker-it)

***Work in progress***


## Compile - https://github.com/moby/moby/issues/28269
```
$ glide create                            # Start a new workspace
$ glide install -v                        # Install packages and dependencies + strip-vendor
$ go build                                # Go tools work normally
$ glide up                                # Update to newest versions of the package
```


### Tasks 

* [ ] add unit tests
* [ ] add wait tests
* [ ] document usage
* [ ] evaluate sprig as template functions for Go templates

