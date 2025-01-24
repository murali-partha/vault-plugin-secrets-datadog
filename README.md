# Vault Plugin Datadog (WIP)

This is just an attempt to understand Hashicorp Vault's plugin ecosystem and is not meant to be used for production scenarios.


## Usage

To write the datadog config run the below command.

```sh
$ vault write datadog/config \
  datadog_app_key="<datadog-application-key>" \
  datadog_api_key="<datadog-api-key>" \
  host="<datadog-host-url>"
...
Success! Data written to: datadog/config 
```

Once the config is successfully written, the secrets can be read using

```sh
$ vault read datadog/creds/<secret-name>
...

Key                Value
---                -----
lease_id           datadog/creds/secret1/QvpDdFRbjPzy3EbOQG82YK3q
lease_duration     5s
lease_renewable    false
key_id             <datadog-api-key-id>
secret             <datadog-api-key>
```


At the moment, the lease duration is set to just 5s.


[//]: <> (Provide usage instructions and/or links to this plugin)

## Developing

To compile a development version of this plugin, run `make` or `make dev`.
This will put the plugin binary in the `bin` and `$GOPATH/bin` folders. `dev`
mode will only generate the binary for your platform and is faster:

```sh
$ make dev
```

Put the plugin binary into a location of your choice. This directory
will be specified as the [`plugin_directory`](https://developer.hashicorp.com/vault/docs/configuration#plugin_directory)
in the Vault config used to start the server.

```hcl
# config.hcl
plugin_directory = "path/to/plugin/directory"
...
```

Start a Vault server with this config file:

```sh
$ vault server -dev -config=path/to/config.hcl ...
...
```

Once the server is started, register the plugin in the Vault server's [plugin catalog](https://developer.hashicorp.com/vault/docs/plugins/plugin-architecture#plugin-catalog):

```sh
$ SHA256=$(openssl dgst -sha256 $GOPATH/vault-plugin-secrets-myplugin | cut -d ' ' -f2)
$ vvault plugin register -sha256=$SHA256 secret vault-plugin-secrets-datadog
...
Success! Registered plugin: vault-plugin-secrets-datadog```

Enable the secrets engine to use this plugin:

```sh
$ vault secrets enable -path=datadog vault-plugin-secrets-datadog
...

Success! Enabled the vault-plugin-secrets-datadog secrets engine at: datadog/
```
