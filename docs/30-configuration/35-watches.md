# Watches

A `watch` is a configuration of a surveillance entity to monitor.

- The Consul watch polls the state of the service and emits events when the service becomes healthy, becomes unhealthy, or has a change in the number of instances.
- The Vault watch monitors the entire secret or a field of the secret and emits events when the secret has a change.
- The File watch monitors the state of the file and emits events when the file has been changed.

Note that a watch does not include a behavior; watches only emit the event so that jobs can consume that event.

Watch configurations include only the following fields:

```json5
watches: [
  {
    name: "backend",
    source: "consul", // default consul if not specified
    interval: 3,
    tag: "prod",    // optional
    dc: "us-east-1" // optional
  },
  {
    name: "secret/data/database",
    source: "vault",
    interval: 10,
    tag: "password"    // secret field, optional
  },
  {
    name: "/etc/ssl/cert.pem",
    source: "file",
    interval: 10
  }
]
```

The `interval` is the time (in seconds) between polling attempts to Consul. The `name` is the service or path to query, the `tag` is the optional tag (or field in case of vault query) to add to the query, and the `dc` is the optional Consul [datacenter](https://www.consul.io/docs/guides/datacenters.html) to query.

A Consul watch keeps an in-memory list of the healthy IP addresses associated with the service. The list is not persisted to disk and if ContainerPilot is restarted it will need to check back in with the canonical data store, which is Consul. If this list changes between polls, the watch emits one or two events:

- A `changed` event is emitted whenever there is a change.
- A `healthy` event is emitted whenever the watched service becomes healthy. This might mean that the state was previously unknown (as when ContainerPilot first starts up) or that it was previously unhealthy and is now healthy. This event will only be fired once for each change in status or count of instances. Subsequent polls that return the same value will not emit the event again.
- A `unhealthy` event is emitted whenever the watched service becomes unhealthy. This might mean that the service is not yet running when we first poll, or that it was previously healthy and is now unhealthy. This event will only be fired once for each change of status. Subsequent polls that return the same value will not emit the event again.

The name of the events emitted by watches are namespaced so as not to collide with internal job names. These events are prefixed by `watch`. Here is an example configuration for a job listening for a watch event:

```json5
jobs: [
  {
    name: "update-app",
    exec: "/bin/update-app.sh",
    when: {
      source: "watch.backend",
      each: "changed"
    }
  }
],
watches: [
  {
    name: "backend",
    interval: 3
  }
]
```

In this example, the watch `backend` will be checked every 3 seconds. Each time the watch emits the `changed` event, the `update-app` job will execute `/bin/update-app.sh`.

A Vault watch keeps an in-memory list of the secrets associated with the path. The list is not persisted to disk and if ContainerPilot is restarted it will need to check back in with the canonical data store, which is Vault. If this list changes between polls, the watch emits one or two events:

- A `changed` event is emitted whenever there is a change.
- A `healthy` event is emitted whenever ContainerPilot is able to read the path from Vault. This might mean that the state was previously unknown (as when ContainerPilot first starts up) or that it was previously unhealthy and is now healthy. This event will only be fired once for each change in status. Subsequent polls that return the same value will not emit the event again.
- A `unhealthy` event is emitted whenever ContainerPilot is unable to read the path from Vault. This might mean that the path does not yet exist when we first poll, or that Vault is unreachable. This event will only be fired once for each change of status. Subsequent polls that return the same value will not emit the event again

```json5
jobs: [
  {
    name: "update-app",
    exec: "/bin/update-app.sh",
    when: {
      source: "watch.secret/data/database",
      each: "changed"
    }
  }
],
watches: [
  {
    name: "secret/data/database",
    source: "vault",
    tag: "password",
    interval: 10
  }
]
```

In this example, the watch `secret/data/database` will be checked every 10 seconds. Each time the watch emits the `changed` event, the `update-app` job will execute `/bin/update-app.sh`.


A File watch keeps an in-memory list of the md5 checksums associated with the file. The list is not persisted to disk and if ContainerPilot is restarted it will need to check back. If this list changes between polls, the watch emits one or two events:

- A `changed` event is emitted whenever there is a change (md5 checksum is different).
- A `healthy` event is emitted whenever ContainerPilot is able to read the file. This might mean that the state was previously unknown (as when ContainerPilot first starts up) or that it was previously unhealthy and is now healthy. This event will only be fired once for each change in md5 checksum of the file. Subsequent polls that return the same md5 checksum will not emit the event again.
- A `unhealthy` event is emitted whenever ContainerPilot is unable to read the file. This might mean that the file does not yet exist when we first poll. This event will only be fired once for each change of status. Subsequent polls that return the same md5 checksum will not emit the event again

```json5
jobs: [
  {
    name: "update-app",
    exec: "/bin/update-app.sh",
    when: {
      source: "watch./etc/ssl/cert.pem",
      each: "changed"
    }
  }
],
watches: [
  {
    name: "/etc/ssl/cert.pem",
    source: "file",
    tag: "password",
    interval: 60
  }
]
```

In this example, the watch `/etc/ssl/cert.pem` will be checked every 60 seconds. Each time the watch emits the `changed` event, the `update-app` job will execute `/bin/update-app.sh`.
