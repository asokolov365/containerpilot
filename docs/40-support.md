# Support

## Where to file issues

Please report any issues you encounter with ContainerPilot or its documentation by [opening a Github issue](https://github.com/asokolov365/containerpilot/issues). When creating a bug report, please include as many details as possible, including the output of `containerpilot -version` and any steps needed to reproduce the issue if possible. If you can reproduce the issue with debug logging on, please include any logs you can provide. If you're reporting a crash, include the stack trace. Check the open issues to see if anyone else is reporting a similar problem.

If you are a Joyent support customer, we encourage you to report ContainerPilot issues on [Joyent GitHub](https://github.com/joyent/containerpilot/issues) so their resolution can be shared with the community. But the Support team will be happy give you direct support via Zendesk as well.

## Contributing

ContainerPilot is open source under the [Mozilla Public License 2.0](https://github.com/asokolov365/containerpilot/blob/master/LICENSE).

Pull requests on GitHub are welcome on any issue. If you'd like to propose a new feature, it's often a good idea to discuss the design by opening an issue first. We'll mark these as [`proposals`](https://github.com/asokolov365/containerpilot/issues?q=is%3Aopen+is%3Aissue+label%3Aproposal), and roadmap items will be maintained as [`enhancements`](https://github.com/asokolov365/containerpilot/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement).

Many of our contributors have never contributed to an open source golang project before. If you are looking for a good first contribution, check out the [`help wanted` label](https://github.com/asokolov365/containerpilot/issues?q=is%3Aopen+is%3Aissue+label%3A"help+wanted"); not that we don't want help anywhere else of course! But these are low-hanging fruit to get started.

Please make sure you've added tests for any new feature or tests that prove a bug has been fixed. Run `make lint` before submitting your PR.


## Backwards compatibility

While it's easy to say "just use semver", in practice there are several interfaces to consider for what it means to be backwards compatible.

##### Interface with Consul

ContainerPilot includes bindings to the Consul API via the [Hashicorp library](https://github.com/hashicorp/consul/tree/master/api). The Consul API has to date been backwards compatible but not necessarily forwards compatible. Bumping the required version of the Consul API will be considered a minor version bump to ContainerPilot unless upstream breaks backwards compatibility.

##### Interface with Vault

ContainerPilot includes bindings to the Vault API via the [Hashicorp library](https://github.com/hashicorp/vault/tree/master/api). The Vault API has to date been backwards compatible but not necessarily forwards compatible. Bumping the required version of the Vault API will be considered a minor version bump to ContainerPilot unless upstream breaks backwards compatibility.

##### Interface with Prometheus-compatible servers

ContainerPilot acts as a Prometheus client via the [Prometheus client library](https://github.com/prometheus/client_golang). Bumping the required version of the Prometheus API will be considered a minor version bump to ContainerPilot unless upstream breaks backwards compatibility.

##### Interface with other ContainerPilot instances

ContainerPilot does not coordinate between instances of ContainerPilot except through Consul service discovery. It is safe to mix versions of ContainerPilot in different containers.

##### ContainerPilot configuration and behavior

The ContainerPilot configuration syntax will follow a semver approach. Fixing a bug in an existing feature will be a patch version bump. Adding a new configuration feature will result in a minor version bump. Existing features will not have their behavior or configuration syntax changed.

The core behavior of behavior hooks fired by job `exec` or health check `exec` is to fork/exec the hook, forward stdout/stderr to ContainerPilot's own stdout/stderr, and interpret a non-zero exit code as an error which fires `ExitFailed` event, and a zero exit code as a success which fires an `ExitSuccess` event. This behavior is guaranteed to be stable for >= 3.x major version.

##### The internal ContainerPilot APIs with its golang packages

Although the ContainerPilot code base is broken into multiple packages, the interfaces have not been designed for independent consumption. The stability of these APIs is not guaranteed.

##### The interface with the ContainerPilot development community

The master branch on GitHub for this fork of ContainerPilot will be for 4.x development. See the contribution guidelines for more information.
