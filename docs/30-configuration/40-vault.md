# Vault

ContainerPilot uses Hashicorp's [Vault](https://www.vaultproject.io/) to watch secrets. Watches look to Vault to find out the secret has been changed.

## Client configuration

The `vault` field in the ContainerPilot config file configures ContainerPilot's Vault client. For use with Vault's ACL system, use the `VAULT_TOKEN` environment variable or you can specify file, which contains token `token: "file:///secrets/vault_token"`. If you are communicating with Vault over TLS you may include the scheme (ex. https://vault:8200). If you need extra configuration options for TLS, you can use the following optional fields (or environment variable options described in the [Vault documentation](https://www.vaultproject.io/docs/commands#environment-variables)) instead of a simple string:

```json5
vault: {
  address: "vault.example.com:8200",
  scheme: "https",
  token: "s.7BzHD0F5XZdlqgqBRDmf2oyX", // or file:///secrets/vault_token or VAULT_TOKEN
  tls: {
    cafile: "ca.crt",                 // or VAULT_CACERT
    capath: "ca_certs/",              // or VAULT_CAPATH
    clientcert: "client.crt",         // or VAULT_CLIENT_CERT
    clientkey: "client.key",          // or VAULT_CLIENT_KEY
    servername: "vault.example.com",  // or VAULT_TLS_SERVER_NAME
    verify: false,                    // or VAULT_SKIP_VERIFY
  }
}
```
