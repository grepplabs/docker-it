# docker-it
***Work in progress***



## Compile - https://github.com/moby/moby/issues/28269
```
$ glide create                            # Start a new workspace
$ glide install -v                        # Install packages and dependencies + strip-vendor
$ go build                                # Go tools work normally
$ glide up                                # Update to newest versions of the package
```